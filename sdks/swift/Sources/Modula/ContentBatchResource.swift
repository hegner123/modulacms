import Foundation

public final class ContentBatchResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    public func update<T: Encodable & Sendable>(request: T) async throws -> Data {
        try await http.post(path: "/api/v1/content/batch", body: request) as Data
    }
}
