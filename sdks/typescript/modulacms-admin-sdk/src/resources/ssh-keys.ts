/**
 * SSH keys resource providing list, create, and remove operations.
 * Uses a path-parameter DELETE pattern (`/ssh-keys/{id}`) instead
 * of the standard query-parameter pattern.
 *
 * @remarks List returns {@link import('../types/users.js').SshKeyListItem} (without `public_key`)
 * while create returns a full {@link import('../types/users.js').SshKey} (with `public_key`).
 *
 * @module resources/ssh-keys
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'
import type { SshKey, SshKeyListItem, CreateSshKeyRequest } from '../types/users.js'

/**
 * SSH key management operations available on `client.sshKeys`.
 */
type SshKeysResource = {
  /**
   * List all SSH keys for the authenticated user.
   * Returns summary items without the public key material.
   * @param opts - Optional request options.
   * @returns Array of SSH key summaries.
   */
  list: (opts?: RequestOptions) => Promise<SshKeyListItem[]>

  /**
   * Register a new SSH public key.
   * @param params - The public key material and label.
   * @param opts - Optional request options.
   * @returns The full SSH key record including `public_key`.
   */
  create: (params: CreateSshKeyRequest, opts?: RequestOptions) => Promise<SshKey>

  /**
   * Remove an SSH key by its ID. Uses path-parameter deletion (`DELETE /ssh-keys/{id}`).
   * @param id - The SSH key ID to remove.
   * @param opts - Optional request options.
   */
  remove: (id: string, opts?: RequestOptions) => Promise<void>
}

/**
 * Create the SSH keys resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link SshKeysResource} with list, create, and remove methods.
 * @internal
 */
function createSshKeysResource(http: HttpClient): SshKeysResource {
  return {
    list(opts?: RequestOptions): Promise<SshKeyListItem[]> {
      return http.get<SshKeyListItem[]>('/ssh-keys', undefined, opts)
    },

    create(params: CreateSshKeyRequest, opts?: RequestOptions): Promise<SshKey> {
      return http.post<SshKey>('/ssh-keys', params as Record<string, unknown>, opts)
    },

    remove(id: string, opts?: RequestOptions): Promise<void> {
      return http.del(`/ssh-keys/${id}`, undefined, opts)
    },
  }
}

export type { SshKeysResource }
export { createSshKeysResource }
