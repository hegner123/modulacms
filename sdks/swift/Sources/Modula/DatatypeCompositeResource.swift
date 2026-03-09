import Foundation

/// Provides composite datatype operations for cascade deletion.
///
/// ``deleteCascade(id:)`` deletes a datatype and all content that uses it.
public final class DatatypeCompositeResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Delete a datatype and cascade-delete all content_data rows that use it.
    ///
    /// - Parameter id: The datatype ID to delete.
    /// - Returns: A ``DatatypeCascadeDeleteResponse`` with the deleted datatype ID, content count, and any errors.
    public func deleteCascade(id: DatatypeID) async throws -> DatatypeCascadeDeleteResponse {
        let queryItems = [
            URLQueryItem(name: "q", value: id.rawValue),
            URLQueryItem(name: "cascade", value: "true"),
        ]
        return try await http.delete(path: "/api/v1/datatype/", queryItems: queryItems)
    }
}
