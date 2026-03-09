import Foundation

public final class RolePermissionsResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    public func list() async throws -> [RolePermission] {
        try await http.get(path: "/api/v1/role-permissions")
    }

    public func get(id: RolePermissionID) async throws -> RolePermission {
        let queryItems = [URLQueryItem(name: "q", value: id.rawValue)]
        return try await http.get(path: "/api/v1/role-permissions/", queryItems: queryItems)
    }

    public func create(params: CreateRolePermissionParams) async throws -> RolePermission {
        try await http.post(path: "/api/v1/role-permissions", body: params)
    }

    public func delete(id: RolePermissionID) async throws {
        let queryItems = [URLQueryItem(name: "q", value: id.rawValue)]
        try await http.delete(path: "/api/v1/role-permissions/", queryItems: queryItems)
    }

    public func listByRole(roleID: RoleID) async throws -> [RolePermission] {
        let queryItems = [URLQueryItem(name: "q", value: roleID.rawValue)]
        return try await http.get(path: "/api/v1/role-permissions/role/", queryItems: queryItems)
    }
}
