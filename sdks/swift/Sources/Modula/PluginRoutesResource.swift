import Foundation

public final class PluginRoutesResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// List all plugin-registered routes with approval status.
    public func list() async throws -> [PluginRoute] {
        let envelope: RouteListEnvelope = try await http.get(path: "/api/v1/admin/plugins/routes")
        return envelope.routes
    }

    /// Approve one or more plugin routes.
    public func approve(routes: [RouteApprovalItem]) async throws {
        try await http.post(path: "/api/v1/admin/plugins/routes/approve", body: RouteApprovalBody(routes: routes)) as Void
    }

    /// Revoke approval for one or more plugin routes.
    public func revoke(routes: [RouteApprovalItem]) async throws {
        try await http.post(path: "/api/v1/admin/plugins/routes/revoke", body: RouteApprovalBody(routes: routes)) as Void
    }
}

// MARK: - Private

private struct RouteListEnvelope: Decodable {
    let routes: [PluginRoute]
}

private struct RouteApprovalBody: Encodable {
    let routes: [RouteApprovalItem]
}
