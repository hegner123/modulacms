import Foundation

/// Provides access to composed content data endpoints that return content items
/// with joined author, datatype, and field data.
///
/// These endpoints return richer responses than the basic CRUD endpoints.
/// The response is returned as raw `Data` because the shape includes nested
/// associations that vary by datatype configuration.
///
/// ```swift
/// let data = try await client.contentDataFull.getFull(id: contentID)
/// let full = try JSONSerialization.jsonObject(with: data)
/// ```
public final class ContentDataFullResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns a single content data item with its author, datatype, and fields.
    ///
    /// - Parameter id: The content data ID to look up.
    /// - Returns: Raw JSON data containing the full content item with associations.
    public func getFull(id: ContentID) async throws -> Data {
        let queryItems = [URLQueryItem(name: "q", value: id.rawValue)]
        return try await executeGetRaw(path: "/api/v1/contentdata/full", queryItems: queryItems)
    }

    /// Returns all content data items belonging to the given route.
    ///
    /// - Parameter routeID: The route to filter by.
    /// - Returns: Raw JSON data containing the content items for the route.
    public func listByRoute(routeID: RouteID) async throws -> Data {
        let queryItems = [URLQueryItem(name: "q", value: routeID.rawValue)]
        return try await executeGetRaw(path: "/api/v1/contentdata/by-route", queryItems: queryItems)
    }

    /// Returns a single admin content data item with its author, datatype, and fields.
    ///
    /// - Parameter id: The admin content data ID to look up.
    /// - Returns: Raw JSON data containing the full admin content item with associations.
    public func adminGetFull(id: AdminContentID) async throws -> Data {
        let queryItems = [URLQueryItem(name: "q", value: id.rawValue)]
        return try await executeGetRaw(path: "/api/v1/admincontentdatas/full", queryItems: queryItems)
    }

    // MARK: - Private

    private func executeGetRaw(path: String, queryItems: [URLQueryItem]?) async throws -> Data {
        var urlString = http.baseURL + path
        if let queryItems, !queryItems.isEmpty {
            var components = URLComponents()
            components.queryItems = queryItems
            if let query = components.percentEncodedQuery {
                urlString += "?" + query
            }
        }
        guard let url = URL(string: urlString) else {
            throw APIError(statusCode: 0, message: "Invalid URL: \(urlString)")
        }
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        let (data, response) = try await http.executeRaw(request)
        let statusCode = response.statusCode
        if statusCode < 200 || statusCode >= 300 {
            let bodyStr = String(data: data, encoding: .utf8) ?? ""
            throw APIError(statusCode: statusCode, body: bodyStr)
        }
        return data
    }
}
