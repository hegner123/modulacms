import Foundation

/// Provides access to composed datatype endpoints that return datatypes with
/// their fields and other associated data.
///
/// The response is returned as raw `Data` because the shape includes nested
/// associations that vary by configuration.
///
/// ```swift
/// let data = try await client.datatypeFull.getFull(id: datatypeID)
/// let full = try JSONSerialization.jsonObject(with: data)
/// ```
public final class DatatypeFullResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns a single datatype with its fields and metadata.
    ///
    /// - Parameter id: The datatype ID to look up.
    /// - Returns: Raw JSON data containing the full datatype with associations.
    public func getFull(id: DatatypeID) async throws -> Data {
        let queryItems = [URLQueryItem(name: "q", value: id.rawValue)]
        return try await executeGetRaw(path: "/api/v1/datatype/full", queryItems: queryItems)
    }

    /// Returns all datatypes with their fields and metadata.
    ///
    /// - Returns: Raw JSON data containing all datatypes with associations.
    public func listFull() async throws -> Data {
        try await executeGetRaw(path: "/api/v1/datatype/full/list", queryItems: nil)
    }

    /// Returns a single admin datatype with its fields and metadata.
    ///
    /// - Parameter id: The admin datatype ID to look up.
    /// - Returns: Raw JSON data containing the full admin datatype with associations.
    public func adminGetFull(id: AdminDatatypeID) async throws -> Data {
        let queryItems = [URLQueryItem(name: "q", value: id.rawValue)]
        return try await executeGetRaw(path: "/api/v1/admindatatypes/full", queryItems: queryItems)
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
