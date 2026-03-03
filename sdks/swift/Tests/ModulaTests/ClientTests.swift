import XCTest
@testable import Modula

final class ClientTests: XCTestCase {

    func testInitRequiresBaseURL() {
        XCTAssertThrowsError(try ModulaClient(config: ClientConfig(baseURL: ""))) { error in
            let apiError = error as? APIError
            XCTAssertNotNil(apiError)
        }
    }

    func testInitTrimsTrailingSlash() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertTrue(request.url!.absoluteString.hasPrefix("https://cms.example.com/api/"))
            return mockResponse(json: "[]")
        }
        let client = try makeTestClient(baseURL: "https://cms.example.com///")
        let _: [ContentData] = try awaitAsync { try await client.contentData.list() }
    }

    func testInitDefaultURLSession() throws {
        let client = try ModulaClient(config: ClientConfig(
            baseURL: "https://cms.example.com",
            apiKey: "key"
        ))
        XCTAssertNotNil(client)
    }

    func testInitCustomURLSession() throws {
        let config = URLSessionConfiguration.ephemeral
        config.protocolClasses = [MockURLProtocol.self]
        let session = URLSession(configuration: config)
        let client = try ModulaClient(config: ClientConfig(
            baseURL: "https://cms.example.com",
            apiKey: "key",
            urlSession: session
        ))
        XCTAssertNotNil(client)
    }

    func testAllResourcesInitialized() throws {
        let client = try makeTestClient()
        XCTAssertNotNil(client.contentData)
        XCTAssertNotNil(client.contentFields)
        XCTAssertNotNil(client.contentRelations)
        XCTAssertNotNil(client.datatypes)
        XCTAssertNotNil(client.fields)
        XCTAssertNotNil(client.media)
        XCTAssertNotNil(client.mediaDimensions)
        XCTAssertNotNil(client.routes)
        XCTAssertNotNil(client.roles)
        XCTAssertNotNil(client.users)
        XCTAssertNotNil(client.tokens)
        XCTAssertNotNil(client.usersOauth)
        XCTAssertNotNil(client.tables)
        XCTAssertNotNil(client.adminContentData)
        XCTAssertNotNil(client.adminContentFields)
        XCTAssertNotNil(client.adminDatatypes)
        XCTAssertNotNil(client.adminFields)
        XCTAssertNotNil(client.adminRoutes)
        XCTAssertNotNil(client.auth)
        XCTAssertNotNil(client.mediaUpload)
        XCTAssertNotNil(client.adminTree)
        XCTAssertNotNil(client.content)
        XCTAssertNotNil(client.sshKeys)
        XCTAssertNotNil(client.sessions)
        XCTAssertNotNil(client.importResource)
        XCTAssertNotNil(client.contentBatch)
    }
}

func awaitAsync<T>(_ block: @escaping () async throws -> T) throws -> T {
    let expectation = XCTestExpectation(description: "async")
    var result: Result<T, Error>!
    Task {
        do {
            result = .success(try await block())
        } catch {
            result = .failure(error)
        }
        expectation.fulfill()
    }
    _ = XCTWaiter.wait(for: [expectation], timeout: 5.0)
    return try result.get()
}
