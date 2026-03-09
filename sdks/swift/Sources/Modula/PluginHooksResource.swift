import Foundation

public final class PluginHooksResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// List all plugin-registered hooks with approval status.
    public func list() async throws -> [PluginHook] {
        let envelope: HookListEnvelope = try await http.get(path: "/api/v1/admin/plugins/hooks")
        return envelope.hooks
    }

    /// Approve one or more plugin hooks.
    public func approve(hooks: [HookApprovalItem]) async throws {
        try await http.post(path: "/api/v1/admin/plugins/hooks/approve", body: HookApprovalBody(hooks: hooks)) as Void
    }

    /// Revoke approval for one or more plugin hooks.
    public func revoke(hooks: [HookApprovalItem]) async throws {
        try await http.post(path: "/api/v1/admin/plugins/hooks/revoke", body: HookApprovalBody(hooks: hooks)) as Void
    }
}

// MARK: - Private

private struct HookListEnvelope: Decodable {
    let hooks: [PluginHook]
}

private struct HookApprovalBody: Encodable {
    let hooks: [HookApprovalItem]
}
