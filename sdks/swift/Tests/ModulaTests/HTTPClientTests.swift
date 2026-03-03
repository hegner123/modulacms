import XCTest
@testable import Modula

final class HTTPClientTests: XCTestCase {

    private func makeHTTPClient(apiKey: String = "test-key") -> HTTPClient {
        let config = URLSessionConfiguration.ephemeral
        config.protocolClasses = [MockURLProtocol.self]
        let session = URLSession(configuration: config)
        return HTTPClient(baseURL: "https://cms.example.com", apiKey: apiKey, session: session)
    }

    func testGet() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertEqual(request.httpMethod, "GET")
            return mockResponse(json: """
            [{"role_id":"r1","label":"Admin"}]
            """)
        }
        let http = makeHTTPClient()
        let roles: [Role] = try awaitAsync { try await http.get(path: "/api/v1/roles") }
        XCTAssertEqual(roles.count, 1)
    }

    func testPost() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertEqual(request.httpMethod, "POST")
            XCTAssertEqual(request.value(forHTTPHeaderField: "Content-Type"), "application/json")
            return mockResponse(json: """
            {"role_id":"r1","label":"Admin"}
            """)
        }
        let http = makeHTTPClient()
        let role: Role = try awaitAsync {
            try await http.post(path: "/api/v1/roles", body: CreateRoleParams(label: "Admin"))
        }
        XCTAssertEqual(role.roleID, RoleID("r1"))
    }

    func testPut() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertEqual(request.httpMethod, "PUT")
            return mockResponse(json: """
            {"role_id":"r1","label":"Updated"}
            """)
        }
        let http = makeHTTPClient()
        let role: Role = try awaitAsync {
            try await http.put(path: "/api/v1/roles/", body: UpdateRoleParams(roleID: RoleID("r1"), label: "Updated"))
        }
        XCTAssertEqual(role.label, "Updated")
    }

    func testDelete() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertEqual(request.httpMethod, "DELETE")
            return mockResponse(json: "")
        }
        let http = makeHTTPClient()
        try awaitAsync {
            try await http.delete(path: "/api/v1/roles/", queryItems: [URLQueryItem(name: "q", value: "r1")])
        }
    }

    func testAuthHeader() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer my-secret-key")
            return mockResponse(json: "[]")
        }
        let http = makeHTTPClient(apiKey: "my-secret-key")
        let _: [Role] = try awaitAsync { try await http.get(path: "/api/v1/roles") }
    }

    func testNoAuthHeaderWhenEmpty() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertNil(request.value(forHTTPHeaderField: "Authorization"))
            return mockResponse(json: "[]")
        }
        let http = makeHTTPClient(apiKey: "")
        let _: [Role] = try awaitAsync { try await http.get(path: "/api/v1/roles") }
    }

    func testErrorResponse() throws {
        MockURLProtocol.requestHandler = { _ in
            mockResponse(statusCode: 500, json: """
            {"message":"internal server error"}
            """)
        }
        let http = makeHTTPClient()
        XCTAssertThrowsError(try awaitAsync { try await http.get(path: "/api/v1/roles") as [Role] }) { error in
            let apiError = error as? APIError
            XCTAssertNotNil(apiError)
            XCTAssertEqual(apiError?.statusCode, 500)
            XCTAssertEqual(apiError?.message, "internal server error")
        }
    }

    func testErrorWithJSONErrorField() throws {
        MockURLProtocol.requestHandler = { _ in
            mockResponse(statusCode: 403, json: """
            {"error":"forbidden"}
            """)
        }
        let http = makeHTTPClient()
        XCTAssertThrowsError(try awaitAsync { try await http.get(path: "/api/v1/roles") as [Role] }) { error in
            let apiError = error as? APIError
            XCTAssertEqual(apiError?.message, "forbidden")
        }
    }
}
