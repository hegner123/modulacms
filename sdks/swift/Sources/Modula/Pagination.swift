public struct PaginationParams: Sendable {
    public let limit: Int64
    public let offset: Int64

    public init(limit: Int64, offset: Int64) {
        self.limit = limit
        self.offset = offset
    }
}

public struct PaginatedResponse<T: Decodable & Sendable>: Decodable, Sendable {
    public let data: [T]
    public let total: Int64
    public let limit: Int64
    public let offset: Int64

    enum CodingKeys: String, CodingKey {
        case data
        case total
        case limit
        case offset
    }
}
