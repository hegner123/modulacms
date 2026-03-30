import Foundation

/// A single recent activity event from the audit log.
///
/// Returned by ``ActivityResource/listRecent(limit:)`` and suitable for building
/// activity feeds and admin dashboards.
public struct ActivityItem: Codable, Sendable {
    /// Unique identifier for this change event.
    public let eventID: String

    /// Database table that was modified (e.g. "content_data", "media").
    public let tableName: String

    /// Primary key of the affected record.
    public let recordID: String

    /// Type of database operation (e.g. "INSERT", "UPDATE", "DELETE").
    public let operation: String

    /// Human-readable action label (e.g. "create", "update", "delete").
    public let action: String

    /// User who performed the action, or `nil` for system events.
    public let actor: ActivityActor?

    /// Wall-clock time when the event occurred (ISO 8601).
    public let timestamp: String

    enum CodingKeys: String, CodingKey {
        case eventID = "event_id"
        case tableName = "table_name"
        case recordID = "record_id"
        case operation
        case action
        case actor
        case timestamp
    }
}

/// Identifies the user who performed an activity event.
public struct ActivityActor: Codable, Sendable {
    public let userID: String
    public let username: String
    public let name: String
    public let email: String
    public let role: String

    enum CodingKeys: String, CodingKey {
        case userID = "user_id"
        case username
        case name
        case email
        case role
    }
}

/// Provides access to the audit activity feed.
///
/// Requires `audit:read` permission.
///
/// ```swift
/// let items = try await client.activity.listRecent(limit: 10)
/// for item in items {
///     print("\(item.action) on \(item.tableName) by \(item.actor?.username ?? "system")")
/// }
/// ```
public final class ActivityResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns the most recent activity events, up to the given limit.
    ///
    /// - Parameter limit: Maximum number of events to return. Server default is 25, maximum is 100.
    public func listRecent(limit: Int = 25) async throws -> [ActivityItem] {
        var queryItems: [URLQueryItem] = []
        if limit > 0 {
            queryItems.append(URLQueryItem(name: "limit", value: String(limit)))
        }
        return try await http.get(path: "/api/v1/activity/recent", queryItems: queryItems.isEmpty ? nil : queryItems)
    }
}
