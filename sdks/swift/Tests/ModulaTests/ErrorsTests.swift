import XCTest
@testable import Modula

final class ErrorsTests: XCTestCase {

    func testAPIErrorDescription() {
        let err = APIError(statusCode: 404, message: "not found")
        XCTAssertEqual(err.errorDescription, "modula: 404 not found")
    }

    func testAPIErrorDescriptionFallback() {
        let err = APIError(statusCode: 500)
        XCTAssertTrue(err.errorDescription?.contains("500") ?? false)
    }

    func testIsNotFound() {
        let err = APIError(statusCode: 404, message: "not found")
        XCTAssertTrue(isNotFound(err))
        XCTAssertFalse(isUnauthorized(err))
    }

    func testIsUnauthorized() {
        let err = APIError(statusCode: 401, message: "unauthorized")
        XCTAssertTrue(isUnauthorized(err))
        XCTAssertFalse(isNotFound(err))
    }

    func testIsNotFoundWithOtherError() {
        let err = NSError(domain: "test", code: 404)
        XCTAssertFalse(isNotFound(err))
    }
}
