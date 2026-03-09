import Foundation

public final class AdminTreeResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    public func get(slug: String, format: String = "") async throws -> Data {
        var queryItems: [URLQueryItem]?
        if !format.isEmpty {
            queryItems = [URLQueryItem(name: "format", value: format)]
        }
        return try await http.get(path: "/api/v1/admin/tree/" + slug, queryItems: queryItems) as Data
    }
}
