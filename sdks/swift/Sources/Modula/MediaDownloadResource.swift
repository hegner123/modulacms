import Foundation

/// Provides media file download operations.
///
/// The server generates a pre-signed S3 URL and returns a 302 redirect.
/// These methods extract the redirect URL without following it, so the caller
/// can use the URL directly (e.g. for streaming downloads or passing to a frontend).
///
/// ```swift
/// let url = try await client.mediaDownload.getURL(id: mediaID)
/// // Use url to download the file or pass to a web view
/// ```
public final class MediaDownloadResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns a pre-signed download URL for the given media item.
    ///
    /// The URL is valid for approximately 15 minutes. Returns the URL string
    /// from the server's 302 redirect Location header.
    ///
    /// - Parameter id: The media item to download.
    /// - Returns: A pre-signed S3 download URL.
    public func getURL(id: MediaID) async throws -> String {
        try await getDownloadURL(path: "/api/v1/media/\(id.rawValue)/download")
    }

    /// Returns a pre-signed download URL for the given admin media item.
    ///
    /// The URL is valid for approximately 15 minutes. Returns the URL string
    /// from the server's 302 redirect Location header.
    ///
    /// - Parameter id: The admin media item to download.
    /// - Returns: A pre-signed S3 download URL.
    public func adminGetURL(id: AdminMediaID) async throws -> String {
        try await getDownloadURL(path: "/api/v1/adminmedia/\(id.rawValue)/download")
    }

    // MARK: - Private

    private func getDownloadURL(path: String) async throws -> String {
        let urlString = http.baseURL + path
        guard let url = URL(string: urlString) else {
            throw APIError(statusCode: 0, message: "Invalid URL: \(urlString)")
        }

        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        if !http.apiKey.isEmpty {
            request.setValue("Bearer \(http.apiKey)", forHTTPHeaderField: "Authorization")
        }

        // Use a session that does not follow redirects so we can extract the Location header.
        let noRedirectDelegate = NoRedirectDelegate()
        let noRedirectConfig = URLSessionConfiguration.ephemeral
        noRedirectConfig.timeoutIntervalForRequest = 30
        noRedirectConfig.httpCookieAcceptPolicy = .never
        noRedirectConfig.httpShouldSetCookies = false
        noRedirectConfig.httpCookieStorage = nil
        let noRedirectSession = URLSession(
            configuration: noRedirectConfig,
            delegate: noRedirectDelegate,
            delegateQueue: nil
        )

        let (data, response) = try await noRedirectSession.data(for: request)
        guard let httpResp = response as? HTTPURLResponse else {
            throw APIError(statusCode: 0, message: "Invalid response type")
        }

        // The server returns 302 with a Location header.
        if httpResp.statusCode == 302 || httpResp.statusCode == 307 {
            if let location = httpResp.value(forHTTPHeaderField: "Location"), !location.isEmpty {
                return location
            }
            throw APIError(statusCode: httpResp.statusCode, message: "Redirect response missing Location header")
        }

        if httpResp.statusCode < 200 || httpResp.statusCode >= 300 {
            let bodyStr = String(data: data, encoding: .utf8) ?? ""
            throw APIError(statusCode: httpResp.statusCode, body: bodyStr)
        }

        // Unexpected 2xx -- check for Location header anyway.
        if let location = httpResp.value(forHTTPHeaderField: "Location"), !location.isEmpty {
            return location
        }

        throw APIError(statusCode: httpResp.statusCode, message: "Unexpected status \(httpResp.statusCode) from download endpoint")
    }
}

/// URLSession delegate that prevents automatic redirect following.
private final class NoRedirectDelegate: NSObject, URLSessionTaskDelegate, Sendable {
    func urlSession(
        _ session: URLSession,
        task: URLSessionTask,
        willPerformHTTPRedirection response: HTTPURLResponse,
        newRequest request: URLRequest,
        completionHandler: @escaping (URLRequest?) -> Void
    ) {
        // Return nil to prevent following the redirect.
        completionHandler(nil)
    }
}
