import Foundation

/// Provides bulk content tree operations via `POST /api/v1/content/tree`.
///
/// The tree save endpoint atomically applies creates, deletes, and pointer-field
/// updates to content_data nodes in a single HTTP round-trip. This is the
/// preferred way to persist structural changes from a block editor or tree
/// manipulation UI.
///
/// ## Creates
///
/// New nodes are inserted with server-generated ULIDs. The caller supplies a
/// client-side ID (e.g. a UUID from `UUID().uuidString`) and receives an `idMap`
/// in the response mapping each client ID to the server-generated ULID.
/// Pointer fields (`parentID`, `firstChildID`, etc.) may reference other new
/// nodes by their client IDs -- the server remaps them automatically.
///
/// The server inherits `routeID` from the parent content node, sets `authorID`
/// from the authenticated user, and defaults `status` to `"draft"`.
///
/// ## Deletes
///
/// Nodes listed in `deletes` are removed via audited delete. Deletes are
/// processed before updates so removed nodes don't interfere with pointer rewiring.
///
/// ## Updates
///
/// Each update entry specifies a `contentDataID` and the new values for its
/// four pointer fields (parent, first_child, next_sibling, prev_sibling).
/// The server preserves all non-pointer fields, only overwriting the pointers
/// and bumping `dateModified`.
///
/// ## Partial failures
///
/// The endpoint always returns HTTP 200. Check ``TreeSaveResponse/errors`` for
/// per-node error messages. Created/updated/deleted counts reflect only
/// successful operations.
public final class ContentTreeResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Atomically applies tree structure changes in a single request.
    ///
    /// - Parameter request: The tree save request containing creates, updates, and deletes.
    /// - Returns: A ``TreeSaveResponse`` with counts and an optional ID map.
    ///
    /// ```swift
    /// let response = try await client.contentTree.save(TreeSaveRequest(
    ///     contentID: parentID,
    ///     creates: [TreeNodeCreate(
    ///         clientID: UUID().uuidString,
    ///         datatypeID: dtID,
    ///         parentID: parentID.rawValue
    ///     )],
    ///     updates: [TreeNodeUpdate(
    ///         contentDataID: existingID,
    ///         firstChildID: "temp-uuid-1"
    ///     )],
    ///     deletes: [removedID]
    /// ))
    /// ```
    public func save(_ request: TreeSaveRequest) async throws -> TreeSaveResponse {
        try await http.post(path: "/api/v1/content/tree", body: request)
    }

    /// Returns all content tree nodes belonging to the given route as a flat
    /// array of ``ContentTreeNode`` values.
    ///
    /// The caller reconstructs the tree hierarchy using the pointer fields
    /// (parentID, firstChildID, nextSiblingID, prevSiblingID).
    ///
    /// - Parameter routeID: The route whose content tree to retrieve.
    /// - Returns: A flat array of content tree nodes for the route.
    public func getByRoute(routeID: RouteID) async throws -> [ContentTreeNode] {
        try await http.get(path: "/api/v1/content/tree/\(routeID.rawValue)")
    }
}

/// A new content_data node to insert via the tree save endpoint.
///
/// `clientID` is a caller-generated identifier (e.g. UUID). The server generates
/// a ULID and returns the mapping in ``TreeSaveResponse/idMap``.
public struct TreeNodeCreate: Codable, Sendable {
    /// Caller-generated temporary ID. Must be unique within the request.
    /// Other nodes may reference this ID in their pointer fields.
    public let clientID: String

    /// Datatype to assign. Pass empty string for no datatype.
    public let datatypeID: String

    /// Parent content node, or `nil` for root-level. May be a client ID.
    public let parentID: String?

    /// First child, or `nil`. May be a client ID.
    public let firstChildID: String?

    /// Next sibling, or `nil`. May be a client ID.
    public let nextSiblingID: String?

    /// Previous sibling, or `nil`. May be a client ID.
    public let prevSiblingID: String?

    enum CodingKeys: String, CodingKey {
        case clientID = "client_id"
        case datatypeID = "datatype_id"
        case parentID = "parent_id"
        case firstChildID = "first_child_id"
        case nextSiblingID = "next_sibling_id"
        case prevSiblingID = "prev_sibling_id"
    }

    public init(
        clientID: String,
        datatypeID: String = "",
        parentID: String? = nil,
        firstChildID: String? = nil,
        nextSiblingID: String? = nil,
        prevSiblingID: String? = nil
    ) {
        self.clientID = clientID
        self.datatypeID = datatypeID
        self.parentID = parentID
        self.firstChildID = firstChildID
        self.nextSiblingID = nextSiblingID
        self.prevSiblingID = prevSiblingID
    }
}

/// Pointer-field changes for an existing content_data node.
///
/// Only the four tree-pointer fields are updated; all other fields (route,
/// datatype, author, status, dates) are preserved from the existing row.
public struct TreeNodeUpdate: Codable, Sendable {
    /// ULID of the existing node to update.
    public let contentDataID: ContentID

    /// New parent, or `nil` for SQL NULL. May be a client ID.
    public let parentID: String?

    /// New first child, or `nil`. May be a client ID.
    public let firstChildID: String?

    /// New next sibling, or `nil`. May be a client ID.
    public let nextSiblingID: String?

    /// New previous sibling, or `nil`. May be a client ID.
    public let prevSiblingID: String?

    enum CodingKeys: String, CodingKey {
        case contentDataID = "content_data_id"
        case parentID = "parent_id"
        case firstChildID = "first_child_id"
        case nextSiblingID = "next_sibling_id"
        case prevSiblingID = "prev_sibling_id"
    }

    public init(
        contentDataID: ContentID,
        parentID: String? = nil,
        firstChildID: String? = nil,
        nextSiblingID: String? = nil,
        prevSiblingID: String? = nil
    ) {
        self.contentDataID = contentDataID
        self.parentID = parentID
        self.firstChildID = firstChildID
        self.nextSiblingID = nextSiblingID
        self.prevSiblingID = prevSiblingID
    }
}

/// Request body for `POST /api/v1/content/tree`.
///
/// Atomically applies creates, deletes, and pointer-field updates to
/// content_data nodes in a single HTTP round-trip.
public struct TreeSaveRequest: Codable, Sendable {
    /// Root content node being edited. Used to resolve routeID for new child nodes.
    public let contentID: ContentID

    /// New content_data nodes to insert.
    public let creates: [TreeNodeCreate]?

    /// Existing nodes whose pointer fields changed.
    public let updates: [TreeNodeUpdate]?

    /// Content data IDs to remove.
    public let deletes: [ContentID]?

    enum CodingKeys: String, CodingKey {
        case contentID = "content_id"
        case creates
        case updates
        case deletes
    }

    public init(
        contentID: ContentID,
        creates: [TreeNodeCreate]? = nil,
        updates: [TreeNodeUpdate]? = nil,
        deletes: [ContentID]? = nil
    ) {
        self.contentID = contentID
        self.creates = creates
        self.updates = updates
        self.deletes = deletes
    }
}

/// Response from the tree save endpoint.
///
/// Always returns HTTP 200. Check ``errors`` for per-node failure messages.
/// ``created``, ``updated``, and ``deleted`` counts reflect only successful operations.
public struct TreeSaveResponse: Codable, Sendable {
    /// Number of nodes successfully created.
    public let created: Int

    /// Number of nodes successfully updated.
    public let updated: Int

    /// Number of nodes successfully deleted.
    public let deleted: Int

    /// Maps client-supplied IDs to server-generated ULIDs.
    /// Only present when creates were included in the request.
    public let idMap: [String: String]?

    /// Per-node error messages for partial failures.
    public let errors: [String]?

    enum CodingKeys: String, CodingKey {
        case created
        case updated
        case deleted
        case idMap = "id_map"
        case errors
    }
}

/// A node in the content tree for a route, returned by
/// ``ContentTreeResource/getByRoute(routeID:)``.
///
/// The tree structure is encoded as a doubly-linked sibling list with parent
/// and first-child pointers, allowing O(1) insertion, deletion, and reordering.
public struct ContentTreeNode: Codable, Sendable {
    /// Unique ULID identifying this content node.
    public let contentID: String

    /// Datatype assigned to this node, or `nil` if untyped.
    public let datatypeID: String?

    /// Parent node, or `nil` for root-level nodes.
    public let parentID: String?

    /// Leftmost child, or `nil` if no children.
    public let firstChildID: String?

    /// Next sibling in display order, or `nil` if last.
    public let nextSiblingID: String?

    /// Previous sibling in display order, or `nil` if first.
    public let prevSiblingID: String?

    /// Route this content node belongs to.
    public let routeID: String?

    /// Human-readable title.
    public let title: String

    /// URL-safe identifier used for public content delivery.
    public let slug: String

    /// Publishing status (e.g. "draft", "published").
    public let status: String

    /// ISO 8601 timestamp when the node was created.
    public let dateCreated: String

    /// ISO 8601 timestamp when the node was last modified.
    public let dateModified: String

    enum CodingKeys: String, CodingKey {
        case contentID = "content_id"
        case datatypeID = "datatype_id"
        case parentID = "parent_id"
        case firstChildID = "first_child_id"
        case nextSiblingID = "next_sibling_id"
        case prevSiblingID = "prev_sibling_id"
        case routeID = "route_id"
        case title
        case slug
        case status
        case dateCreated = "date_created"
        case dateModified = "date_modified"
    }
}
