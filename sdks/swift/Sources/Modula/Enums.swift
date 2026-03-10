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
    case title = "_title"
}

// MARK: - RouteType

public enum RouteType: String, Codable, Sendable {
    case `static`
    case dynamic
    case api
    case redirect
}

// MARK: - PluginStatus

public enum PluginStatus: String, Codable, Sendable {
    case installed
    case enabled
}

// MARK: - Operation

public enum Operation: String, Codable, Sendable {
    case insert = "INSERT"
    case update = "UPDATE"
    case delete = "DELETE"
}

// MARK: - Action

public enum Action: String, Codable, Sendable {
    case create
    case update
    case delete
    case publish
}

// MARK: - ConflictPolicy

public enum ConflictPolicy: String, Codable, Sendable {
    case lww
    case manual
}

// MARK: - BackupType

public enum BackupType: String, Codable, Sendable {
    case full
    case incremental
    case differential
}

// MARK: - BackupStatus

public enum BackupStatus: String, Codable, Sendable {
    case pending
    case inProgress = "in_progress"
    case completed
    case failed
}

// MARK: - VerificationStatus

public enum VerificationStatus: String, Codable, Sendable {
    case pending
    case verified
    case failed
}

// MARK: - BackupSetStatus

public enum BackupSetStatus: String, Codable, Sendable {
    case pending
    case complete
    case partial
}
