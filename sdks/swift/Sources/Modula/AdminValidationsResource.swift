import Foundation

/// Provides CRUD operations for admin validations, plus search by name.
public final class AdminValidationsResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    // MARK: - CRUD

    /// List all admin validations.
    public func list() async throws -> [AdminValidation] {
        try await http.get(path: "/api/v1/admin/validations")
    }

    /// Get a single admin validation by ID.
    public func get(id: AdminValidationID) async throws -> AdminValidation {
        try await http.get(path: "/api/v1/admin/validations/\(id.rawValue)")
    }

    /// Create a new admin validation.
    public func create(params: CreateAdminValidationParams) async throws -> AdminValidation {
        try await http.post(path: "/api/v1/admin/validations", body: params)
    }

    /// Update an existing admin validation.
    public func update(id: AdminValidationID, params: UpdateAdminValidationParams) async throws -> AdminValidation {
        try await http.put(path: "/api/v1/admin/validations/\(id.rawValue)", body: params)
    }

    /// Delete an admin validation by ID.
    public func delete(id: AdminValidationID) async throws {
        try await http.delete(path: "/api/v1/admin/validations/\(id.rawValue)")
    }

    // MARK: - Search

    /// Search admin validations by name substring.
    public func search(name: String) async throws -> [AdminValidation] {
        let queryItems = [URLQueryItem(name: "name", value: name)]
        return try await http.get(path: "/api/v1/admin/validations/search", queryItems: queryItems)
    }
}
