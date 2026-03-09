import Foundation

public final class ContentDeliveryResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    public func getPage(slug: String, format: String = "", locale: String = "") async throws -> Data {
        var queryItems: [URLQueryItem] = []
        if !format.isEmpty {
            queryItems.append(URLQueryItem(name: "format", value: format))
        }
        if !locale.isEmpty {
            queryItems.append(URLQueryItem(name: "locale", value: locale))
        }
        let trimmed = slug.hasPrefix("/") ? String(slug.dropFirst()) : slug
        let path = "/api/v1/content/" + trimmed
        return try await http.get(path: path, queryItems: queryItems.isEmpty ? nil : queryItems) as Data
    }
}
