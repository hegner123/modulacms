import Foundation

/// Provides access to session information for a specific user.
///
/// Requires `sessions:read` permission.
///
/// The response shape depends on the session data stored for the user, so
/// raw `Data` is returned.
///
/// ```swift
/// let data = try await client.userSessions.getByUser(userID: userID)
/// let sessions = try JSONSerialization.jsonObject(with: data)
/// ```
public final class UserSessionsResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns session information for the given user as raw JSON.
    ///
    /// - Parameter userID: The user whose sessions to retrieve.
    /// - Returns: Raw JSON data containing the user's sessions.
    public func getByUser(userID: UserID) async throws -> Data {
        let queryItems = [URLQueryItem(name: "q", value: userID.rawValue)]
        var urlString = http.baseURL + "/api/v1/users/sessions"
        var components = URLComponents()
        components.queryItems = queryItems
        if let query = components.percentEncodedQuery {
            urlString += "?" + query
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
