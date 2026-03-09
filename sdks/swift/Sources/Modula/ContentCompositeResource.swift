import Foundation

/// Provides composite content operations that combine multiple steps into single API calls.
///
/// - ``createWithFields(params:)`` creates a content_data node and populates its fields atomically.
/// - ``deleteRecursive(id:)`` deletes a content node and all of its descendants.
public final class ContentCompositeResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Create a content_data node and its field values in a single request.
    ///
    /// - Parameter params: The content data and field values to create.
    /// - Returns: A ``ContentCreateResponse`` with the created node, fields, and any errors.
    public func createWithFields(params: ContentCreateParams) async throws -> ContentCreateResponse {
        try await http.post(path: "/api/v1/content/create", body: params)
    }

    /// Delete a content node and all of its descendants recursively.
    ///
    /// - Parameter id: The content data ID of the root node to delete.
    /// - Returns: A ``RecursiveDeleteResponse`` with the deleted root, total count, and all deleted IDs.
    public func deleteRecursive(id: ContentID) async throws -> RecursiveDeleteResponse {
        let queryItems = [
            URLQueryItem(name: "q", value: id.rawValue),
            URLQueryItem(name: "recursive", value: "true"),
        ]
        return try await http.delete(path: "/api/v1/contentdata/", queryItems: queryItems)
    }
}
