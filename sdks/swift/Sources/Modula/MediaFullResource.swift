import Foundation

/// Provides access to the composed media listing endpoint that returns media
/// items with author name information.
///
/// The response is returned as raw `Data` because the shape includes joined
/// author data that extends the basic Media struct.
///
/// ```swift
/// let data = try await client.mediaFull.list()
/// let items = try JSONSerialization.jsonObject(with: data)
/// ```
public final class MediaFullResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns all media items with their author names as raw JSON.
    ///
    /// - Returns: Raw JSON data containing media items with author information.
    public func list() async throws -> Data {
        let urlString = http.baseURL + "/api/v1/media/full"
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
