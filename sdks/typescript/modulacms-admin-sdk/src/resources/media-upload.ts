/**
 * Media upload resource for uploading files via multipart/form-data
 * to `POST /media`.
 *
 * @remarks The initial upload response may return `srcset: null` because
 * responsive image variants are generated asynchronously. Poll
 * `client.media.get()` after a delay to retrieve the processed `srcset`.
 *
 * @module resources/media-upload
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { ApiError, RequestOptions } from '../types/common.js'
import { isApiError } from '../types/common.js'
import type { Media } from '../types/media.js'

/**
 * Media upload operations available on `client.mediaUpload`.
 */
type MediaUploadResource = {
  /**
   * Upload a file to the CMS media library.
   *
   * Sends a `multipart/form-data` POST request. The `Content-Type` header
   * is set automatically by the browser/runtime (not `application/json`).
   *
   * @param file - The file to upload (browser `File` or `Blob`).
   * @param opts - Optional request options (abort signal).
   * @returns The created media entity. Note: `srcset` may be `null` initially.
   * @throws {@link ApiError} on non-2xx responses.
   * @throws `TypeError` on network failure.
   */
  upload: (file: File | Blob, opts?: RequestOptions) => Promise<Media>
}

/**
 * Create the media upload resource.
 *
 * @param http - Configured HTTP client (used for the `raw` method).
 * @param defaultTimeout - Default timeout in milliseconds.
 * @param credentials - Fetch credentials mode.
 * @param apiKey - Optional API key for the Authorization header.
 * @returns A {@link MediaUploadResource} with an `upload` method.
 * @internal
 */
function createMediaUploadResource(
  http: HttpClient,
  defaultTimeout: number,
  credentials: RequestCredentials,
  apiKey?: string,
): MediaUploadResource {
  return {
    async upload(file: File | Blob, opts?: RequestOptions): Promise<Media> {
      const form = new FormData()
      form.append('file', file)

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

      const response = await http.raw('/media', {
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

      return response.json() as Promise<Media>
    },
  }
}

/**
 * Check if an error is a duplicate media conflict (HTTP 409).
 * Thrown when uploading a file that already exists in the media library.
 */
function isDuplicateMedia(err: unknown): err is ApiError {
  return isApiError(err) && err.status === 409
}

/**
 * Check if an error is an invalid media type rejection (HTTP 400).
 * Thrown when the uploaded file's MIME type is not allowed.
 */
function isInvalidMediaType(err: unknown): err is ApiError {
  return isApiError(err) && err.status === 400 && typeof err.message === 'string' && err.message.toLowerCase().includes('content type')
}

/**
 * Check if an error is a file-too-large rejection (HTTP 400).
 * Thrown when the uploaded file exceeds the server's size limit.
 */
function isFileTooLarge(err: unknown): err is ApiError {
  return isApiError(err) && err.status === 400 && typeof err.message === 'string' && err.message.toLowerCase().includes('too large')
}

export type { MediaUploadResource }
export { createMediaUploadResource, isDuplicateMedia, isInvalidMediaType, isFileTooLarge }
