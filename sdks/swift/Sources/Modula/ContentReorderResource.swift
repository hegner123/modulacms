import Foundation

// MARK: - Content Reorder Types

/// Request body for `POST /api/v1/contentdata/reorder`.
///
/// Atomically rewrites the sibling-pointer linked list for all children of
/// `parentID` so they appear in the order specified by `orderedIDs`. The server
/// updates each node's `next_sibling_id` and `prev_sibling_id` pointers, and sets
/// the parent's `first_child_id` to `orderedIDs[0]`.
public struct ContentReorderRequest: Codable, Sendable {
    /// The parent node whose children are being reordered.
    /// Pass `nil` to reorder root-level (parentless) nodes.
    public let parentID: ContentID?

    /// The desired display order of the sibling nodes.
    /// All IDs must be existing children of `parentID`.
    public let orderedIDs: [ContentID]

    enum CodingKeys: String, CodingKey {
        case parentID = "parent_id"
        case orderedIDs = "ordered_ids"
    }

    public init(parentID: ContentID? = nil, orderedIDs: [ContentID]) {
        self.parentID = parentID
        self.orderedIDs = orderedIDs
    }
}

/// Response from `POST /api/v1/contentdata/reorder`.
public struct ContentReorderResponse: Codable, Sendable {
    /// Number of sibling-pointer fields that were modified.
    public let updated: Int

    /// Echoes back the parent node whose children were reordered.
    /// `nil` when root-level nodes were reordered.
    public let parentID: ContentID?

    enum CodingKeys: String, CodingKey {
        case updated
        case parentID = "parent_id"
    }
}

/// Request body for `POST /api/v1/contentdata/move`.
///
/// Detaches a node from its current position in the sibling linked list,
/// re-links the surrounding siblings, and inserts it under `newParentID` at the
/// given `position`. All sibling pointers and the parent's `first_child_id` are
/// updated atomically.
public struct ContentMoveRequest: Codable, Sendable {
    /// The content node to relocate.
    public let nodeID: ContentID

    /// The destination parent. Pass `nil` to move the node to root level.
    public let newParentID: ContentID?

    /// Zero-based index among the new parent's children where the node
    /// should be inserted. Use 0 to make it the first child.
    public let position: Int

    enum CodingKeys: String, CodingKey {
        case nodeID = "node_id"
        case newParentID = "new_parent_id"
        case position
    }

    public init(nodeID: ContentID, newParentID: ContentID? = nil, position: Int) {
        self.nodeID = nodeID
        self.newParentID = newParentID
        self.position = position
    }
}

/// Response from `POST /api/v1/contentdata/move`.
public struct ContentMoveResponse: Codable, Sendable {
    /// Echoes the moved node's ID.
    public let nodeID: ContentID

    /// The parent the node was detached from. `nil` if previously at root level.
    public let oldParentID: ContentID?

    /// The parent the node was attached to. `nil` if moved to root level.
    public let newParentID: ContentID?

    /// Final zero-based index of the node among its new siblings.
    public let position: Int

    enum CodingKeys: String, CodingKey {
        case nodeID = "node_id"
        case oldParentID = "old_parent_id"
        case newParentID = "new_parent_id"
        case position
    }
}

// MARK: - Admin Content Reorder Types

/// Request body for `POST /api/v1/admincontentdatas/reorder`.
///
/// Admin-content equivalent of ``ContentReorderRequest``, operating on
/// admin content data nodes rather than public content data nodes.
public struct AdminContentReorderRequest: Codable, Sendable {
    /// The admin content parent whose children are being reordered.
    /// Pass `nil` to reorder root-level admin content nodes.
    public let parentID: AdminContentID?

    /// The desired display order of the admin content sibling nodes.
    public let orderedIDs: [AdminContentID]

    enum CodingKeys: String, CodingKey {
        case parentID = "parent_id"
        case orderedIDs = "ordered_ids"
    }

    public init(parentID: AdminContentID? = nil, orderedIDs: [AdminContentID]) {
        self.parentID = parentID
        self.orderedIDs = orderedIDs
    }
}

/// Response from `POST /api/v1/admincontentdatas/reorder`.
public struct AdminContentReorderResponse: Codable, Sendable {
    /// Number of sibling-pointer fields that were modified.
    public let updated: Int

    /// Echoes back the parent node whose children were reordered.
    public let parentID: AdminContentID?

    enum CodingKeys: String, CodingKey {
        case updated
        case parentID = "parent_id"
    }
}

/// Request body for `POST /api/v1/admincontentdatas/move`.
///
/// Admin-content equivalent of ``ContentMoveRequest``, operating on
/// admin content data nodes rather than public content data nodes.
public struct AdminContentMoveRequest: Codable, Sendable {
    /// The admin content node to relocate.
    public let nodeID: AdminContentID

    /// The destination parent. Pass `nil` to move to root level.
    public let newParentID: AdminContentID?

    /// Zero-based index among the new parent's children where the node
    /// should be inserted.
    public let position: Int

    enum CodingKeys: String, CodingKey {
        case nodeID = "node_id"
        case newParentID = "new_parent_id"
        case position
    }

    public init(nodeID: AdminContentID, newParentID: AdminContentID? = nil, position: Int) {
        self.nodeID = nodeID
        self.newParentID = newParentID
        self.position = position
    }
}

/// Response from `POST /api/v1/admincontentdatas/move`.
public struct AdminContentMoveResponse: Codable, Sendable {
    /// Echoes the moved admin content node's ID.
    public let nodeID: AdminContentID

    /// The parent the node was detached from. `nil` if previously at root level.
    public let oldParentID: AdminContentID?

    /// The parent the node was attached to. `nil` if moved to root level.
    public let newParentID: AdminContentID?

    /// Final zero-based index of the node among its new siblings.
    public let position: Int

    enum CodingKeys: String, CodingKey {
        case nodeID = "node_id"
        case oldParentID = "old_parent_id"
        case newParentID = "new_parent_id"
        case position
    }
}

// MARK: - Content Reorder Resource

/// Provides content tree reorder and move operations for public content data nodes.
///
/// These operations manipulate the sibling-pointer linked list (`next_sibling_id`,
/// `prev_sibling_id`) and parent-child pointers (`parent_id`, `first_child_id`)
/// that define the content tree's ordering.
///
/// ```swift
/// let resp = try await client.contentReorder.reorder(ContentReorderRequest(
///     parentID: parentID,
///     orderedIDs: [childA, childB, childC]
/// ))
/// ```
///
/// For admin content data nodes, use ``AdminContentReorderResource`` via
/// `client.adminContentReorder` instead.
public final class ContentReorderResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Atomically rewrites the sibling-pointer linked list so that the children
    /// of the specified parent appear in the given order.
    public func reorder(_ request: ContentReorderRequest) async throws -> ContentReorderResponse {
        try await http.post(path: "/api/v1/contentdata/reorder", body: request)
    }

    /// Detaches a content node from its current position and inserts it under
    /// a new parent at the specified zero-based position.
    public func move(_ request: ContentMoveRequest) async throws -> ContentMoveResponse {
        try await http.post(path: "/api/v1/contentdata/move", body: request)
    }
}

// MARK: - Admin Content Reorder Resource

/// Provides reorder and move operations for admin content data nodes.
///
/// Admin-content counterpart of ``ContentReorderResource``, operating on the
/// admin content tree rather than the public content tree.
///
/// ```swift
/// let resp = try await client.adminContentReorder.reorder(AdminContentReorderRequest(
///     parentID: parentID,
///     orderedIDs: [childA, childB, childC]
/// ))
/// ```
public final class AdminContentReorderResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Atomically rewrites the sibling-pointer linked list so that the admin
    /// content children of the specified parent appear in the given order.
    public func reorder(_ request: AdminContentReorderRequest) async throws -> AdminContentReorderResponse {
        try await http.post(path: "/api/v1/admincontentdatas/reorder", body: request)
    }

    /// Detaches an admin content node from its current position and inserts it
    /// under a new parent at the specified zero-based position.
    public func move(_ request: AdminContentMoveRequest) async throws -> AdminContentMoveResponse {
        try await http.post(path: "/api/v1/admincontentdatas/move", body: request)
    }
}
