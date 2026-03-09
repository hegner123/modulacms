import Foundation

// MARK: - Content Heal Types

public struct HealRepair: Codable, Sendable {
    public let rowID: String
    public let column: String
    public let oldValue: String
    public let newValue: String

    enum CodingKeys: String, CodingKey {
        case rowID = "row_id"
        case column
        case oldValue = "old_value"
        case newValue = "new_value"
    }
}

public struct HealReport: Codable, Sendable {
    public let dryRun: Bool
    public let contentDataScanned: Int
    public let contentDataRepairs: [HealRepair]
    public let contentFieldScanned: Int
    public let contentFieldRepairs: [HealRepair]

    enum CodingKeys: String, CodingKey {
        case dryRun = "dry_run"
        case contentDataScanned = "content_data_scanned"
        case contentDataRepairs = "content_data_repairs"
        case contentFieldScanned = "content_field_scanned"
        case contentFieldRepairs = "content_field_repairs"
    }
}

// MARK: - Content Heal Resource

public final class ContentHealResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Scan content_data and content_field rows for malformed IDs and repair them.
    /// Pass `dryRun: true` to preview repairs without writing changes.
    public func heal(dryRun: Bool = false) async throws -> HealReport {
        var path = "/api/v1/admin/content/heal"
        if dryRun {
            path += "?dry_run=true"
        }
        return try await http.post(path: path, body: Empty?.none)
    }
}

/// Empty body for POST requests that don't need a request body.
private struct Empty: Encodable {}
