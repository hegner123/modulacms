import Foundation

/// Response from `GET /api/v1/environment`.
///
/// Describes the server's current deployment environment and stage classification.
public struct EnvironmentResponse: Codable, Sendable {
    /// Raw environment identifier (e.g. "development", "staging", "production").
    public let environment: String

    /// Normalized classification derived from the environment
    /// (e.g. "local", "preview", "production").
    public let stage: String
}

/// Provides access to the server's environment metadata.
///
/// The environment endpoint is unauthenticated and useful for conditionally
/// enabling client features based on the deployment target.
///
/// ```swift
/// let env = try await client.environment.get()
/// if env.stage == "production" {
///     // enable production-only features
/// }
/// ```
public final class EnvironmentResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns the server's current environment and stage classification.
    public func get() async throws -> EnvironmentResponse {
        try await http.get(path: "/api/v1/environment")
    }
}
