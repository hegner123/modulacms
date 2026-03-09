import Foundation

// MARK: - Deploy Types

public struct DeployHealthResponse: Codable, Sendable {
    public let status: String
    public let version: String
    public let nodeID: String

    enum CodingKeys: String, CodingKey {
        case status
        case version
        case nodeID = "node_id"
    }
}

public struct DeploySyncError: Codable, Sendable {
    public let table: String
    public let phase: String
    public let message: String
    public let rowID: String?

    enum CodingKeys: String, CodingKey {
        case table
        case phase
        case message
        case rowID = "row_id"
    }
}

public struct DeploySyncResult: Codable, Sendable {
    public let success: Bool
    public let dryRun: Bool
    public let strategy: String
    public let tablesAffected: [String]
    public let rowCounts: [String: Int]
    public let backupPath: String
    public let snapshotID: String
    public let duration: String
    public let errors: [DeploySyncError]?
    public let warnings: [String]?

    enum CodingKeys: String, CodingKey {
        case success
        case dryRun = "dry_run"
        case strategy
        case tablesAffected = "tables_affected"
        case rowCounts = "row_counts"
        case backupPath = "backup_path"
        case snapshotID = "snapshot_id"
        case duration
        case errors
        case warnings
    }
}

// MARK: - Deploy Resource

public final class DeployResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Check deploy endpoint health and version info.
    public func health() async throws -> DeployHealthResponse {
        try await http.get(path: "/api/v1/deploy/health")
    }

    /// Export CMS data as a sync payload. Optionally limit to specific tables.
    /// Returns raw JSON bytes representing the sync payload, suitable for passing to `importPayload` or `dryRunImport`.
    public func export(tables: [String]? = nil) async throws -> Data {
        let path = "/api/v1/deploy/export"
        if let tables, !tables.isEmpty {
            return try await http.post(path: path, body: DeployExportBody(tables: tables)) as Data
        }
        return try await http.post(path: path, body: DeployEmptyBody?.none) as Data
    }

    /// Import a sync payload into this instance.
    public func importPayload(_ payload: Data) async throws -> DeploySyncResult {
        try await doImport(payload: payload, dryRun: false)
    }

    /// Validate a sync payload without modifying the database.
    public func dryRunImport(_ payload: Data) async throws -> DeploySyncResult {
        try await doImport(payload: payload, dryRun: true)
    }

    private func doImport(payload: Data, dryRun: Bool) async throws -> DeploySyncResult {
        var path = "/api/v1/deploy/import"
        if dryRun {
            path += "?dry_run=true"
        }

        guard let url = URL(string: http.baseURL + path) else {
            throw APIError(statusCode: 0, message: "Invalid URL for deploy import")
        }

        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.httpBody = payload
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let (data, response) = try await http.executeRaw(request)

        let statusCode = response.statusCode
        if statusCode < 200 || statusCode >= 300 {
            let bodyStr = String(data: data, encoding: .utf8) ?? ""
            throw APIError(statusCode: statusCode, body: bodyStr)
        }

        return try JSON.decoder.decode(DeploySyncResult.self, from: data)
    }
}

private struct DeployExportBody: Encodable {
    let tables: [String]
}

private struct DeployEmptyBody: Encodable {}
