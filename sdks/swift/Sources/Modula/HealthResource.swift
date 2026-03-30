import Foundation

/// Response from `GET /api/v1/health`.
///
/// Reports the server's operational status including individual subsystem checks.
/// A non-error response does not guarantee all subsystems are healthy -- check
/// ``status`` and ``checks`` for application-level health information.
public struct HealthResponse: Codable, Sendable {
    /// Aggregate health state (e.g. "ok", "degraded").
    public let status: String

    /// Environment identifier (e.g. "development", "staging", "production").
    public let environment: String

    /// Maps subsystem names to their pass/fail status.
    /// Common keys: "database", "storage", "plugins".
    public let checks: [String: Bool]

    /// Human-readable additional information for each subsystem.
    public let details: [String: String]?
}

/// Provides health check operations for monitoring the CMS server.
///
/// The health endpoint is unauthenticated and suitable for load balancer probes,
/// uptime monitoring, and readiness checks.
///
/// ```swift
/// let health = try await client.health.check()
/// if health.status != "ok" {
///     print("Server degraded: \(health.details ?? [:])")
/// }
/// ```
public final class HealthResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns the server health status, including individual subsystem checks
    /// and their details.
    public func check() async throws -> HealthResponse {
        try await http.get(path: "/api/v1/health")
    }
}
