import Foundation

/// Provides access to global content trees via `GET /api/v1/globals`.
///
/// Global content items are global-typed root nodes whose published trees are
/// available site-wide (e.g. navigation, footer, site settings).
///
/// The response is returned as raw `Data` because each entry contains a
/// recursively nested content tree whose shape depends on the datatype schema
/// configuration.
///
/// ```swift
/// let data = try await client.globals.list()
/// let globals = try JSONSerialization.jsonObject(with: data)
/// ```
public final class GlobalsResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns all published global content trees as raw JSON data.
    ///
    /// The response is an array of global content entries, each containing
    /// a content_data_id, datatype metadata, and a fully composed content tree.
    public func list() async throws -> Data {
        let request = try buildGetRequest(path: "/api/v1/globals")
        let (data, response) = try await http.executeRaw(request)
        let statusCode = response.statusCode
        if statusCode < 200 || statusCode >= 300 {
            let bodyStr = String(data: data, encoding: .utf8) ?? ""
            throw APIError(statusCode: statusCode, body: bodyStr)
        }
        return data
    }

    private func buildGetRequest(path: String) throws -> URLRequest {
        let urlString = http.baseURL + path
        guard let url = URL(string: urlString) else {
            throw APIError(statusCode: 0, message: "Invalid URL: \(urlString)")
        }
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        return request
    }
}
