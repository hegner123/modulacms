import Foundation

public final class ImportResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    public func contentful<T: Encodable & Sendable>(data: T) async throws -> Data {
        try await http.post(path: "/api/v1/import/contentful", body: data) as Data
    }

    public func sanity<T: Encodable & Sendable>(data: T) async throws -> Data {
        try await http.post(path: "/api/v1/import/sanity", body: data) as Data
    }

    public func strapi<T: Encodable & Sendable>(data: T) async throws -> Data {
        try await http.post(path: "/api/v1/import/strapi", body: data) as Data
    }

    public func wordpress<T: Encodable & Sendable>(data: T) async throws -> Data {
        try await http.post(path: "/api/v1/import/wordpress", body: data) as Data
    }

    public func clean<T: Encodable & Sendable>(data: T) async throws -> Data {
        try await http.post(path: "/api/v1/import/clean", body: data) as Data
    }

    public func bulk<T: Encodable & Sendable>(data: T) async throws -> Data {
        try await http.post(path: "/api/v1/import", body: data) as Data
    }
}
