/**
 * Admin media resource for managing admin media assets.
 *
 * Provides CRUD operations, upload, download, and batch move
 * for admin media at `POST /adminmedia`, `GET /adminmedia`, etc.
 *
 * @module resources/admin-media
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { ApiError, RequestOptions, AdminMediaID } from '../types/common.js'
import { isApiError } from '../types/common.js'
import type { AdminMedia } from '../types/admin-media.js'
import type { MoveAdminMediaParams, MoveAdminMediaResponse } from '../types/admin-media.js'

/**
 * Options for admin media upload operations.
 */
type AdminMediaUploadOptions = RequestOptions & {
  /**
   * Optional S3 key path prefix for organizing admin media files.
   *
   * Controls the directory structure in object storage. Segments are
   * separated by `/`. Leading and trailing slashes are stripped server-side.
   *
   * @example `'admin/icons'` -- stores as `admin/icons/filename.jpg`
   *
   * When omitted, the server defaults to date-based organization (`YYYY/M`).
   */
  path?: string
  /**
   * Optional admin media folder ID to place the uploaded file into.
   */
  folderId?: string
}

/**
 * Admin media resource operations available on `client.adminMedia`.
 */
type AdminMediaResource = {
  /**
   * Upload a file to the admin media library.
   *
   * Sends a `multipart/form-data` POST request. The `Content-Type` header
   * is set automatically by the browser/runtime (not `application/json`).
   *
   * @param file - The file to upload (browser `File` or `Blob`).
   * @param opts - Optional upload options (abort signal, path, folderId).
   * @returns The created admin media entity.
   * @throws {@link ApiError} on non-2xx responses.
   * @throws `TypeError` on network failure.
   */
  upload: (file: File | Blob, opts?: AdminMediaUploadOptions) => Promise<AdminMedia>

  /**
   * Batch move admin media items to a folder (or root).
   *
   * @param params - Move parameters with media IDs and target folder.
   * @param opts - Optional request options.
   * @returns The move result with count of moved items.
   */
  move: (params: MoveAdminMediaParams, opts?: RequestOptions) => Promise<MoveAdminMediaResponse>
}

/**
 * Create the admin media resource.
 *
 * @param http - Configured HTTP client (used for the `raw` method).
 * @param defaultTimeout - Default timeout in milliseconds.
 * @param credentials - Fetch credentials mode.
 * @param apiKey - Optional API key for the Authorization header.
 * @returns An {@link AdminMediaResource} with upload and move methods.
 * @internal
 */
function createAdminMediaResource(
  http: HttpClient,
  defaultTimeout: number,
  credentials: RequestCredentials,
  apiKey?: string,
): AdminMediaResource {
  return {
    async upload(file: File | Blob, opts?: AdminMediaUploadOptions): Promise<AdminMedia> {
      const form = new FormData()
      form.append('file', file)
      if (opts?.path !== undefined) {
        form.append('path', opts.path)
      }
      if (opts?.folderId !== undefined) {
        form.append('folder_id', opts.folderId)
      }

      const headers: Record<string, string> = {}
      if (apiKey) {
        headers['Authorization'] = `Bearer ${apiKey}`
      }

      let signal: AbortSignal
      if (opts?.signal) {
        const controller = new AbortController()
        const timeoutSignal = AbortSignal.timeout(defaultTimeout)

        const onAbort = (source: AbortSignal) => () => {
          if (!controller.signal.aborted) controller.abort(source.reason)
        }

        timeoutSignal.addEventListener('abort', onAbort(timeoutSignal), { once: true })
        opts.signal.addEventListener('abort', onAbort(opts.signal), { once: true })

        if (opts.signal.aborted) controller.abort(opts.signal.reason)

        signal = controller.signal
      } else {
        signal = AbortSignal.timeout(defaultTimeout)
      }

      const response = await http.raw('/adminmedia', {
        method: 'POST',
        headers,
        credentials,
        signal,
        body: form,
      })

      if (!response.ok) {
        const ct = response.headers.get('content-type')
        const isJson = ct !== null && ct.includes('application/json')
        let body: unknown
        let message = response.statusText
        if (isJson) {
          body = await response.json()
        } else {
          const text = await response.text()
          if (text) {
            message = text.trim()
            body = text.trim()
          }
        }
        const err: ApiError = {
          _tag: 'ApiError' as const,
          status: response.status,
          message,
          body,
        }
        throw err
      }

      return response.json() as Promise<AdminMedia>
    },

    move(params: MoveAdminMediaParams, opts?: RequestOptions): Promise<MoveAdminMediaResponse> {
      return http.post<MoveAdminMediaResponse>('/adminmedia/move', params as unknown as Record<string, unknown>, opts)
    },
  }
}

export type { AdminMediaResource, AdminMediaUploadOptions }
export { createAdminMediaResource }
