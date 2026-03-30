import Foundation

/// Provides access to the admin metrics endpoint, which returns a snapshot of
/// server runtime metrics (request counts, latencies, error rates, etc.).
///
/// Requires `config:read` permission.
///
/// The response shape depends on which metrics the server is collecting and may
/// change between versions. The raw `Data` is returned so the caller can parse
/// it into their own struct as needed.
///
/// ```swift
/// let data = try await client.metrics.get()
/// let snapshot = try JSONSerialization.jsonObject(with: data)
/// ```
public final class MetricsResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns the current server metrics snapshot as raw JSON data.
    public func get() async throws -> Data {
        let request = try buildGetRequest(path: "/api/v1/admin/metrics")
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
