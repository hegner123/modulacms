import Foundation

final class HTTPClient: Sendable {
    let baseURL: String
    let apiKey: String
    let session: URLSession

    init(baseURL: String, apiKey: String, session: URLSession) {
        self.baseURL = baseURL
        self.apiKey = apiKey
        self.session = session
    }

    func get<T: Decodable>(path: String, queryItems: [URLQueryItem]? = nil) async throws -> T {
        let request = try buildRequest(method: "GET", path: path, queryItems: queryItems)
        let (data, response) = try await session.data(for: request)
        try checkStatus(response: response, data: data)
        return try JSON.decoder.decode(T.self, from: data)
    }

    func post<T: Decodable, B: Encodable>(path: String, body: B) async throws -> T {
        let request = try buildJSONRequest(method: "POST", path: path, body: body)
        let (data, response) = try await session.data(for: request)
        try checkStatus(response: response, data: data)
        return try JSON.decoder.decode(T.self, from: data)
    }

    func post<B: Encodable>(path: String, body: B) async throws {
        let request = try buildJSONRequest(method: "POST", path: path, body: body)
        let (data, response) = try await session.data(for: request)
        try checkStatus(response: response, data: data)
    }

    func postNoBody(path: String) async throws {
        let request = try buildRequest(method: "POST", path: path)
        let (data, response) = try await session.data(for: request)
        try checkStatus(response: response, data: data)
    }

    func put<T: Decodable, B: Encodable>(path: String, body: B) async throws -> T {
        let request = try buildJSONRequest(method: "PUT", path: path, body: body)
        let (data, response) = try await session.data(for: request)
        try checkStatus(response: response, data: data)
        return try JSON.decoder.decode(T.self, from: data)
    }

    func patch<T: Decodable, B: Encodable>(path: String, body: B) async throws -> T {
        let request = try buildJSONRequest(method: "PATCH", path: path, body: body)
        let (data, response) = try await session.data(for: request)
        try checkStatus(response: response, data: data)
        return try JSON.decoder.decode(T.self, from: data)
    }

    func delete(path: String, queryItems: [URLQueryItem]? = nil) async throws {
        let request = try buildRequest(method: "DELETE", path: path, queryItems: queryItems)
        let (data, response) = try await session.data(for: request)
        try checkStatus(response: response, data: data)
    }

    func delete<T: Decodable>(path: String, queryItems: [URLQueryItem]? = nil) async throws -> T {
        let request = try buildRequest(method: "DELETE", path: path, queryItems: queryItems)
        let (data, response) = try await session.data(for: request)
        try checkStatus(response: response, data: data)
        return try JSON.decoder.decode(T.self, from: data)
    }

    func executeRaw(_ request: URLRequest) async throws -> (Data, HTTPURLResponse) {
        var req = request
        setAuth(&req)
        let data: Data
        let response: URLResponse
        if let body = req.httpBody {
            req.httpBody = nil
            (data, response) = try await session.upload(for: req, from: body)
        } else {
            (data, response) = try await session.data(for: req)
        }
        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError(statusCode: 0, message: "Invalid response type")
        }
        return (data, httpResponse)
    }

    // MARK: - Private

    private func buildRequest(method: String, path: String, queryItems: [URLQueryItem]? = nil) throws -> URLRequest {
        var urlString = baseURL + path
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
        request.httpMethod = method
        setAuth(&request)
        return request
    }

    private func buildJSONRequest<B: Encodable>(method: String, path: String, body: B) throws -> URLRequest {
        var request = try buildRequest(method: method, path: path)
        request.httpBody = try JSON.encoder.encode(body)
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        return request
    }

    private func setAuth(_ request: inout URLRequest) {
        if !apiKey.isEmpty {
            request.setValue("Bearer \(apiKey)", forHTTPHeaderField: "Authorization")
        }
    }

    private func checkStatus(response: URLResponse, data: Data) throws {
        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError(statusCode: 0, message: "Invalid response type")
        }
        let statusCode = httpResponse.statusCode
        if statusCode < 200 || statusCode >= 300 {
            throw buildError(statusCode: statusCode, data: data)
        }
    }

    private func buildError(statusCode: Int, data: Data) -> APIError {
        let bodyStr = String(data: data, encoding: .utf8)?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""

        var message = ""
        if let parsed = try? JSONSerialization.jsonObject(with: data) as? [String: Any] {
            if let msg = parsed["message"] as? String, !msg.isEmpty {
                message = msg
            } else if let errMsg = parsed["error"] as? String, !errMsg.isEmpty {
                message = errMsg
            }
        }

        return APIError(statusCode: statusCode, message: message, body: bodyStr)
    }
}
