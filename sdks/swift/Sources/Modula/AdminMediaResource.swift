import Foundation

/// Provides admin media upload and CRUD operations.
///
/// Admin media uses multipart upload (not JSON POST), so creation goes through
/// ``upload(data:filename:options:)`` rather than the generic Resource.
/// Standard CRUD (list, get, update, delete) is available via ``adminMedia``
/// on the client.
public final class AdminMediaResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Options for admin media upload operations.
    public struct UploadOptions: Sendable {
        /// S3 key path prefix for organizing media files.
        /// Segments are separated by "/". Leading and trailing slashes are stripped server-side.
        /// Examples: "admin/icons", "admin/backgrounds".
        /// When nil, the server defaults to date-based organization (YYYY/M).
        public let path: String?

        public init(path: String? = nil) {
            self.path = path
        }
    }

    /// Upload a file to the admin media library.
    ///
    /// - Parameters:
    ///   - fileData: The raw file bytes to upload.
    ///   - filename: The original filename (used for display name and MIME detection).
    ///   - options: Optional upload configuration (path prefix).
    /// - Returns: The newly created ``AdminMedia`` record.
    public func upload(data fileData: Data, filename: String, options: UploadOptions? = nil) async throws -> AdminMedia {
        let boundary = UUID().uuidString

        var body = Data()
        body.append("--\(boundary)\r\n")
        body.append("Content-Disposition: form-data; name=\"file\"; filename=\"\(filename)\"\r\n")
        body.append("Content-Type: application/octet-stream\r\n\r\n")
        body.append(fileData)

        if let path = options?.path {
            body.append("\r\n--\(boundary)\r\n")
            body.append("Content-Disposition: form-data; name=\"path\"\r\n\r\n")
            body.append(path)
        }

        body.append("\r\n--\(boundary)--\r\n")

        guard let url = URL(string: http.baseURL + "/api/v1/adminmedia") else {
            throw APIError(statusCode: 0, message: "Invalid URL for admin media upload")
        }

        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.httpBody = body
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")

        let (responseData, response) = try await http.executeRaw(request)

        let statusCode = response.statusCode
        if statusCode < 200 || statusCode >= 300 {
            let bodyStr = String(data: responseData, encoding: .utf8) ?? ""
            throw APIError(statusCode: statusCode, body: bodyStr)
        }

        return try JSON.decoder.decode(AdminMedia.self, from: responseData)
    }
}

private extension Data {
    mutating func append(_ string: String) {
        if let data = string.data(using: .utf8) {
            append(data)
        }
    }
}
