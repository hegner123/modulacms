import Foundation

/// Provides composite media operations for reference scanning and deletion with cleanup.
///
/// - ``getReferences(id:)`` scans content fields for references to a media item.
/// - ``deleteWithCleanup(id:)`` deletes a media item and clears all content field references.
public final class MediaCompositeResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Check media storage health by scanning for orphaned files.
    ///
    /// Orphaned files exist in the storage bucket but have no corresponding
    /// database record. Requires `media:admin` permission.
    ///
    /// - Returns: A ``MediaHealthResponse`` with orphan counts and keys.
    public func health() async throws -> MediaHealthResponse {
        try await http.get(path: "/api/v1/media/health")
    }

    /// Remove orphaned files from media storage.
    ///
    /// Deletes files that exist in the storage bucket but have no corresponding
    /// database record. Requires `media:admin` permission.
    ///
    /// - Returns: A ``MediaCleanupResponse`` with deleted and failed counts.
    public func cleanup() async throws -> MediaCleanupResponse {
        try await http.delete(path: "/api/v1/media/cleanup")
    }

    /// Scan content fields for references to a media item.
    ///
    /// - Parameter id: The media ID to scan for.
    /// - Returns: A ``MediaReferenceScanResponse`` listing all content fields that reference this media.
    public func getReferences(id: MediaID) async throws -> MediaReferenceScanResponse {
        let queryItems = [URLQueryItem(name: "q", value: id.rawValue)]
        return try await http.get(path: "/api/v1/media/references", queryItems: queryItems)
    }

    /// Delete a media item and clean up all content field references to it.
    ///
    /// - Parameter id: The media ID to delete.
    public func deleteWithCleanup(id: MediaID) async throws {
        let queryItems = [
            URLQueryItem(name: "q", value: id.rawValue),
            URLQueryItem(name: "clean_refs", value: "true"),
        ]
        try await http.delete(path: "/api/v1/media/", queryItems: queryItems) as Void
    }
}
