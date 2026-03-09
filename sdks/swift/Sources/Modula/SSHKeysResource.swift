import Foundation

public final class SSHKeysResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    public func list() async throws -> [SshKeyListItem] {
        try await http.get(path: "/api/v1/ssh-keys")
    }

    public func create(params: CreateSSHKeyParams) async throws -> SshKey {
        try await http.post(path: "/api/v1/ssh-keys", body: params)
    }

    public func delete(id: UserSshKeyID) async throws {
        let queryItems = [URLQueryItem(name: "q", value: id.rawValue)]
        try await http.delete(path: "/api/v1/ssh-keys/", queryItems: queryItems)
    }
}
