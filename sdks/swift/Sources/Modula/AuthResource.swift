import Foundation

public final class AuthResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    public func login(params: LoginParams) async throws -> LoginResponse {
        try await http.post(path: "/api/v1/auth/login", body: params)
    }

    public func logout() async throws {
        try await http.postNoBody(path: "/api/v1/auth/logout")
    }

    public func me() async throws -> User {
        try await http.get(path: "/api/v1/auth/me")
    }

    public func register(params: CreateUserParams) async throws -> User {
        try await http.post(path: "/api/v1/auth/register", body: params)
    }

    public func resetPassword(params: ResetPasswordParams) async throws {
        try await http.post(path: "/api/v1/auth/reset", body: params) as Void
    }

    public func requestPasswordReset(params: RequestPasswordResetParams) async throws -> MessageResponse {
        try await http.post(path: "/api/v1/auth/request-password-reset", body: params)
    }

    public func confirmPasswordReset(params: ConfirmPasswordResetParams) async throws -> MessageResponse {
        try await http.post(path: "/api/v1/auth/confirm-password-reset", body: params)
    }
}
