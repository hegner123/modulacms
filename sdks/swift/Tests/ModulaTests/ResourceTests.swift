import XCTest
@testable import Modula

final class ResourceTests: XCTestCase {

    func testList() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertEqual(request.httpMethod, "GET")
            XCTAssertTrue(request.url!.absoluteString.contains("/api/v1/roles"))
            return mockResponse(json: """
            [{"role_id":"r1","label":"Admin"}]
            """)
        }
        let client = try makeTestClient()
        let roles: [Role] = try awaitAsync { try await client.roles.list() }
        XCTAssertEqual(roles.count, 1)
        XCTAssertEqual(roles[0].roleID, RoleID("r1"))
        XCTAssertEqual(roles[0].label, "Admin")
    }

    func testGet() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertEqual(request.httpMethod, "GET")
            XCTAssertTrue(request.url!.absoluteString.contains("q=r1"))
            XCTAssertTrue(request.url!.absoluteString.contains("/api/v1/roles/"))
            return mockResponse(json: """
            {"role_id":"r1","label":"Admin"}
            """)
        }
        let client = try makeTestClient()
        let role: Role = try awaitAsync { try await client.roles.get(id: RoleID("r1")) }
        XCTAssertEqual(role.roleID, RoleID("r1"))
    }

    func testCreate() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertEqual(request.httpMethod, "POST")
            XCTAssertTrue(request.url!.absoluteString.contains("/api/v1/roles"))
            let body = readBody(from: request)
            XCTAssertNotNil(body)
            return mockResponse(json: """
            {"role_id":"r2","label":"Editor"}
            """)
        }
        let client = try makeTestClient()
        let role: Role = try awaitAsync {
            try await client.roles.create(params: CreateRoleParams(label: "Editor"))
        }
        XCTAssertEqual(role.label, "Editor")
    }

    func testUpdate() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertEqual(request.httpMethod, "PUT")
            XCTAssertTrue(request.url!.absoluteString.contains("/api/v1/roles/"))
            return mockResponse(json: """
            {"role_id":"r1","label":"SuperAdmin"}
            """)
        }
        let client = try makeTestClient()
        let role: Role = try awaitAsync {
            try await client.roles.update(params: UpdateRoleParams(roleID: RoleID("r1"), label: "SuperAdmin"))
        }
        XCTAssertEqual(role.label, "SuperAdmin")
    }

    func testDelete() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertEqual(request.httpMethod, "DELETE")
            XCTAssertTrue(request.url!.absoluteString.contains("q=r1"))
            return mockResponse(json: "")
        }
        let client = try makeTestClient()
        try awaitAsync { try await client.roles.delete(id: RoleID("r1")) }
    }

    func testListPaginated() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertTrue(request.url!.absoluteString.contains("limit=10"))
            XCTAssertTrue(request.url!.absoluteString.contains("offset=5"))
            return mockResponse(json: """
            {"data":[{"role_id":"r1","label":"Admin"}],"total":1,"limit":10,"offset":5}
            """)
        }
        let client = try makeTestClient()
        let result: PaginatedResponse<Role> = try awaitAsync {
            try await client.roles.listPaginated(params: PaginationParams(limit: 10, offset: 5))
        }
        XCTAssertEqual(result.data.count, 1)
        XCTAssertEqual(result.total, 1)
        XCTAssertEqual(result.limit, 10)
        XCTAssertEqual(result.offset, 5)
    }

    func testCount() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertTrue(request.url!.absoluteString.contains("count=true"))
            return mockResponse(json: """
            {"count":42}
            """)
        }
        let client = try makeTestClient()
        let count = try awaitAsync { try await client.roles.count() }
        XCTAssertEqual(count, 42)
    }

    func testNotFoundError() throws {
        MockURLProtocol.requestHandler = { _ in
            mockResponse(statusCode: 404, json: """
            {"message":"not found"}
            """)
        }
        let client = try makeTestClient()
        XCTAssertThrowsError(try awaitAsync { try await client.roles.get(id: RoleID("missing")) }) { error in
            XCTAssertTrue(isNotFound(error))
            XCTAssertFalse(isUnauthorized(error))
        }
    }
}
