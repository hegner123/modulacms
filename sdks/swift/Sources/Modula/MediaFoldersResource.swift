import Foundation

public final class MediaFoldersResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns the full media folder tree as a nested hierarchy.
    public func tree() async throws -> [MediaFolderTreeNode] {
        try await http.get(path: "/api/v1/media-folders/tree")
    }

    /// Returns a paginated list of media items within a specific folder.
    public func listMedia(folderID: MediaFolderID, params: PaginationParams) async throws -> PaginatedResponse<Media> {
        let queryItems = [
            URLQueryItem(name: "limit", value: String(params.limit)),
            URLQueryItem(name: "offset", value: String(params.offset)),
        ]
        return try await http.get(
            path: "/api/v1/media-folders/\(folderID.rawValue)/media",
            queryItems: queryItems
        )
    }

    /// Batch-moves media items to a target folder (or to root if folderID is nil).
    public func moveMedia(params: MoveMediaParams) async throws -> MoveMediaResponse {
        try await http.post(path: "/api/v1/media/move", body: params)
    }
}
