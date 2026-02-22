import Foundation

// MARK: - ResourceID Protocol

public protocol ResourceID: Codable, Sendable, Hashable, CustomStringConvertible, ExpressibleByStringLiteral {
    var rawValue: String { get }
    init(_ rawValue: String)
}

public extension ResourceID {
    var description: String { rawValue }
    var isZero: Bool { rawValue.isEmpty }

    init(stringLiteral value: String) {
        self.init(value)
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        let value = try container.decode(String.self)
        self.init(value)
    }

    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        try container.encode(rawValue)
    }
}

// MARK: - Content IDs

public struct ContentID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct ContentFieldID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct ContentRelationID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

// MARK: - Admin Content IDs

public struct AdminContentID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct AdminContentFieldID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct AdminContentRelationID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

// MARK: - Schema IDs

public struct DatatypeID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct DatatypeFieldID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct FieldID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

// MARK: - Admin Schema IDs

public struct AdminDatatypeID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct AdminDatatypeFieldID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct AdminFieldID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

// MARK: - Field Type IDs

public struct FieldTypeID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct AdminFieldTypeID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

// MARK: - Media IDs

public struct MediaID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct MediaDimensionID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

// MARK: - Auth IDs

public struct UserID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct RoleID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct SessionID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct TokenID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct UserOauthID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct UserSshKeyID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct PermissionID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct RolePermissionID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

// MARK: - Routing IDs

public struct RouteID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct AdminRouteID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

// MARK: - Other IDs

public struct TableID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct EventID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct BackupID: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

// MARK: - Value Types

public struct Slug: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct Email: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

public struct URLValue: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }
}

// MARK: - Timestamp

public struct Timestamp: ResourceID {
    public let rawValue: String
    public init(_ rawValue: String) { self.rawValue = rawValue }

    private static let formatter: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime]
        return f
    }()

    public func date() -> Date? {
        Timestamp.formatter.date(from: rawValue)
    }

    public init(date: Date) {
        self.rawValue = Timestamp.formatter.string(from: date)
    }

    public static func now() -> Timestamp {
        Timestamp(date: Date())
    }
}
