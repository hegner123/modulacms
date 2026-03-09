import Foundation

public final class SessionsResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    public func update(params: UpdateSessionParams) async throws -> Session {
        try await http.put(path: "/api/v1/sessions/", body: params)
    }

    public func remove(id: SessionID) async throws {
        let queryItems = [URLQueryItem(name: "q", value: id.rawValue)]
        try await http.delete(path: "/api/v1/sessions/", queryItems: queryItems)
    }
}
