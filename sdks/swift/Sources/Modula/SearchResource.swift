import Foundation

public struct SearchOptions: Sendable {
    public var type: String
    public var locale: String
    public var limit: Int
    public var offset: Int
    public var prefix: Bool?

    public init(
        type: String = "",
        locale: String = "",
        limit: Int = 0,
        offset: Int = 0,
        prefix: Bool? = nil
    ) {
        self.type = type
        self.locale = locale
        self.limit = limit
        self.offset = offset
        self.prefix = prefix
    }
}

public struct SearchResult: Codable, Sendable {
    public let id: String
    public let contentDataID: String
    public let routeSlug: String
    public let routeTitle: String
    public let datatypeName: String
    public let datatypeLabel: String
    public let locale: String?
    public let section: String?
    public let sectionAnchor: String?
    public let score: Double
    public let snippet: String
    public let publishedAt: String

    enum CodingKeys: String, CodingKey {
        case id
        case contentDataID = "content_data_id"
        case routeSlug = "route_slug"
        case routeTitle = "route_title"
        case datatypeName = "datatype_name"
        case datatypeLabel = "datatype_label"
        case locale
        case section
        case sectionAnchor = "section_anchor"
        case score
        case snippet
        case publishedAt = "published_at"
    }
}

public struct SearchResponse: Codable, Sendable {
    public let query: String
    public let results: [SearchResult]
    public let total: Int
    public let limit: Int
    public let offset: Int
}

/// Response from `POST /api/v1/admin/search/rebuild` after a successful
/// search index rebuild.
public struct SearchRebuildResponse: Codable, Sendable {
    /// "ok" on success.
    public let status: String

    /// Number of documents in the rebuilt index.
    public let documents: Int

    /// Number of unique terms in the rebuilt index.
    public let terms: Int

    /// Estimated memory usage of the rebuilt index in bytes.
    public let memBytes: Int64

    enum CodingKeys: String, CodingKey {
        case status
        case documents
        case terms
        case memBytes = "mem_bytes"
    }
}

public final class SearchResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    public func search(_ query: String, options: SearchOptions = SearchOptions()) async throws -> SearchResponse {
        var queryItems: [URLQueryItem] = [
            URLQueryItem(name: "q", value: query)
        ]
        if !options.type.isEmpty {
            queryItems.append(URLQueryItem(name: "type", value: options.type))
        }
        if !options.locale.isEmpty {
            queryItems.append(URLQueryItem(name: "locale", value: options.locale))
        }
        if options.limit > 0 {
            queryItems.append(URLQueryItem(name: "limit", value: String(options.limit)))
        }
        if options.offset > 0 {
            queryItems.append(URLQueryItem(name: "offset", value: String(options.offset)))
        }
        if let prefix = options.prefix {
            queryItems.append(URLQueryItem(name: "prefix", value: String(prefix)))
        }
        return try await http.get(path: "/api/v1/search", queryItems: queryItems)
    }

    /// Triggers a full rebuild of the search index and returns the resulting
    /// index statistics. Requires `search:update` permission. This is an admin
    /// operation and should not be called in response to user actions.
    public func rebuild() async throws -> SearchRebuildResponse {
        try await http.postNoBody(path: "/api/v1/admin/search/rebuild")
    }
}
