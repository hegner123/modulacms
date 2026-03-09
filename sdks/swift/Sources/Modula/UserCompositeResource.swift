import Foundation

/// Provides composite user operations that combine deletion with content reassignment.
///
/// ``reassignDelete(params:)`` deletes a user and reassigns all of their owned content
/// (content_data, datatypes, admin_content_data) to another user in a single request.
public final class UserCompositeResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    /// Delete a user and reassign all of their content to another user.
    ///
    /// - Parameter params: The user to delete and the optional target user for reassignment.
    /// - Returns: A ``UserReassignDeleteResponse`` with counts of reassigned entities.
    public func reassignDelete(params: UserReassignDeleteParams) async throws -> UserReassignDeleteResponse {
        try await http.post(path: "/api/v1/users/reassign-delete", body: params)
    }
}
