import Foundation

/// Provides CRUD operations for public validations, plus search by name.
public final class ValidationsResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    // MARK: - CRUD

    /// List all validations.
    public func list() async throws -> [Validation] {
        try await http.get(path: "/api/v1/validations")
    }

    /// Get a single validation by ID.
    public func get(id: ValidationID) async throws -> Validation {
        try await http.get(path: "/api/v1/validations/\(id.rawValue)")
    }

    /// Create a new validation.
    public func create(params: CreateValidationParams) async throws -> Validation {
        try await http.post(path: "/api/v1/validations", body: params)
    }

    /// Update an existing validation.
    public func update(id: ValidationID, params: UpdateValidationParams) async throws -> Validation {
        try await http.put(path: "/api/v1/validations/\(id.rawValue)", body: params)
    }

    /// Delete a validation by ID.
    public func delete(id: ValidationID) async throws {
        try await http.delete(path: "/api/v1/validations/\(id.rawValue)")
    }

    // MARK: - Search

    /// Search validations by name substring.
    public func search(name: String) async throws -> [Validation] {
        let queryItems = [URLQueryItem(name: "name", value: name)]
        return try await http.get(path: "/api/v1/validations/search", queryItems: queryItems)
    }
}
