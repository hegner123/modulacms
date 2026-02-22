import Foundation

// MARK: - NoCreate

/// Uninhabitable type for Resource CreateParams when creation is not supported
/// (e.g., Media is created via multipart upload, not JSON POST).
/// Using this instead of Never avoids macOS 14+ availability requirements.
public enum NoCreate: Encodable, Sendable {
    public func encode(to encoder: Encoder) throws {
        // Unreachable — NoCreate has no cases
    }
}

// MARK: - Content Data

public struct ContentData: Codable, Sendable {
    public let contentDataID: ContentID
    public let parentID: ContentID?
    public let firstChildID: String?
    public let nextSiblingID: String?
    public let prevSiblingID: String?
    public let routeID: RouteID?
    public let datatypeID: DatatypeID?
    public let authorID: UserID?
    public let status: ContentStatus
    public let dateCreated: Timestamp
    public let dateModified: Timestamp

    enum CodingKeys: String, CodingKey {
        case contentDataID = "content_data_id"
        case parentID = "parent_id"
        case firstChildID = "first_child_id"
        case nextSiblingID = "next_sibling_id"
        case prevSiblingID = "prev_sibling_id"
        case routeID = "route_id"
        case datatypeID = "datatype_id"
        case authorID = "author_id"
        case status
        case dateCreated = "date_created"
        case dateModified = "date_modified"
    }
}

public struct CreateContentDataParams: Encodable, Sendable {
    public let parentID: ContentID?
    public let firstChildID: String?
    public let nextSiblingID: String?
    public let prevSiblingID: String?
    public let routeID: RouteID?
    public let datatypeID: DatatypeID?
    public let authorID: UserID
    public let status: ContentStatus

    public init(
        parentID: ContentID? = nil,
        firstChildID: String? = nil,
        nextSiblingID: String? = nil,
        prevSiblingID: String? = nil,
        routeID: RouteID? = nil,
        datatypeID: DatatypeID? = nil,
        authorID: UserID,
        status: ContentStatus
    ) {
        self.parentID = parentID
        self.firstChildID = firstChildID
        self.nextSiblingID = nextSiblingID
        self.prevSiblingID = prevSiblingID
        self.routeID = routeID
        self.datatypeID = datatypeID
        self.authorID = authorID
        self.status = status
    }

    enum CodingKeys: String, CodingKey {
        case parentID = "parent_id"
        case firstChildID = "first_child_id"
        case nextSiblingID = "next_sibling_id"
        case prevSiblingID = "prev_sibling_id"
        case routeID = "route_id"
        case datatypeID = "datatype_id"
        case authorID = "author_id"
        case status
    }
}

public struct UpdateContentDataParams: Encodable, Sendable {
    public let contentDataID: ContentID
    public let parentID: ContentID?
    public let firstChildID: String?
    public let nextSiblingID: String?
    public let prevSiblingID: String?
    public let routeID: RouteID?
    public let datatypeID: DatatypeID?
    public let authorID: UserID
    public let status: ContentStatus

    public init(
        contentDataID: ContentID,
        parentID: ContentID? = nil,
        firstChildID: String? = nil,
        nextSiblingID: String? = nil,
        prevSiblingID: String? = nil,
        routeID: RouteID? = nil,
        datatypeID: DatatypeID? = nil,
        authorID: UserID,
        status: ContentStatus
    ) {
        self.contentDataID = contentDataID
        self.parentID = parentID
        self.firstChildID = firstChildID
        self.nextSiblingID = nextSiblingID
        self.prevSiblingID = prevSiblingID
        self.routeID = routeID
        self.datatypeID = datatypeID
        self.authorID = authorID
        self.status = status
    }

    enum CodingKeys: String, CodingKey {
        case contentDataID = "content_data_id"
        case parentID = "parent_id"
        case firstChildID = "first_child_id"
        case nextSiblingID = "next_sibling_id"
        case prevSiblingID = "prev_sibling_id"
        case routeID = "route_id"
        case datatypeID = "datatype_id"
        case authorID = "author_id"
        case status
    }
}

// MARK: - Content Field

public struct ContentField: Codable, Sendable {
    public let contentFieldID: ContentFieldID
    public let routeID: RouteID?
    public let contentDataID: ContentID?
    public let fieldID: FieldID?
    public let fieldValue: String
    public let authorID: UserID?
    public let dateCreated: Timestamp
    public let dateModified: Timestamp

    enum CodingKeys: String, CodingKey {
        case contentFieldID = "content_field_id"
        case routeID = "route_id"
        case contentDataID = "content_data_id"
        case fieldID = "field_id"
        case fieldValue = "field_value"
        case authorID = "author_id"
        case dateCreated = "date_created"
        case dateModified = "date_modified"
    }
}

public struct CreateContentFieldParams: Encodable, Sendable {
    public let routeID: RouteID?
    public let contentDataID: ContentID?
    public let fieldID: FieldID?
    public let fieldValue: String
    public let authorID: UserID

    public init(
        routeID: RouteID? = nil,
        contentDataID: ContentID? = nil,
        fieldID: FieldID? = nil,
        fieldValue: String,
        authorID: UserID
    ) {
        self.routeID = routeID
        self.contentDataID = contentDataID
        self.fieldID = fieldID
        self.fieldValue = fieldValue
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case routeID = "route_id"
        case contentDataID = "content_data_id"
        case fieldID = "field_id"
        case fieldValue = "field_value"
        case authorID = "author_id"
    }
}

public struct UpdateContentFieldParams: Encodable, Sendable {
    public let contentFieldID: ContentFieldID
    public let routeID: RouteID?
    public let contentDataID: ContentID?
    public let fieldID: FieldID?
    public let fieldValue: String
    public let authorID: UserID

    public init(
        contentFieldID: ContentFieldID,
        routeID: RouteID? = nil,
        contentDataID: ContentID? = nil,
        fieldID: FieldID? = nil,
        fieldValue: String,
        authorID: UserID
    ) {
        self.contentFieldID = contentFieldID
        self.routeID = routeID
        self.contentDataID = contentDataID
        self.fieldID = fieldID
        self.fieldValue = fieldValue
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case contentFieldID = "content_field_id"
        case routeID = "route_id"
        case contentDataID = "content_data_id"
        case fieldID = "field_id"
        case fieldValue = "field_value"
        case authorID = "author_id"
    }
}

// MARK: - Content Relation

public struct ContentRelation: Codable, Sendable {
    public let contentRelationID: ContentRelationID
    public let sourceContentID: ContentID
    public let targetContentID: ContentID
    public let fieldID: FieldID
    public let sortOrder: Int64
    public let dateCreated: Timestamp

    enum CodingKeys: String, CodingKey {
        case contentRelationID = "content_relation_id"
        case sourceContentID = "source_content_id"
        case targetContentID = "target_content_id"
        case fieldID = "field_id"
        case sortOrder = "sort_order"
        case dateCreated = "date_created"
    }
}

public struct CreateContentRelationParams: Encodable, Sendable {
    public let sourceContentID: ContentID
    public let targetContentID: ContentID
    public let fieldID: FieldID
    public let sortOrder: Int64

    public init(
        sourceContentID: ContentID,
        targetContentID: ContentID,
        fieldID: FieldID,
        sortOrder: Int64
    ) {
        self.sourceContentID = sourceContentID
        self.targetContentID = targetContentID
        self.fieldID = fieldID
        self.sortOrder = sortOrder
    }

    enum CodingKeys: String, CodingKey {
        case sourceContentID = "source_content_id"
        case targetContentID = "target_content_id"
        case fieldID = "field_id"
        case sortOrder = "sort_order"
    }
}

public struct UpdateContentRelationParams: Encodable, Sendable {
    public let contentRelationID: ContentRelationID
    public let sortOrder: Int64

    public init(contentRelationID: ContentRelationID, sortOrder: Int64) {
        self.contentRelationID = contentRelationID
        self.sortOrder = sortOrder
    }

    enum CodingKeys: String, CodingKey {
        case contentRelationID = "content_relation_id"
        case sortOrder = "sort_order"
    }
}

// MARK: - Datatype

public struct Datatype: Codable, Sendable {
    public let datatypeID: DatatypeID
    public let parentID: DatatypeID?
    public let label: String
    public let type: String
    public let authorID: UserID?
    public let dateCreated: Timestamp
    public let dateModified: Timestamp

    enum CodingKeys: String, CodingKey {
        case datatypeID = "datatype_id"
        case parentID = "parent_id"
        case label
        case type
        case authorID = "author_id"
        case dateCreated = "date_created"
        case dateModified = "date_modified"
    }
}

public struct CreateDatatypeParams: Encodable, Sendable {
    public let datatypeID: DatatypeID?
    public let parentID: DatatypeID?
    public let label: String
    public let type: String
    public let authorID: UserID

    public init(
        datatypeID: DatatypeID? = nil,
        parentID: DatatypeID? = nil,
        label: String,
        type: String,
        authorID: UserID
    ) {
        self.datatypeID = datatypeID
        self.parentID = parentID
        self.label = label
        self.type = type
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case datatypeID = "datatype_id"
        case parentID = "parent_id"
        case label
        case type
        case authorID = "author_id"
    }
}

public struct UpdateDatatypeParams: Encodable, Sendable {
    public let datatypeID: DatatypeID
    public let parentID: DatatypeID?
    public let label: String
    public let type: String
    public let authorID: UserID

    public init(
        datatypeID: DatatypeID,
        parentID: DatatypeID? = nil,
        label: String,
        type: String,
        authorID: UserID
    ) {
        self.datatypeID = datatypeID
        self.parentID = parentID
        self.label = label
        self.type = type
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case datatypeID = "datatype_id"
        case parentID = "parent_id"
        case label
        case type
        case authorID = "author_id"
    }
}

// MARK: - Datatype Field

public struct DatatypeField: Codable, Sendable {
    public let id: DatatypeFieldID
    public let datatypeID: DatatypeID
    public let fieldID: FieldID
    public let sortOrder: Int64

    enum CodingKeys: String, CodingKey {
        case id
        case datatypeID = "datatype_id"
        case fieldID = "field_id"
        case sortOrder = "sort_order"
    }
}

public struct CreateDatatypeFieldParams: Encodable, Sendable {
    public let datatypeID: DatatypeID
    public let fieldID: FieldID
    public let sortOrder: Int64

    public init(datatypeID: DatatypeID, fieldID: FieldID, sortOrder: Int64) {
        self.datatypeID = datatypeID
        self.fieldID = fieldID
        self.sortOrder = sortOrder
    }

    enum CodingKeys: String, CodingKey {
        case datatypeID = "datatype_id"
        case fieldID = "field_id"
        case sortOrder = "sort_order"
    }
}

public struct UpdateDatatypeFieldParams: Encodable, Sendable {
    public let id: DatatypeFieldID
    public let datatypeID: DatatypeID
    public let fieldID: FieldID
    public let sortOrder: Int64

    public init(id: DatatypeFieldID, datatypeID: DatatypeID, fieldID: FieldID, sortOrder: Int64) {
        self.id = id
        self.datatypeID = datatypeID
        self.fieldID = fieldID
        self.sortOrder = sortOrder
    }

    enum CodingKeys: String, CodingKey {
        case id
        case datatypeID = "datatype_id"
        case fieldID = "field_id"
        case sortOrder = "sort_order"
    }
}

// MARK: - Field

public struct Field: Codable, Sendable {
    public let fieldID: FieldID
    public let parentID: DatatypeID?
    public let label: String
    public let data: String
    public let validation: String
    public let uiConfig: String
    public let type: FieldType
    public let authorID: UserID?
    public let dateCreated: Timestamp
    public let dateModified: Timestamp

    enum CodingKeys: String, CodingKey {
        case fieldID = "field_id"
        case parentID = "parent_id"
        case label
        case data
        case validation
        case uiConfig = "ui_config"
        case type
        case authorID = "author_id"
        case dateCreated = "date_created"
        case dateModified = "date_modified"
    }
}

public struct CreateFieldParams: Encodable, Sendable {
    public let fieldID: FieldID?
    public let parentID: DatatypeID?
    public let label: String
    public let data: String
    public let validation: String
    public let uiConfig: String
    public let type: FieldType
    public let authorID: UserID

    public init(
        fieldID: FieldID? = nil,
        parentID: DatatypeID? = nil,
        label: String,
        data: String,
        validation: String,
        uiConfig: String,
        type: FieldType,
        authorID: UserID
    ) {
        self.fieldID = fieldID
        self.parentID = parentID
        self.label = label
        self.data = data
        self.validation = validation
        self.uiConfig = uiConfig
        self.type = type
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case fieldID = "field_id"
        case parentID = "parent_id"
        case label
        case data
        case validation
        case uiConfig = "ui_config"
        case type
        case authorID = "author_id"
    }
}

public struct UpdateFieldParams: Encodable, Sendable {
    public let fieldID: FieldID
    public let parentID: DatatypeID?
    public let label: String
    public let data: String
    public let validation: String
    public let uiConfig: String
    public let type: FieldType
    public let authorID: UserID

    public init(
        fieldID: FieldID,
        parentID: DatatypeID? = nil,
        label: String,
        data: String,
        validation: String,
        uiConfig: String,
        type: FieldType,
        authorID: UserID
    ) {
        self.fieldID = fieldID
        self.parentID = parentID
        self.label = label
        self.data = data
        self.validation = validation
        self.uiConfig = uiConfig
        self.type = type
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case fieldID = "field_id"
        case parentID = "parent_id"
        case label
        case data
        case validation
        case uiConfig = "ui_config"
        case type
        case authorID = "author_id"
    }
}

// MARK: - Media

public struct Media: Codable, Sendable {
    public let mediaID: MediaID
    public let name: String?
    public let displayName: String?
    public let alt: String?
    public let caption: String?
    public let description: String?
    public let mediaClass: String?
    public let mimetype: String?
    public let dimensions: String?
    public let url: URLValue
    public let srcset: String?
    public let focalX: Double?
    public let focalY: Double?
    public let authorID: UserID?
    public let dateCreated: Timestamp
    public let dateModified: Timestamp

    enum CodingKeys: String, CodingKey {
        case mediaID = "media_id"
        case name
        case displayName = "display_name"
        case alt
        case caption
        case description
        case mediaClass = "class"
        case mimetype
        case dimensions
        case url
        case srcset
        case focalX = "focal_x"
        case focalY = "focal_y"
        case authorID = "author_id"
        case dateCreated = "date_created"
        case dateModified = "date_modified"
    }
}

public struct UpdateMediaParams: Encodable, Sendable {
    public let mediaID: MediaID
    public let name: String?
    public let displayName: String?
    public let alt: String?
    public let caption: String?
    public let description: String?
    public let mediaClass: String?
    public let focalX: Double?
    public let focalY: Double?

    public init(
        mediaID: MediaID,
        name: String? = nil,
        displayName: String? = nil,
        alt: String? = nil,
        caption: String? = nil,
        description: String? = nil,
        mediaClass: String? = nil,
        focalX: Double? = nil,
        focalY: Double? = nil
    ) {
        self.mediaID = mediaID
        self.name = name
        self.displayName = displayName
        self.alt = alt
        self.caption = caption
        self.description = description
        self.mediaClass = mediaClass
        self.focalX = focalX
        self.focalY = focalY
    }

    enum CodingKeys: String, CodingKey {
        case mediaID = "media_id"
        case name
        case displayName = "display_name"
        case alt
        case caption
        case description
        case mediaClass = "class"
        case focalX = "focal_x"
        case focalY = "focal_y"
    }
}

// MARK: - Media Dimension

public struct MediaDimension: Codable, Sendable {
    public let mdID: MediaDimensionID
    public let label: String?
    public let width: Int64?
    public let height: Int64?
    public let aspectRatio: String?

    enum CodingKeys: String, CodingKey {
        case mdID = "md_id"
        case label
        case width
        case height
        case aspectRatio = "aspect_ratio"
    }
}

public struct CreateMediaDimensionParams: Encodable, Sendable {
    public let label: String?
    public let width: Int64?
    public let height: Int64?
    public let aspectRatio: String?

    public init(
        label: String? = nil,
        width: Int64? = nil,
        height: Int64? = nil,
        aspectRatio: String? = nil
    ) {
        self.label = label
        self.width = width
        self.height = height
        self.aspectRatio = aspectRatio
    }

    enum CodingKeys: String, CodingKey {
        case label
        case width
        case height
        case aspectRatio = "aspect_ratio"
    }
}

public struct UpdateMediaDimensionParams: Encodable, Sendable {
    public let mdID: MediaDimensionID
    public let label: String?
    public let width: Int64?
    public let height: Int64?
    public let aspectRatio: String?

    public init(
        mdID: MediaDimensionID,
        label: String? = nil,
        width: Int64? = nil,
        height: Int64? = nil,
        aspectRatio: String? = nil
    ) {
        self.mdID = mdID
        self.label = label
        self.width = width
        self.height = height
        self.aspectRatio = aspectRatio
    }

    enum CodingKeys: String, CodingKey {
        case mdID = "md_id"
        case label
        case width
        case height
        case aspectRatio = "aspect_ratio"
    }
}

// MARK: - Route

public struct Route: Codable, Sendable {
    public let routeID: RouteID
    public let slug: Slug
    public let title: String
    public let status: Int64
    public let authorID: UserID?
    public let dateCreated: Timestamp
    public let dateModified: Timestamp

    enum CodingKeys: String, CodingKey {
        case routeID = "route_id"
        case slug
        case title
        case status
        case authorID = "author_id"
        case dateCreated = "date_created"
        case dateModified = "date_modified"
    }
}

public struct CreateRouteParams: Encodable, Sendable {
    public let slug: Slug
    public let title: String
    public let status: Int64
    public let authorID: UserID

    public init(
        slug: Slug,
        title: String,
        status: Int64,
        authorID: UserID
    ) {
        self.slug = slug
        self.title = title
        self.status = status
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case slug
        case title
        case status
        case authorID = "author_id"
    }
}

public struct UpdateRouteParams: Encodable, Sendable {
    public let slug: Slug
    public let title: String
    public let status: Int64
    public let authorID: UserID
    public let slug2: Slug

    public init(
        slug: Slug,
        title: String,
        status: Int64,
        authorID: UserID,
        slug2: Slug
    ) {
        self.slug = slug
        self.title = title
        self.status = status
        self.authorID = authorID
        self.slug2 = slug2
    }

    enum CodingKeys: String, CodingKey {
        case slug
        case title
        case status
        case authorID = "author_id"
        case slug2 = "slug_2"
    }
}

// MARK: - User

public struct User: Codable, Sendable {
    public let userID: UserID
    public let username: String
    public let name: String
    public let email: Email
    public let role: String
    public let dateCreated: Timestamp
    public let dateModified: Timestamp

    enum CodingKeys: String, CodingKey {
        case userID = "user_id"
        case username
        case name
        case email
        case role
        case dateCreated = "date_created"
        case dateModified = "date_modified"
    }
}

public struct CreateUserParams: Encodable, Sendable {
    public let username: String
    public let name: String
    public let email: Email
    public let password: String
    public let role: String

    public init(username: String, name: String, email: Email, password: String, role: String) {
        self.username = username
        self.name = name
        self.email = email
        self.password = password
        self.role = role
    }

    enum CodingKeys: String, CodingKey {
        case username
        case name
        case email
        case password
        case role
    }
}

public struct UpdateUserParams: Encodable, Sendable {
    public let userID: UserID
    public let username: String
    public let name: String
    public let email: Email
    public let password: String?
    public let role: String

    public init(
        userID: UserID,
        username: String,
        name: String,
        email: Email,
        password: String? = nil,
        role: String
    ) {
        self.userID = userID
        self.username = username
        self.name = name
        self.email = email
        self.password = password
        self.role = role
    }

    enum CodingKeys: String, CodingKey {
        case userID = "user_id"
        case username
        case name
        case email
        case password
        case role
    }
}

public struct ResetPasswordParams: Encodable, Sendable {
    public let email: Email
    public let newPassword: String
    public let token: String

    public init(email: Email, newPassword: String, token: String) {
        self.email = email
        self.newPassword = newPassword
        self.token = token
    }

    enum CodingKeys: String, CodingKey {
        case email
        case newPassword = "new_password"
        case token
    }
}

// MARK: - Role

public struct Role: Codable, Sendable {
    public let roleID: RoleID
    public let label: String

    enum CodingKeys: String, CodingKey {
        case roleID = "role_id"
        case label
    }
}

public struct CreateRoleParams: Encodable, Sendable {
    public let label: String

    public init(label: String) {
        self.label = label
    }

    enum CodingKeys: String, CodingKey {
        case label
    }
}

public struct UpdateRoleParams: Encodable, Sendable {
    public let roleID: RoleID
    public let label: String

    public init(roleID: RoleID, label: String) {
        self.roleID = roleID
        self.label = label
    }

    enum CodingKeys: String, CodingKey {
        case roleID = "role_id"
        case label
    }
}

// MARK: - Permission

public struct Permission: Codable, Sendable {
    public let permissionID: PermissionID
    public let label: String

    enum CodingKeys: String, CodingKey {
        case permissionID = "permission_id"
        case label
    }
}

public struct CreatePermissionParams: Encodable, Sendable {
    public let label: String

    public init(label: String) {
        self.label = label
    }

    enum CodingKeys: String, CodingKey {
        case label
    }
}

public struct UpdatePermissionParams: Encodable, Sendable {
    public let permissionID: PermissionID
    public let label: String

    public init(permissionID: PermissionID, label: String) {
        self.permissionID = permissionID
        self.label = label
    }

    enum CodingKeys: String, CodingKey {
        case permissionID = "permission_id"
        case label
    }
}

// MARK: - RolePermission

public struct RolePermission: Codable, Sendable {
    public let id: RolePermissionID
    public let roleID: RoleID
    public let permissionID: PermissionID

    enum CodingKeys: String, CodingKey {
        case id
        case roleID = "role_id"
        case permissionID = "permission_id"
    }
}

public struct CreateRolePermissionParams: Encodable, Sendable {
    public let roleID: RoleID
    public let permissionID: PermissionID

    public init(roleID: RoleID, permissionID: PermissionID) {
        self.roleID = roleID
        self.permissionID = permissionID
    }

    enum CodingKeys: String, CodingKey {
        case roleID = "role_id"
        case permissionID = "permission_id"
    }
}

// MARK: - Session

public struct Session: Codable, Sendable {
    public let sessionID: SessionID
    public let userID: UserID?
    public let dateCreated: Timestamp
    public let expiresAt: Timestamp
    public let lastAccess: String?
    public let ipAddress: String?
    public let userAgent: String?
    public let sessionData: String?

    enum CodingKeys: String, CodingKey {
        case sessionID = "session_id"
        case userID = "user_id"
        case dateCreated = "date_created"
        case expiresAt = "expires_at"
        case lastAccess = "last_access"
        case ipAddress = "ip_address"
        case userAgent = "user_agent"
        case sessionData = "session_data"
    }
}

public struct UpdateSessionParams: Encodable, Sendable {
    public let sessionID: SessionID
    public let userID: UserID?
    public let expiresAt: Timestamp
    public let lastAccess: String?
    public let ipAddress: String?
    public let userAgent: String?
    public let sessionData: String?

    public init(
        sessionID: SessionID,
        userID: UserID? = nil,
        expiresAt: Timestamp,
        lastAccess: String? = nil,
        ipAddress: String? = nil,
        userAgent: String? = nil,
        sessionData: String? = nil
    ) {
        self.sessionID = sessionID
        self.userID = userID
        self.expiresAt = expiresAt
        self.lastAccess = lastAccess
        self.ipAddress = ipAddress
        self.userAgent = userAgent
        self.sessionData = sessionData
    }

    enum CodingKeys: String, CodingKey {
        case sessionID = "session_id"
        case userID = "user_id"
        case expiresAt = "expires_at"
        case lastAccess = "last_access"
        case ipAddress = "ip_address"
        case userAgent = "user_agent"
        case sessionData = "session_data"
    }
}

// MARK: - Token

public struct Token: Codable, Sendable {
    public let id: TokenID
    public let userID: UserID?
    public let tokenType: String
    public let token: String
    public let issuedAt: String
    public let expiresAt: Timestamp
    public let revoked: Bool

    enum CodingKeys: String, CodingKey {
        case id
        case userID = "user_id"
        case tokenType = "token_type"
        case token
        case issuedAt = "issued_at"
        case expiresAt = "expires_at"
        case revoked
    }
}

public struct CreateTokenParams: Encodable, Sendable {
    public let userID: UserID?
    public let tokenType: String
    public let token: String
    public let issuedAt: String
    public let expiresAt: Timestamp
    public let revoked: Bool

    public init(
        userID: UserID? = nil,
        tokenType: String,
        token: String,
        issuedAt: String,
        expiresAt: Timestamp,
        revoked: Bool
    ) {
        self.userID = userID
        self.tokenType = tokenType
        self.token = token
        self.issuedAt = issuedAt
        self.expiresAt = expiresAt
        self.revoked = revoked
    }

    enum CodingKeys: String, CodingKey {
        case userID = "user_id"
        case tokenType = "token_type"
        case token
        case issuedAt = "issued_at"
        case expiresAt = "expires_at"
        case revoked
    }
}

public struct UpdateTokenParams: Encodable, Sendable {
    public let id: TokenID
    public let token: String
    public let issuedAt: String
    public let expiresAt: Timestamp
    public let revoked: Bool

    public init(
        id: TokenID,
        token: String,
        issuedAt: String,
        expiresAt: Timestamp,
        revoked: Bool
    ) {
        self.id = id
        self.token = token
        self.issuedAt = issuedAt
        self.expiresAt = expiresAt
        self.revoked = revoked
    }

    enum CodingKeys: String, CodingKey {
        case id
        case token
        case issuedAt = "issued_at"
        case expiresAt = "expires_at"
        case revoked
    }
}

// MARK: - User OAuth

public struct UserOauth: Codable, Sendable {
    public let userOauthID: UserOauthID
    public let userID: UserID?
    public let oauthProvider: String
    public let oauthProviderUserID: String
    public let accessToken: String
    public let refreshToken: String
    public let tokenExpiresAt: String
    public let dateCreated: Timestamp

    enum CodingKeys: String, CodingKey {
        case userOauthID = "user_oauth_id"
        case userID = "user_id"
        case oauthProvider = "oauth_provider"
        case oauthProviderUserID = "oauth_provider_user_id"
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
        case tokenExpiresAt = "token_expires_at"
        case dateCreated = "date_created"
    }
}

public struct CreateUserOauthParams: Encodable, Sendable {
    public let userID: UserID?
    public let oauthProvider: String
    public let oauthProviderUserID: String
    public let accessToken: String
    public let refreshToken: String
    public let tokenExpiresAt: String
    public let dateCreated: Timestamp

    public init(
        userID: UserID? = nil,
        oauthProvider: String,
        oauthProviderUserID: String,
        accessToken: String,
        refreshToken: String,
        tokenExpiresAt: String,
        dateCreated: Timestamp
    ) {
        self.userID = userID
        self.oauthProvider = oauthProvider
        self.oauthProviderUserID = oauthProviderUserID
        self.accessToken = accessToken
        self.refreshToken = refreshToken
        self.tokenExpiresAt = tokenExpiresAt
        self.dateCreated = dateCreated
    }

    enum CodingKeys: String, CodingKey {
        case userID = "user_id"
        case oauthProvider = "oauth_provider"
        case oauthProviderUserID = "oauth_provider_user_id"
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
        case tokenExpiresAt = "token_expires_at"
        case dateCreated = "date_created"
    }
}

public struct UpdateUserOauthParams: Encodable, Sendable {
    public let userOauthID: UserOauthID
    public let accessToken: String
    public let refreshToken: String
    public let tokenExpiresAt: String

    public init(
        userOauthID: UserOauthID,
        accessToken: String,
        refreshToken: String,
        tokenExpiresAt: String
    ) {
        self.userOauthID = userOauthID
        self.accessToken = accessToken
        self.refreshToken = refreshToken
        self.tokenExpiresAt = tokenExpiresAt
    }

    enum CodingKeys: String, CodingKey {
        case userOauthID = "user_oauth_id"
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
        case tokenExpiresAt = "token_expires_at"
    }
}

// MARK: - User SSH Key

public struct SshKey: Codable, Sendable {
    public let sshKeyID: UserSshKeyID
    public let userID: UserID?
    public let publicKey: String
    public let keyType: String
    public let fingerprint: String
    public let label: String
    public let dateCreated: Timestamp
    public let lastUsed: String

    enum CodingKeys: String, CodingKey {
        case sshKeyID = "ssh_key_id"
        case userID = "user_id"
        case publicKey = "public_key"
        case keyType = "key_type"
        case fingerprint
        case label
        case dateCreated = "date_created"
        case lastUsed = "last_used"
    }
}

public struct SshKeyListItem: Codable, Sendable {
    public let sshKeyID: UserSshKeyID
    public let keyType: String
    public let fingerprint: String
    public let label: String
    public let dateCreated: Timestamp
    public let lastUsed: String

    enum CodingKeys: String, CodingKey {
        case sshKeyID = "ssh_key_id"
        case keyType = "key_type"
        case fingerprint
        case label
        case dateCreated = "date_created"
        case lastUsed = "last_used"
    }
}

public struct CreateSSHKeyParams: Encodable, Sendable {
    public let publicKey: String
    public let label: String

    public init(publicKey: String, label: String) {
        self.publicKey = publicKey
        self.label = label
    }

    enum CodingKeys: String, CodingKey {
        case publicKey = "public_key"
        case label
    }
}

// MARK: - Table

public struct Table: Codable, Sendable {
    public let id: TableID
    public let label: String
    public let authorID: UserID?

    enum CodingKeys: String, CodingKey {
        case id
        case label
        case authorID = "author_id"
    }
}

public struct CreateTableParams: Encodable, Sendable {
    public let label: String

    public init(label: String) {
        self.label = label
    }

    enum CodingKeys: String, CodingKey {
        case label
    }
}

public struct UpdateTableParams: Encodable, Sendable {
    public let id: TableID
    public let label: String

    public init(id: TableID, label: String) {
        self.id = id
        self.label = label
    }

    enum CodingKeys: String, CodingKey {
        case id
        case label
    }
}

// MARK: - Admin Content Data

public struct AdminContentData: Codable, Sendable {
    public let adminContentDataID: AdminContentID
    public let parentID: AdminContentID?
    public let firstChildID: String?
    public let nextSiblingID: String?
    public let prevSiblingID: String?
    public let adminRouteID: AdminRouteID?
    public let adminDatatypeID: AdminDatatypeID?
    public let authorID: UserID?
    public let status: ContentStatus
    public let dateCreated: Timestamp
    public let dateModified: Timestamp

    enum CodingKeys: String, CodingKey {
        case adminContentDataID = "admin_content_data_id"
        case parentID = "parent_id"
        case firstChildID = "first_child_id"
        case nextSiblingID = "next_sibling_id"
        case prevSiblingID = "prev_sibling_id"
        case adminRouteID = "admin_route_id"
        case adminDatatypeID = "admin_datatype_id"
        case authorID = "author_id"
        case status
        case dateCreated = "date_created"
        case dateModified = "date_modified"
    }
}

public struct CreateAdminContentDataParams: Encodable, Sendable {
    public let parentID: AdminContentID?
    public let firstChildID: String?
    public let nextSiblingID: String?
    public let prevSiblingID: String?
    public let adminRouteID: AdminRouteID?
    public let adminDatatypeID: AdminDatatypeID?
    public let authorID: UserID
    public let status: ContentStatus

    public init(
        parentID: AdminContentID? = nil,
        firstChildID: String? = nil,
        nextSiblingID: String? = nil,
        prevSiblingID: String? = nil,
        adminRouteID: AdminRouteID? = nil,
        adminDatatypeID: AdminDatatypeID? = nil,
        authorID: UserID,
        status: ContentStatus
    ) {
        self.parentID = parentID
        self.firstChildID = firstChildID
        self.nextSiblingID = nextSiblingID
        self.prevSiblingID = prevSiblingID
        self.adminRouteID = adminRouteID
        self.adminDatatypeID = adminDatatypeID
        self.authorID = authorID
        self.status = status
    }

    enum CodingKeys: String, CodingKey {
        case parentID = "parent_id"
        case firstChildID = "first_child_id"
        case nextSiblingID = "next_sibling_id"
        case prevSiblingID = "prev_sibling_id"
        case adminRouteID = "admin_route_id"
        case adminDatatypeID = "admin_datatype_id"
        case authorID = "author_id"
        case status
    }
}

public struct UpdateAdminContentDataParams: Encodable, Sendable {
    public let adminContentDataID: AdminContentID
    public let parentID: AdminContentID?
    public let firstChildID: String?
    public let nextSiblingID: String?
    public let prevSiblingID: String?
    public let adminRouteID: AdminRouteID?
    public let adminDatatypeID: AdminDatatypeID?
    public let authorID: UserID
    public let status: ContentStatus

    public init(
        adminContentDataID: AdminContentID,
        parentID: AdminContentID? = nil,
        firstChildID: String? = nil,
        nextSiblingID: String? = nil,
        prevSiblingID: String? = nil,
        adminRouteID: AdminRouteID? = nil,
        adminDatatypeID: AdminDatatypeID? = nil,
        authorID: UserID,
        status: ContentStatus
    ) {
        self.adminContentDataID = adminContentDataID
        self.parentID = parentID
        self.firstChildID = firstChildID
        self.nextSiblingID = nextSiblingID
        self.prevSiblingID = prevSiblingID
        self.adminRouteID = adminRouteID
        self.adminDatatypeID = adminDatatypeID
        self.authorID = authorID
        self.status = status
    }

    enum CodingKeys: String, CodingKey {
        case adminContentDataID = "admin_content_data_id"
        case parentID = "parent_id"
        case firstChildID = "first_child_id"
        case nextSiblingID = "next_sibling_id"
        case prevSiblingID = "prev_sibling_id"
        case adminRouteID = "admin_route_id"
        case adminDatatypeID = "admin_datatype_id"
        case authorID = "author_id"
        case status
    }
}

// MARK: - Admin Content Field

public struct AdminContentField: Codable, Sendable {
    public let adminContentFieldID: AdminContentFieldID
    public let adminRouteID: AdminRouteID?
    public let adminContentDataID: AdminContentID?
    public let adminFieldID: AdminFieldID?
    public let adminFieldValue: String
    public let authorID: UserID?
    public let dateCreated: Timestamp
    public let dateModified: Timestamp

    enum CodingKeys: String, CodingKey {
        case adminContentFieldID = "admin_content_field_id"
        case adminRouteID = "admin_route_id"
        case adminContentDataID = "admin_content_data_id"
        case adminFieldID = "admin_field_id"
        case adminFieldValue = "admin_field_value"
        case authorID = "author_id"
        case dateCreated = "date_created"
        case dateModified = "date_modified"
    }
}

public struct CreateAdminContentFieldParams: Encodable, Sendable {
    public let adminRouteID: AdminRouteID?
    public let adminContentDataID: AdminContentID?
    public let adminFieldID: AdminFieldID?
    public let adminFieldValue: String
    public let authorID: UserID

    public init(
        adminRouteID: AdminRouteID? = nil,
        adminContentDataID: AdminContentID? = nil,
        adminFieldID: AdminFieldID? = nil,
        adminFieldValue: String,
        authorID: UserID
    ) {
        self.adminRouteID = adminRouteID
        self.adminContentDataID = adminContentDataID
        self.adminFieldID = adminFieldID
        self.adminFieldValue = adminFieldValue
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case adminRouteID = "admin_route_id"
        case adminContentDataID = "admin_content_data_id"
        case adminFieldID = "admin_field_id"
        case adminFieldValue = "admin_field_value"
        case authorID = "author_id"
    }
}

public struct UpdateAdminContentFieldParams: Encodable, Sendable {
    public let adminContentFieldID: AdminContentFieldID
    public let adminRouteID: AdminRouteID?
    public let adminContentDataID: AdminContentID?
    public let adminFieldID: AdminFieldID?
    public let adminFieldValue: String
    public let authorID: UserID

    public init(
        adminContentFieldID: AdminContentFieldID,
        adminRouteID: AdminRouteID? = nil,
        adminContentDataID: AdminContentID? = nil,
        adminFieldID: AdminFieldID? = nil,
        adminFieldValue: String,
        authorID: UserID
    ) {
        self.adminContentFieldID = adminContentFieldID
        self.adminRouteID = adminRouteID
        self.adminContentDataID = adminContentDataID
        self.adminFieldID = adminFieldID
        self.adminFieldValue = adminFieldValue
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case adminContentFieldID = "admin_content_field_id"
        case adminRouteID = "admin_route_id"
        case adminContentDataID = "admin_content_data_id"
        case adminFieldID = "admin_field_id"
        case adminFieldValue = "admin_field_value"
        case authorID = "author_id"
    }
}

// MARK: - Admin Content Relation (read-only)

public struct AdminContentRelation: Codable, Sendable {
    public let adminContentRelationID: AdminContentRelationID
    public let sourceContentID: AdminContentID
    public let targetContentID: AdminContentID
    public let adminFieldID: AdminFieldID
    public let sortOrder: Int64
    public let dateCreated: Timestamp

    enum CodingKeys: String, CodingKey {
        case adminContentRelationID = "admin_content_relation_id"
        case sourceContentID = "source_content_id"
        case targetContentID = "target_content_id"
        case adminFieldID = "admin_field_id"
        case sortOrder = "sort_order"
        case dateCreated = "date_created"
    }
}

// MARK: - Admin Datatype

public struct AdminDatatype: Codable, Sendable {
    public let adminDatatypeID: AdminDatatypeID
    public let parentID: AdminDatatypeID?
    public let label: String
    public let type: String
    public let authorID: UserID?
    public let dateCreated: Timestamp
    public let dateModified: Timestamp

    enum CodingKeys: String, CodingKey {
        case adminDatatypeID = "admin_datatype_id"
        case parentID = "parent_id"
        case label
        case type
        case authorID = "author_id"
        case dateCreated = "date_created"
        case dateModified = "date_modified"
    }
}

public struct CreateAdminDatatypeParams: Encodable, Sendable {
    public let parentID: AdminDatatypeID?
    public let label: String
    public let type: String
    public let authorID: UserID

    public init(
        parentID: AdminDatatypeID? = nil,
        label: String,
        type: String,
        authorID: UserID
    ) {
        self.parentID = parentID
        self.label = label
        self.type = type
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case parentID = "parent_id"
        case label
        case type
        case authorID = "author_id"
    }
}

public struct UpdateAdminDatatypeParams: Encodable, Sendable {
    public let adminDatatypeID: AdminDatatypeID
    public let parentID: AdminDatatypeID?
    public let label: String
    public let type: String
    public let authorID: UserID

    public init(
        adminDatatypeID: AdminDatatypeID,
        parentID: AdminDatatypeID? = nil,
        label: String,
        type: String,
        authorID: UserID
    ) {
        self.adminDatatypeID = adminDatatypeID
        self.parentID = parentID
        self.label = label
        self.type = type
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case adminDatatypeID = "admin_datatype_id"
        case parentID = "parent_id"
        case label
        case type
        case authorID = "author_id"
    }
}

// MARK: - Admin Datatype Field

public struct AdminDatatypeField: Codable, Sendable {
    public let id: AdminDatatypeFieldID
    public let adminDatatypeID: AdminDatatypeID
    public let adminFieldID: AdminFieldID

    enum CodingKeys: String, CodingKey {
        case id
        case adminDatatypeID = "admin_datatype_id"
        case adminFieldID = "admin_field_id"
    }
}

public struct CreateAdminDatatypeFieldParams: Encodable, Sendable {
    public let adminDatatypeID: AdminDatatypeID
    public let adminFieldID: AdminFieldID

    public init(adminDatatypeID: AdminDatatypeID, adminFieldID: AdminFieldID) {
        self.adminDatatypeID = adminDatatypeID
        self.adminFieldID = adminFieldID
    }

    enum CodingKeys: String, CodingKey {
        case adminDatatypeID = "admin_datatype_id"
        case adminFieldID = "admin_field_id"
    }
}

public struct UpdateAdminDatatypeFieldParams: Encodable, Sendable {
    public let id: AdminDatatypeFieldID
    public let adminDatatypeID: AdminDatatypeID
    public let adminFieldID: AdminFieldID

    public init(id: AdminDatatypeFieldID, adminDatatypeID: AdminDatatypeID, adminFieldID: AdminFieldID) {
        self.id = id
        self.adminDatatypeID = adminDatatypeID
        self.adminFieldID = adminFieldID
    }

    enum CodingKeys: String, CodingKey {
        case id
        case adminDatatypeID = "admin_datatype_id"
        case adminFieldID = "admin_field_id"
    }
}

// MARK: - Admin Field

public struct AdminField: Codable, Sendable {
    public let adminFieldID: AdminFieldID
    public let parentID: AdminDatatypeID?
    public let label: String
    public let data: String
    public let validation: String
    public let uiConfig: String
    public let type: FieldType
    public let authorID: UserID?
    public let dateCreated: Timestamp
    public let dateModified: Timestamp

    enum CodingKeys: String, CodingKey {
        case adminFieldID = "admin_field_id"
        case parentID = "parent_id"
        case label
        case data
        case validation
        case uiConfig = "ui_config"
        case type
        case authorID = "author_id"
        case dateCreated = "date_created"
        case dateModified = "date_modified"
    }
}

public struct CreateAdminFieldParams: Encodable, Sendable {
    public let parentID: AdminDatatypeID?
    public let label: String
    public let data: String
    public let validation: String
    public let uiConfig: String
    public let type: FieldType
    public let authorID: UserID

    public init(
        parentID: AdminDatatypeID? = nil,
        label: String,
        data: String,
        validation: String,
        uiConfig: String,
        type: FieldType,
        authorID: UserID
    ) {
        self.parentID = parentID
        self.label = label
        self.data = data
        self.validation = validation
        self.uiConfig = uiConfig
        self.type = type
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case parentID = "parent_id"
        case label
        case data
        case validation
        case uiConfig = "ui_config"
        case type
        case authorID = "author_id"
    }
}

public struct UpdateAdminFieldParams: Encodable, Sendable {
    public let adminFieldID: AdminFieldID
    public let parentID: AdminDatatypeID?
    public let label: String
    public let data: String
    public let validation: String
    public let uiConfig: String
    public let type: FieldType
    public let authorID: UserID

    public init(
        adminFieldID: AdminFieldID,
        parentID: AdminDatatypeID? = nil,
        label: String,
        data: String,
        validation: String,
        uiConfig: String,
        type: FieldType,
        authorID: UserID
    ) {
        self.adminFieldID = adminFieldID
        self.parentID = parentID
        self.label = label
        self.data = data
        self.validation = validation
        self.uiConfig = uiConfig
        self.type = type
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case adminFieldID = "admin_field_id"
        case parentID = "parent_id"
        case label
        case data
        case validation
        case uiConfig = "ui_config"
        case type
        case authorID = "author_id"
    }
}

// MARK: - Field Type

public struct FieldTypeInfo: Codable, Sendable {
    public let fieldTypeID: FieldTypeID
    public let type: String
    public let label: String

    enum CodingKeys: String, CodingKey {
        case fieldTypeID = "field_type_id"
        case type, label
    }
}

public struct CreateFieldTypeParams: Encodable, Sendable {
    public let type: String
    public let label: String

    public init(type: String, label: String) {
        self.type = type
        self.label = label
    }

    enum CodingKeys: String, CodingKey {
        case type, label
    }
}

public struct UpdateFieldTypeParams: Encodable, Sendable {
    public let fieldTypeID: FieldTypeID
    public let type: String
    public let label: String

    public init(fieldTypeID: FieldTypeID, type: String, label: String) {
        self.fieldTypeID = fieldTypeID
        self.type = type
        self.label = label
    }

    enum CodingKeys: String, CodingKey {
        case fieldTypeID = "field_type_id"
        case type, label
    }
}

// MARK: - Admin Field Type

public struct AdminFieldTypeInfo: Codable, Sendable {
    public let adminFieldTypeID: AdminFieldTypeID
    public let type: String
    public let label: String

    enum CodingKeys: String, CodingKey {
        case adminFieldTypeID = "admin_field_type_id"
        case type, label
    }
}

public struct CreateAdminFieldTypeParams: Encodable, Sendable {
    public let type: String
    public let label: String

    public init(type: String, label: String) {
        self.type = type
        self.label = label
    }

    enum CodingKeys: String, CodingKey {
        case type, label
    }
}

public struct UpdateAdminFieldTypeParams: Encodable, Sendable {
    public let adminFieldTypeID: AdminFieldTypeID
    public let type: String
    public let label: String

    public init(adminFieldTypeID: AdminFieldTypeID, type: String, label: String) {
        self.adminFieldTypeID = adminFieldTypeID
        self.type = type
        self.label = label
    }

    enum CodingKeys: String, CodingKey {
        case adminFieldTypeID = "admin_field_type_id"
        case type, label
    }
}

// MARK: - Admin Route

public struct AdminRoute: Codable, Sendable {
    public let adminRouteID: AdminRouteID
    public let slug: Slug
    public let title: String
    public let status: Int64
    public let authorID: UserID?
    public let dateCreated: Timestamp
    public let dateModified: Timestamp

    enum CodingKeys: String, CodingKey {
        case adminRouteID = "admin_route_id"
        case slug
        case title
        case status
        case authorID = "author_id"
        case dateCreated = "date_created"
        case dateModified = "date_modified"
    }
}

public struct CreateAdminRouteParams: Encodable, Sendable {
    public let slug: Slug
    public let title: String
    public let status: Int64
    public let authorID: UserID

    public init(slug: Slug, title: String, status: Int64, authorID: UserID) {
        self.slug = slug
        self.title = title
        self.status = status
        self.authorID = authorID
    }

    enum CodingKeys: String, CodingKey {
        case slug
        case title
        case status
        case authorID = "author_id"
    }
}

public struct UpdateAdminRouteParams: Encodable, Sendable {
    public let slug: Slug
    public let title: String
    public let status: Int64
    public let authorID: UserID
    public let slug2: Slug

    public init(slug: Slug, title: String, status: Int64, authorID: UserID, slug2: Slug) {
        self.slug = slug
        self.title = title
        self.status = status
        self.authorID = authorID
        self.slug2 = slug2
    }

    enum CodingKeys: String, CodingKey {
        case slug
        case title
        case status
        case authorID = "author_id"
        case slug2 = "slug_2"
    }
}

// MARK: - Auth Types

public struct LoginParams: Encodable, Sendable {
    public let email: String
    public let password: String

    public init(email: String, password: String) {
        self.email = email
        self.password = password
    }

    enum CodingKeys: String, CodingKey {
        case email
        case password
    }
}

public struct LoginResponse: Codable, Sendable {
    public let userID: UserID
    public let email: Email
    public let username: String
    public let dateCreated: Timestamp

    enum CodingKeys: String, CodingKey {
        case userID = "user_id"
        case email
        case username
        case dateCreated = "date_created"
    }
}

// MARK: - Import Types

public struct ImportResult: Codable, Sendable {
    public let success: Bool
    public let datatypesCreated: Int
    public let fieldsCreated: Int
    public let contentCreated: Int
    public let errors: [String]?
    public let message: String

    enum CodingKeys: String, CodingKey {
        case success
        case datatypesCreated = "datatypes_created"
        case fieldsCreated = "fields_created"
        case contentCreated = "content_created"
        case errors
        case message
    }
}

// MARK: - Change Event (read-only)

public struct ChangeEvent: Codable, Sendable {
    public let eventID: EventID
    public let hlcTimestamp: Int64
    public let wallTimestamp: Timestamp
    public let nodeID: String
    public let tableName: String
    public let recordID: String
    public let operation: String
    public let action: String?
    public let userID: UserID?
    public let oldValues: JSONValue?
    public let newValues: JSONValue?
    public let metadata: JSONValue?
    public let requestID: String?
    public let ip: String?
    public let syncedAt: Timestamp?
    public let consumedAt: Timestamp?

    enum CodingKeys: String, CodingKey {
        case eventID = "event_id"
        case hlcTimestamp = "hlc_timestamp"
        case wallTimestamp = "wall_timestamp"
        case nodeID = "node_id"
        case tableName = "table_name"
        case recordID = "record_id"
        case operation
        case action
        case userID = "user_id"
        case oldValues = "old_values"
        case newValues = "new_values"
        case metadata
        case requestID = "request_id"
        case ip
        case syncedAt = "synced_at"
        case consumedAt = "consumed_at"
    }
}

// MARK: - Backup (read-only)

public struct Backup: Codable, Sendable {
    public let backupID: BackupID
    public let nodeID: String
    public let backupType: String
    public let status: String
    public let startedAt: Timestamp
    public let completedAt: Timestamp
    public let durationMs: Int64?
    public let recordCount: Int64?
    public let sizeBytes: Int64?
    public let replicationLsn: String?
    public let hlcTimestamp: Int64
    public let storagePath: String
    public let checksum: String?
    public let triggeredBy: String?
    public let errorMessage: String?
    public let metadata: JSONValue?

    enum CodingKeys: String, CodingKey {
        case backupID = "backup_id"
        case nodeID = "node_id"
        case backupType = "backup_type"
        case status
        case startedAt = "started_at"
        case completedAt = "completed_at"
        case durationMs = "duration_ms"
        case recordCount = "record_count"
        case sizeBytes = "size_bytes"
        case replicationLsn = "replication_lsn"
        case hlcTimestamp = "hlc_timestamp"
        case storagePath = "storage_path"
        case checksum
        case triggeredBy = "triggered_by"
        case errorMessage = "error_message"
        case metadata
    }
}

// MARK: - Plugin Types

/// Summary item returned by the plugin list endpoint.
public struct PluginListItem: Codable, Sendable {
    public let name: String
    public let version: String
    public let description: String
    public let state: String
    public let circuitBreakerState: String?

    enum CodingKeys: String, CodingKey {
        case name
        case version
        case description
        case state
        case circuitBreakerState = "circuit_breaker_state"
    }
}

/// Schema drift entry describing a missing or extra column.
public struct DriftEntry: Codable, Sendable {
    public let table: String
    public let kind: String
    public let column: String
}

/// Detailed plugin information.
public struct PluginInfo: Codable, Sendable {
    public let name: String
    public let version: String
    public let description: String
    public let author: String?
    public let license: String?
    public let state: String
    public let failedReason: String?
    public let circuitBreakerState: String?
    public let circuitBreakerErrors: Int?
    public let vmsAvailable: Int
    public let vmsTotal: Int
    public let dependencies: [String]?
    public let schemaDrift: [DriftEntry]?

    enum CodingKeys: String, CodingKey {
        case name
        case version
        case description
        case author
        case license
        case state
        case failedReason = "failed_reason"
        case circuitBreakerState = "circuit_breaker_state"
        case circuitBreakerErrors = "circuit_breaker_errors"
        case vmsAvailable = "vms_available"
        case vmsTotal = "vms_total"
        case dependencies
        case schemaDrift = "schema_drift"
    }
}

/// Response from plugin reload action.
public struct PluginActionResponse: Codable, Sendable {
    public let ok: Bool
    public let plugin: String
}

/// Response from plugin enable/disable actions.
public struct PluginStateResponse: Codable, Sendable {
    public let ok: Bool
    public let plugin: String
    public let state: String
}

/// Response from cleanup dry-run (GET).
public struct CleanupDryRunResponse: Codable, Sendable {
    public let orphanedTables: [String]
    public let count: Int
    public let action: String

    enum CodingKeys: String, CodingKey {
        case orphanedTables = "orphaned_tables"
        case count
        case action
    }
}

/// Parameters for cleanup drop (POST).
public struct CleanupDropParams: Codable, Sendable {
    public let confirm: Bool
    public let tables: [String]

    public init(confirm: Bool, tables: [String]) {
        self.confirm = confirm
        self.tables = tables
    }
}

/// Response from cleanup drop (POST).
public struct CleanupDropResponse: Codable, Sendable {
    public let dropped: [String]
    public let count: Int
}

/// A route registered by a plugin.
public struct PluginRoute: Codable, Sendable {
    public let plugin: String
    public let method: String
    public let path: String
    public let isPublic: Bool
    public let approved: Bool

    enum CodingKeys: String, CodingKey {
        case plugin
        case method
        case path
        case isPublic = "public"
        case approved
    }
}

/// Identifies a specific route for approval/revocation.
public struct RouteApprovalItem: Codable, Sendable {
    public let plugin: String
    public let method: String
    public let path: String

    public init(plugin: String, method: String, path: String) {
        self.plugin = plugin
        self.method = method
        self.path = path
    }
}

/// A hook registered by a plugin.
public struct PluginHook: Codable, Sendable {
    public let pluginName: String
    public let event: String
    public let table: String
    public let priority: Int
    public let approved: Bool
    public let isWildcard: Bool

    enum CodingKeys: String, CodingKey {
        case pluginName = "plugin_name"
        case event
        case table
        case priority
        case approved
        case isWildcard = "is_wildcard"
    }
}

/// Identifies a specific hook for approval/revocation.
public struct HookApprovalItem: Codable, Sendable {
    public let plugin: String
    public let event: String
    public let table: String

    public init(plugin: String, event: String, table: String) {
        self.plugin = plugin
        self.event = event
        self.table = table
    }
}
