import Foundation

/// Provides publishing, versioning, and restore operations for content.
public final class PublishingResource: Sendable {
    let http: HTTPClient
    let prefix: String

    init(http: HTTPClient, prefix: String) {
        self.http = http
        self.prefix = prefix
    }

    // MARK: - Publish / Unpublish

    /// Publish content, creating a new version snapshot.
    public func publish(req: PublishRequest) async throws -> PublishResponse {
        try await http.post(path: "/api/v1/\(prefix)/publish", body: req)
    }

    /// Remove the published state from content.
    public func unpublish(req: PublishRequest) async throws -> PublishResponse {
        try await http.post(path: "/api/v1/\(prefix)/unpublish", body: req)
    }

    // MARK: - Schedule

    /// Set a future publication time for content.
    public func schedule(req: ScheduleRequest) async throws -> ScheduleResponse {
        try await http.post(path: "/api/v1/\(prefix)/schedule", body: req)
    }

    // MARK: - Versions

    /// List all version snapshots for a given content data ID.
    public func listVersions(contentDataID: String) async throws -> [ContentVersion] {
        try await http.get(
            path: "/api/v1/\(prefix)/versions",
            queryItems: [URLQueryItem(name: "q", value: contentDataID)]
        )
    }

    /// Get a single content version by its ID.
    public func getVersion(versionID: String) async throws -> ContentVersion {
        try await http.get(
            path: "/api/v1/\(prefix)/versions/",
            queryItems: [URLQueryItem(name: "q", value: versionID)]
        )
    }

    /// Manually create a new version snapshot.
    public func createVersion(req: CreateVersionRequest) async throws -> ContentVersion {
        try await http.post(path: "/api/v1/\(prefix)/versions", body: req)
    }

    /// Remove a content version by its ID.
    public func deleteVersion(versionID: String) async throws {
        try await http.delete(
            path: "/api/v1/\(prefix)/versions/",
            queryItems: [URLQueryItem(name: "q", value: versionID)]
        )
    }

    // MARK: - Restore

    /// Restore content to a previous version.
    public func restore(req: RestoreRequest) async throws -> RestoreResponse {
        try await http.post(path: "/api/v1/\(prefix)/restore", body: req)
    }
}

/// Provides publishing, versioning, and restore operations for admin content.
public final class AdminPublishingResource: Sendable {
    let http: HTTPClient
    let prefix: String

    init(http: HTTPClient, prefix: String) {
        self.http = http
        self.prefix = prefix
    }

    // MARK: - Publish / Unpublish

    /// Publish admin content, creating a new version snapshot.
    public func publish(req: AdminPublishRequest) async throws -> AdminPublishResponse {
        try await http.post(path: "/api/v1/\(prefix)/publish", body: req)
    }

    /// Remove the published state from admin content.
    public func unpublish(req: AdminPublishRequest) async throws -> AdminPublishResponse {
        try await http.post(path: "/api/v1/\(prefix)/unpublish", body: req)
    }

    // MARK: - Schedule

    /// Set a future publication time for admin content.
    public func schedule(req: AdminScheduleRequest) async throws -> AdminScheduleResponse {
        try await http.post(path: "/api/v1/\(prefix)/schedule", body: req)
    }

    // MARK: - Versions

    /// List all version snapshots for a given admin content data ID.
    public func listVersions(adminContentDataID: String) async throws -> [AdminContentVersion] {
        try await http.get(
            path: "/api/v1/\(prefix)/versions",
            queryItems: [URLQueryItem(name: "q", value: adminContentDataID)]
        )
    }

    /// Get a single admin content version by its ID.
    public func getVersion(versionID: String) async throws -> AdminContentVersion {
        try await http.get(
            path: "/api/v1/\(prefix)/versions/",
            queryItems: [URLQueryItem(name: "q", value: versionID)]
        )
    }

    /// Manually create a new admin content version snapshot.
    public func createVersion(req: CreateAdminVersionRequest) async throws -> AdminContentVersion {
        try await http.post(path: "/api/v1/\(prefix)/versions", body: req)
    }

    /// Remove an admin content version by its ID.
    public func deleteVersion(versionID: String) async throws {
        try await http.delete(
            path: "/api/v1/\(prefix)/versions/",
            queryItems: [URLQueryItem(name: "q", value: versionID)]
        )
    }

    // MARK: - Restore

    /// Restore admin content to a previous version.
    public func restore(req: AdminRestoreRequest) async throws -> AdminRestoreResponse {
        try await http.post(path: "/api/v1/\(prefix)/restore", body: req)
    }
}
