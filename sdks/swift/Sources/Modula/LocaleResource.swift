import Foundation

/// Provides CRUD operations for locales and translation creation.
public final class LocaleResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    // MARK: - CRUD

    /// List all locales.
    public func list() async throws -> [Locale] {
        try await http.get(path: "/api/v1/locales")
    }

    /// Get a single locale by ID.
    public func get(id: LocaleID) async throws -> Locale {
        try await http.get(
            path: "/api/v1/locales/",
            queryItems: [URLQueryItem(name: "q", value: id.rawValue)]
        )
    }

    /// Create a new locale.
    public func create(params: CreateLocaleRequest) async throws -> Locale {
        try await http.post(path: "/api/v1/locales", body: params)
    }

    /// Update an existing locale.
    public func update(params: UpdateLocaleRequest) async throws -> Locale {
        try await http.put(path: "/api/v1/locales/", body: params)
    }

    /// Delete a locale by ID.
    public func delete(id: LocaleID) async throws {
        try await http.delete(
            path: "/api/v1/locales/",
            queryItems: [URLQueryItem(name: "q", value: id.rawValue)]
        )
    }

    // MARK: - Translations

    /// Create translated content fields for a content data node in the given locale.
    public func createTranslation(contentDataID: String, req: CreateTranslationRequest) async throws -> CreateTranslationResponse {
        try await http.post(
            path: "/api/v1/admin/contentdata/\(contentDataID)/translations",
            body: req
        )
    }

    /// Create translated content fields for an admin content data node in the given locale.
    public func createAdminTranslation(adminContentDataID: String, req: CreateTranslationRequest) async throws -> CreateTranslationResponse {
        try await http.post(
            path: "/api/v1/admin/admincontentdata/\(adminContentDataID)/translations",
            body: req
        )
    }
}
