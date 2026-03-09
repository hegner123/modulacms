import Foundation

public struct APIError: Error, LocalizedError, Sendable {
    public let statusCode: Int
    public let message: String
    public let body: String

    public init(statusCode: Int, message: String = "", body: String = "") {
        self.statusCode = statusCode
        self.message = message
        self.body = body
    }

    public var errorDescription: String? {
        if !message.isEmpty {
            return "modula: \(statusCode) \(message)"
        }
        return "modula: \(statusCode) \(HTTPURLResponse.localizedString(forStatusCode: statusCode))"
    }
}

public func isNotFound(_ error: Error) -> Bool {
    guard let apiError = error as? APIError else { return false }
    return apiError.statusCode == 404
}

public func isUnauthorized(_ error: Error) -> Bool {
    guard let apiError = error as? APIError else { return false }
    return apiError.statusCode == 401
}

public func isDuplicateMedia(_ error: Error) -> Bool {
    guard let apiError = error as? APIError else { return false }
    return apiError.statusCode == 409
}

public func isInvalidMediaPath(_ error: Error) -> Bool {
    guard let apiError = error as? APIError else { return false }
    return apiError.statusCode == 400
        && (apiError.body.contains("path traversal") || apiError.body.contains("invalid character in path"))
}
