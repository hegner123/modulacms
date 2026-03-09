// MARK: - ContentStatus

public enum ContentStatus: String, Codable, Sendable {
    case draft
    case published
}

// MARK: - FieldType

public enum FieldType: String, Codable, Sendable {
    case text
    case textarea
    case number
    case date
    case datetime
    case boolean
    case select
    case media
    case id = "_id"
    case json
    case richtext
    case slug
    case email
    case url
}

// MARK: - RouteType

public enum RouteType: String, Codable, Sendable {
    case `static`
    case dynamic
    case api
    case redirect
}
