import Foundation

public final class Resource<
    Entity: Decodable & Sendable,
    CreateParams: Encodable & Sendable,
    UpdateParams: Encodable & Sendable,
    ID: ResourceID
>: Sendable {
    let path: String
    let http: HTTPClient

    init(path: String, http: HTTPClient) {
        self.path = path
        self.http = http
    }

    public func list() async throws -> [Entity] {
        try await http.get(path: path)
    }

    public func get(id: ID) async throws -> Entity {
        let queryItems = [URLQueryItem(name: "q", value: id.rawValue)]
        return try await http.get(path: path + "/", queryItems: queryItems)
    }

    public func create(params: CreateParams) async throws -> Entity {
        try await http.post(path: path, body: params)
    }

    public func update(params: UpdateParams) async throws -> Entity {
        try await http.put(path: path + "/", body: params)
    }

    public func delete(id: ID) async throws {
        let queryItems = [URLQueryItem(name: "q", value: id.rawValue)]
        try await http.delete(path: path + "/", queryItems: queryItems)
    }

    public func listPaginated(params: PaginationParams) async throws -> PaginatedResponse<Entity> {
        let queryItems = [
            URLQueryItem(name: "limit", value: String(params.limit)),
            URLQueryItem(name: "offset", value: String(params.offset)),
        ]
        return try await http.get(path: path, queryItems: queryItems)
    }

    public func count() async throws -> Int64 {
        let queryItems = [URLQueryItem(name: "count", value: "true")]
        let result: CountResponse = try await http.get(path: path, queryItems: queryItems)
        return result.count
    }

    public func rawList(queryItems: [URLQueryItem]? = nil) async throws -> Data {
        let request = try buildGetRequest(queryItems: queryItems)
        let (data, response) = try await http.executeRaw(request)
        let statusCode = response.statusCode
        if statusCode < 200 || statusCode >= 300 {
            let bodyStr = String(data: data, encoding: .utf8) ?? ""
            throw APIError(statusCode: statusCode, body: bodyStr)
        }
        return data
    }

    private func buildGetRequest(queryItems: [URLQueryItem]?) throws -> URLRequest {
        var urlString = http.baseURL + path
        if let queryItems, !queryItems.isEmpty {
            var components = URLComponents()
            components.queryItems = queryItems
            if let query = components.percentEncodedQuery {
                urlString += "?" + query
            }
        }
        guard let url = URL(string: urlString) else {
            throw APIError(statusCode: 0, message: "Invalid URL: \(urlString)")
        }
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        return request
    }
}

private struct CountResponse: Decodable {
    let count: Int64
}
