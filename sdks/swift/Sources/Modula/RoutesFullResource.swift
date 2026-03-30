import Foundation

/// Provides access to the composed route endpoint that returns a route with
/// its content tree.
///
/// The response shape depends on the content structure and datatype configuration
/// for the route, so raw `Data` is returned.
///
/// ```swift
/// let data = try await client.routesFull.getFull(id: routeID)
/// let full = try JSONSerialization.jsonObject(with: data)
/// ```
public final class RoutesFullResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns a route with its associated content tree as raw JSON.
    ///
    /// - Parameter id: The route ID to look up.
    /// - Returns: Raw JSON data containing the route and its content tree.
    public func getFull(id: RouteID) async throws -> Data {
        let queryItems = [URLQueryItem(name: "q", value: id.rawValue)]
        return try await executeGetRaw(path: "/api/v1/routes/full", queryItems: queryItems)
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
