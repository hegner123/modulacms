import Foundation

// MARK: - Shared Response Types

/// Response from a max sort order query.
struct MaxSortOrderResponse: Codable, Sendable {
    let maxSortOrder: Int64

    enum CodingKeys: String, CodingKey {
        case maxSortOrder = "max_sort_order"
    }
}

/// Request body for updating a sort order.
struct UpdateSortOrderBody: Codable, Sendable {
    let sortOrder: Int64

    enum CodingKeys: String, CodingKey {
        case sortOrder = "sort_order"
    }
}

// MARK: - Datatype Sort Order

/// Provides sort order management for public datatypes.
///
/// ```swift
/// let max = try await client.datatypeSortOrder.maxSortOrder()
/// try await client.datatypeSortOrder.updateSortOrder(id: dtID, sortOrder: max + 1)
/// ```
public final class DatatypeSortOrderResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Sets the sort order position for a datatype within its parent group.
    ///
    /// - Parameters:
    ///   - id: The datatype to update.
    ///   - sortOrder: The new sort order value. Lower values appear first.
    public func updateSortOrder(id: DatatypeID, sortOrder: Int64) async throws {
        let body = UpdateSortOrderBody(sortOrder: sortOrder)
        let _: StatusResponse = try await http.put(path: "/api/v1/datatype/\(id.rawValue)/sort-order", body: body)
    }

    /// Returns the highest sort order value currently assigned to datatypes
    /// under the given parent. Returns 0 if no datatypes exist under the parent.
    ///
    /// - Parameter parentID: Filter by parent datatype. Pass `nil` for root-level datatypes.
    /// - Returns: The maximum sort order value.
    public func maxSortOrder(parentID: DatatypeID? = nil) async throws -> Int64 {
        var queryItems: [URLQueryItem]?
        if let parentID {
            queryItems = [URLQueryItem(name: "parent_id", value: parentID.rawValue)]
        }
        let result: MaxSortOrderResponse = try await http.get(path: "/api/v1/datatype/max-sort-order", queryItems: queryItems)
        return result.maxSortOrder
    }
}

// MARK: - Field Sort Order

/// Provides sort order management for fields within a datatype.
///
/// ```swift
/// let max = try await client.fieldSortOrder.maxSortOrder(parentID: dtID)
/// try await client.fieldSortOrder.updateSortOrder(id: fieldID, sortOrder: max + 1)
/// ```
public final class FieldSortOrderResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Sets the sort order position for a field within its parent datatype.
    ///
    /// - Parameters:
    ///   - id: The field to update.
    ///   - sortOrder: The new sort order value. Lower values appear first.
    public func updateSortOrder(id: FieldID, sortOrder: Int64) async throws {
        let body = UpdateSortOrderBody(sortOrder: sortOrder)
        let _: StatusResponse = try await http.put(path: "/api/v1/fields/\(id.rawValue)/sort-order", body: body)
    }

    /// Returns the highest sort order value currently assigned to fields
    /// under the given parent datatype. Returns 0 if the datatype has no fields.
    ///
    /// - Parameter parentID: The datatype whose fields to query.
    /// - Returns: The maximum sort order value.
    public func maxSortOrder(parentID: DatatypeID) async throws -> Int64 {
        let queryItems = [URLQueryItem(name: "parent_id", value: parentID.rawValue)]
        let result: MaxSortOrderResponse = try await http.get(path: "/api/v1/fields/max-sort-order", queryItems: queryItems)
        return result.maxSortOrder
    }
}

// MARK: - Admin Datatype Sort Order

/// Provides sort order management for admin datatypes.
///
/// ```swift
/// let max = try await client.adminDatatypeSortOrder.maxSortOrder()
/// try await client.adminDatatypeSortOrder.updateSortOrder(id: adtID, sortOrder: max + 1)
/// ```
public final class AdminDatatypeSortOrderResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Sets the sort order position for an admin datatype within its parent group.
    ///
    /// - Parameters:
    ///   - id: The admin datatype to update.
    ///   - sortOrder: The new sort order value. Lower values appear first.
    public func updateSortOrder(id: AdminDatatypeID, sortOrder: Int64) async throws {
        let body = UpdateSortOrderBody(sortOrder: sortOrder)
        let _: StatusResponse = try await http.put(path: "/api/v1/admindatatypes/\(id.rawValue)/sort-order", body: body)
    }

    /// Returns the highest sort order value currently assigned to admin datatypes
    /// under the given parent. Returns 0 if no admin datatypes exist under the parent.
    ///
    /// - Parameter parentID: Filter by parent admin datatype. Pass `nil` for root-level.
    /// - Returns: The maximum sort order value.
    public func maxSortOrder(parentID: AdminDatatypeID? = nil) async throws -> Int64 {
        var queryItems: [URLQueryItem]?
        if let parentID {
            queryItems = [URLQueryItem(name: "parent_id", value: parentID.rawValue)]
        }
        let result: MaxSortOrderResponse = try await http.get(path: "/api/v1/admindatatypes/max-sort-order", queryItems: queryItems)
        return result.maxSortOrder
    }
}

// MARK: - Status Response

/// Generic status response from sort order update endpoints.
struct StatusResponse: Codable, Sendable {
    let status: String
}
