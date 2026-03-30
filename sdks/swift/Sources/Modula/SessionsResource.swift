import Foundation

public final class SessionsResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns all active sessions. For admin users this includes all sessions
    /// across all users; for non-admin users only their own sessions are returned.
    public func list() async throws -> [Session] {
        try await http.get(path: "/api/v1/sessions")
    }

    public func update(params: UpdateSessionParams) async throws -> Session {
        try await http.put(path: "/api/v1/sessions/", body: params)
    }

    public func remove(id: SessionID) async throws {
        let queryItems = [URLQueryItem(name: "q", value: id.rawValue)]
        try await http.delete(path: "/api/v1/sessions/", queryItems: queryItems)
    }
}
