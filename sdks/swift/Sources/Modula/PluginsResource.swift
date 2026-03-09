import Foundation

public final class PluginsResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// List all installed plugins.
    public func list() async throws -> [PluginListItem] {
        let envelope: PluginListEnvelope = try await http.get(path: "/api/v1/admin/plugins")
        return envelope.plugins
    }

    /// Get detailed info for a specific plugin.
    public func get(name: String) async throws -> PluginInfo {
        let escaped = name.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? name
        return try await http.get(path: "/api/v1/admin/plugins/\(escaped)")
    }

    /// Reload a plugin from disk.
    public func reload(name: String) async throws -> PluginActionResponse {
        let escaped = name.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? name
        return try await http.post(path: "/api/v1/admin/plugins/\(escaped)/reload", body: EmptyBody())
    }

    /// Enable a disabled plugin.
    public func enable(name: String) async throws -> PluginStateResponse {
        let escaped = name.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? name
        return try await http.post(path: "/api/v1/admin/plugins/\(escaped)/enable", body: EmptyBody())
    }

    /// Disable an active plugin.
    public func disable(name: String) async throws -> PluginStateResponse {
        let escaped = name.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? name
        return try await http.post(path: "/api/v1/admin/plugins/\(escaped)/disable", body: EmptyBody())
    }

    /// Dry-run cleanup to list orphaned tables.
    public func cleanupDryRun() async throws -> CleanupDryRunResponse {
        try await http.get(path: "/api/v1/admin/plugins/cleanup")
    }

    /// Drop orphaned plugin tables.
    public func cleanupDrop(params: CleanupDropParams) async throws -> CleanupDropResponse {
        try await http.post(path: "/api/v1/admin/plugins/cleanup", body: params)
    }
}

// MARK: - Private

private struct PluginListEnvelope: Decodable {
    let plugins: [PluginListItem]
}

struct EmptyBody: Encodable {}
