import Foundation

/// Provides admin media folder operations: tree navigation, folder media listing, and batch moves.
public final class AdminMediaFoldersResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Returns the full admin media folder tree as a nested hierarchy.
    public func tree() async throws -> [AdminMediaFolderTreeNode] {
        try await http.get(path: "/api/v1/adminmedia-folders/tree")
    }

    /// Returns a paginated list of admin media items within a specific folder.
    public func listMedia(folderID: AdminMediaFolderID, params: PaginationParams) async throws -> PaginatedResponse<AdminMedia> {
        let queryItems = [
            URLQueryItem(name: "limit", value: String(params.limit)),
            URLQueryItem(name: "offset", value: String(params.offset)),
        ]
        return try await http.get(
            path: "/api/v1/adminmedia-folders/\(folderID.rawValue)/media",
            queryItems: queryItems
        )
    }

    /// Batch-moves admin media items to a target folder (or to root if folderID is nil).
    public func moveMedia(params: MoveAdminMediaParams) async throws -> MoveAdminMediaResponse {
        try await http.post(path: "/api/v1/adminmedia/move", body: params)
    }
}
