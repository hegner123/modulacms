import Foundation

public final class ConfigResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Get the current config (redacted). Optional category filter.
    public func get(category: String? = nil) async throws -> ConfigGetResponse {
        var queryItems: [URLQueryItem]?
        if let category {
            queryItems = [URLQueryItem(name: "category", value: category)]
        }
        return try await http.get(path: "/api/v1/admin/config", queryItems: queryItems)
    }

    /// Update config fields.
    public func update(updates: [String: JSONValue]) async throws -> ConfigUpdateResponse {
        let wrapper = ConfigUpdateRequest(updates: updates)
        return try await http.patch(path: "/api/v1/admin/config", body: wrapper)
    }

    /// Get field metadata registry.
    public func meta() async throws -> ConfigMetaResponse {
        try await http.get(path: "/api/v1/admin/config/meta")
    }

    /// Get a searchable index of all config fields, combining field
    /// registry metadata with help text.
    public func searchIndex() async throws -> [ConfigSearchIndexEntry] {
        try await http.get(path: "/api/v1/admin/config/search-index")
    }
}

// MARK: - Types

public struct ConfigFieldMeta: Codable, Sendable {
    public let jsonKey: String
    public let label: String
    public let category: String
    public let hotReloadable: Bool
    public let sensitive: Bool
    public let required: Bool
    public let description: String

    enum CodingKeys: String, CodingKey {
        case jsonKey = "json_key"
        case label
        case category
        case hotReloadable = "hot_reloadable"
        case sensitive
        case required
        case description
    }
}

public struct ConfigGetResponse: Codable, Sendable {
    public let config: [String: JSONValue]
}

public struct ConfigUpdateResponse: Codable, Sendable {
    public let ok: Bool
    public let config: [String: JSONValue]
    public let restartRequired: [String]?
    public let warnings: [String]?

    enum CodingKeys: String, CodingKey {
        case ok
        case config
        case restartRequired = "restart_required"
        case warnings
    }
}

public struct ConfigMetaResponse: Codable, Sendable {
    public let fields: [ConfigFieldMeta]
    public let categories: [String]
}

public struct ConfigSearchIndexEntry: Codable, Sendable {
    public let key: String
    public let label: String
    public let category: String
    public let categoryLabel: String
    public let description: String
    public let helpText: String
    public let defaultValue: String
    public let example: String
    public let hotReloadable: Bool
    public let sensitive: Bool
    public let required: Bool

    enum CodingKeys: String, CodingKey {
        case key
        case label
        case category
        case categoryLabel = "category_label"
        case description
        case helpText = "help_text"
        case defaultValue = "default"
        case example
        case hotReloadable = "hot_reloadable"
        case sensitive
        case required
    }
}

// MARK: - Private

private struct ConfigUpdateRequest: Encodable {
    let updates: [String: JSONValue]
}
