import Foundation

public struct QueryParams: Sendable {
    public var sort: String
    public var limit: Int
    public var offset: Int
    public var locale: String
    public var status: String
    public var filters: [String: String]

    public init(
        sort: String = "",
        limit: Int = 0,
        offset: Int = 0,
        locale: String = "",
        status: String = "",
        filters: [String: String] = [:]
    ) {
        self.sort = sort
        self.limit = limit
        self.offset = offset
        self.locale = locale
        self.status = status
        self.filters = filters
    }
}

public struct QueryItem: Codable, Sendable {
    public let contentDataID: String
    public let datatypeID: String
    public let authorID: String
    public let status: String
    public let dateCreated: String
    public let dateModified: String
    public let publishedAt: String
    public let fields: [String: String]

    enum CodingKeys: String, CodingKey {
        case contentDataID = "content_data_id"
        case datatypeID = "datatype_id"
        case authorID = "author_id"
        case status
        case dateCreated = "date_created"
        case dateModified = "date_modified"
        case publishedAt = "published_at"
        case fields
    }
}

public struct QueryDatatype: Codable, Sendable {
    public let name: String
    public let label: String
}

public struct QueryResult: Codable, Sendable {
    public let data: [QueryItem]
    public let total: Int
    public let limit: Int
    public let offset: Int
    public let datatype: QueryDatatype
}

public final class QueryResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    public func query(datatype: String, params: QueryParams = QueryParams()) async throws -> QueryResult {
        var queryItems: [URLQueryItem] = []
        if !params.sort.isEmpty {
            queryItems.append(URLQueryItem(name: "sort", value: params.sort))
        }
        if params.limit > 0 {
            queryItems.append(URLQueryItem(name: "limit", value: String(params.limit)))
        }
        if params.offset > 0 {
            queryItems.append(URLQueryItem(name: "offset", value: String(params.offset)))
        }
        if !params.locale.isEmpty {
            queryItems.append(URLQueryItem(name: "locale", value: params.locale))
        }
        if !params.status.isEmpty {
            queryItems.append(URLQueryItem(name: "status", value: params.status))
        }
        for (key, value) in params.filters {
            queryItems.append(URLQueryItem(name: key, value: value))
        }

        let encoded = datatype.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? datatype
        let path = "/api/v1/query/" + encoded
        return try await http.get(path: path, queryItems: queryItems.isEmpty ? nil : queryItems)
    }
}
