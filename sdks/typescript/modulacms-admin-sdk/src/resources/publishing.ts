/**
 * Publishing resource providing publish, unpublish, schedule, version
 * management, and restore operations for content and admin content.
 *
 * @module resources/publishing
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions, ContentID, AdminContentID, ContentVersionID, AdminContentVersionID } from '../types/common.js'
import type { ContentVersion, AdminContentVersion } from '../types/content.js'

// ---------------------------------------------------------------------------
// Request/Response types
// ---------------------------------------------------------------------------

/** Request body for publishing content. */
export type PublishRequest = {
  content_data_id: ContentID
  /** Optional locale code to publish content for a specific locale. */
  locale?: string
}

/** Request body for publishing admin content. */
export type AdminPublishRequest = {
  admin_content_data_id: AdminContentID
  /** Optional locale code to publish admin content for a specific locale. */
  locale?: string
}

/** Response from a publish or unpublish operation. */
export type PublishResponse = {
  status: string
  version_number?: number
  content_version_id?: string
  content_data_id: string
}

/** Response from an admin publish or unpublish operation. */
export type AdminPublishResponse = {
  status: string
  version_number?: number
  admin_content_version_id?: string
  admin_content_data_id: string
}

/** Request body for scheduling content publication. */
export type ScheduleRequest = {
  content_data_id: ContentID
  publish_at: string
}

/** Request body for scheduling admin content publication. */
export type AdminScheduleRequest = {
  admin_content_data_id: AdminContentID
  publish_at: string
}

/** Response from a schedule operation. */
export type ScheduleResponse = {
  status: string
  content_data_id: string
  publish_at: string
}

/** Response from an admin schedule operation. */
export type AdminScheduleResponse = {
  status: string
  admin_content_data_id: string
  publish_at: string
}

/** Request body for manually creating a content version. */
export type CreateVersionRequest = {
  content_data_id: ContentID
  label?: string
}

/** Request body for manually creating an admin content version. */
export type CreateAdminVersionRequest = {
  admin_content_data_id: AdminContentID
  label?: string
}

/** Request body for restoring content to a previous version. */
export type RestoreRequest = {
  content_data_id: ContentID
  content_version_id: ContentVersionID
}

/** Request body for restoring admin content to a previous version. */
export type AdminRestoreRequest = {
  admin_content_data_id: AdminContentID
  admin_content_version_id: AdminContentVersionID
}

/** Response from a restore operation. */
export type RestoreResponse = {
  status: string
  content_data_id: string
  restored_version_id: string
  fields_restored: number
  unmapped_fields?: string[]
}

/** Response from an admin restore operation. */
export type AdminRestoreResponse = {
  status: string
  admin_content_data_id: string
  restored_version_id: string
  fields_restored: number
  unmapped_fields?: string[]
}

// ---------------------------------------------------------------------------
// Publishing resource type
// ---------------------------------------------------------------------------

/** Publishing operations available on `client.publishing` and `client.adminPublishing`. */
export type PublishingResource = {
  /** Publish content, creating a new version snapshot. */
  publish: (req: PublishRequest, opts?: RequestOptions) => Promise<PublishResponse>
  /** Remove the published state from content. */
  unpublish: (req: PublishRequest, opts?: RequestOptions) => Promise<PublishResponse>
  /** Set a future publication time for content. */
  schedule: (req: ScheduleRequest, opts?: RequestOptions) => Promise<ScheduleResponse>
  /** List all version snapshots for a content data node. */
  listVersions: (contentDataID: string, opts?: RequestOptions) => Promise<ContentVersion[]>
  /** Get a single content version by its ID. */
  getVersion: (versionID: string, opts?: RequestOptions) => Promise<ContentVersion>
  /** Manually create a new version snapshot. */
  createVersion: (req: CreateVersionRequest, opts?: RequestOptions) => Promise<ContentVersion>
  /** Remove a content version by its ID. */
  deleteVersion: (versionID: string, opts?: RequestOptions) => Promise<void>
  /** Restore content to a previous version. */
  restore: (req: RestoreRequest, opts?: RequestOptions) => Promise<RestoreResponse>
}

/** Admin publishing operations available on `client.adminPublishing`. */
export type AdminPublishingResource = {
  /** Publish admin content, creating a new version snapshot. */
  publish: (req: AdminPublishRequest, opts?: RequestOptions) => Promise<AdminPublishResponse>
  /** Remove the published state from admin content. */
  unpublish: (req: AdminPublishRequest, opts?: RequestOptions) => Promise<AdminPublishResponse>
  /** Set a future publication time for admin content. */
  schedule: (req: AdminScheduleRequest, opts?: RequestOptions) => Promise<AdminScheduleResponse>
  /** List all version snapshots for an admin content data node. */
  listVersions: (adminContentDataID: string, opts?: RequestOptions) => Promise<AdminContentVersion[]>
  /** Get a single admin content version by its ID. */
  getVersion: (versionID: string, opts?: RequestOptions) => Promise<AdminContentVersion>
  /** Manually create a new admin content version snapshot. */
  createVersion: (req: CreateAdminVersionRequest, opts?: RequestOptions) => Promise<AdminContentVersion>
  /** Remove an admin content version by its ID. */
  deleteVersion: (versionID: string, opts?: RequestOptions) => Promise<void>
  /** Restore admin content to a previous version. */
  restore: (req: AdminRestoreRequest, opts?: RequestOptions) => Promise<AdminRestoreResponse>
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Create the publishing resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @param prefix - API path prefix (`"content"` or `"admin/content"`).
 * @returns A {@link PublishingResource} with all publishing operations.
 * @internal
 */
function createPublishingResource(http: HttpClient, prefix: string): PublishingResource {
  return {
    publish(req: PublishRequest, opts?: RequestOptions): Promise<PublishResponse> {
      return http.post<PublishResponse>(`/${prefix}/publish`, req as unknown as Record<string, unknown>, opts)
    },

    unpublish(req: PublishRequest, opts?: RequestOptions): Promise<PublishResponse> {
      return http.post<PublishResponse>(`/${prefix}/unpublish`, req as unknown as Record<string, unknown>, opts)
    },

    schedule(req: ScheduleRequest, opts?: RequestOptions): Promise<ScheduleResponse> {
      return http.post<ScheduleResponse>(`/${prefix}/schedule`, req as unknown as Record<string, unknown>, opts)
    },

    listVersions(contentDataID: string, opts?: RequestOptions): Promise<ContentVersion[]> {
      return http.get<ContentVersion[]>(`/${prefix}/versions`, { q: contentDataID }, opts)
    },

    getVersion(versionID: string, opts?: RequestOptions): Promise<ContentVersion> {
      return http.get<ContentVersion>(`/${prefix}/versions/`, { q: versionID }, opts)
    },

    createVersion(req: CreateVersionRequest, opts?: RequestOptions): Promise<ContentVersion> {
      return http.post<ContentVersion>(`/${prefix}/versions`, req as unknown as Record<string, unknown>, opts)
    },

    deleteVersion(versionID: string, opts?: RequestOptions): Promise<void> {
      return http.del(`/${prefix}/versions/`, { q: versionID }, opts)
    },

    restore(req: RestoreRequest, opts?: RequestOptions): Promise<RestoreResponse> {
      return http.post<RestoreResponse>(`/${prefix}/restore`, req as unknown as Record<string, unknown>, opts)
    },
  }
}

/**
 * Create the admin publishing resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @param prefix - API path prefix (e.g. `"admin/content"`).
 * @returns An {@link AdminPublishingResource} with all admin publishing operations.
 * @internal
 */
function createAdminPublishingResource(http: HttpClient, prefix: string): AdminPublishingResource {
  return {
    publish(req: AdminPublishRequest, opts?: RequestOptions): Promise<AdminPublishResponse> {
      return http.post<AdminPublishResponse>(`/${prefix}/publish`, req as unknown as Record<string, unknown>, opts)
    },

    unpublish(req: AdminPublishRequest, opts?: RequestOptions): Promise<AdminPublishResponse> {
      return http.post<AdminPublishResponse>(`/${prefix}/unpublish`, req as unknown as Record<string, unknown>, opts)
    },

    schedule(req: AdminScheduleRequest, opts?: RequestOptions): Promise<AdminScheduleResponse> {
      return http.post<AdminScheduleResponse>(`/${prefix}/schedule`, req as unknown as Record<string, unknown>, opts)
    },

    listVersions(adminContentDataID: string, opts?: RequestOptions): Promise<AdminContentVersion[]> {
      return http.get<AdminContentVersion[]>(`/${prefix}/versions`, { q: adminContentDataID }, opts)
    },

    getVersion(versionID: string, opts?: RequestOptions): Promise<AdminContentVersion> {
      return http.get<AdminContentVersion>(`/${prefix}/versions/`, { q: versionID }, opts)
    },

    createVersion(req: CreateAdminVersionRequest, opts?: RequestOptions): Promise<AdminContentVersion> {
      return http.post<AdminContentVersion>(`/${prefix}/versions`, req as unknown as Record<string, unknown>, opts)
    },

    deleteVersion(versionID: string, opts?: RequestOptions): Promise<void> {
      return http.del(`/${prefix}/versions/`, { q: versionID }, opts)
    },

    restore(req: AdminRestoreRequest, opts?: RequestOptions): Promise<AdminRestoreResponse> {
      return http.post<AdminRestoreResponse>(`/${prefix}/restore`, req as unknown as Record<string, unknown>, opts)
    },
  }
}

export { createPublishingResource, createAdminPublishingResource }
