import XCTest
@testable import Modula

final class IDsTests: XCTestCase {

    func testIDDescription() {
        let id = ContentID("abc123")
        XCTAssertEqual(id.description, "abc123")
        XCTAssertEqual(String(describing: id), "abc123")
    }

    func testIDIsZero() {
        let empty = ContentID("")
        XCTAssertTrue(empty.isZero)

        let nonEmpty = ContentID("abc")
        XCTAssertFalse(nonEmpty.isZero)
    }

    func testTimestampDate() {
        let ts = Timestamp("2024-01-15T10:30:00Z")
        let date = ts.date()
        XCTAssertNotNil(date)

        let calendar = Calendar(identifier: .gregorian)
        let components = calendar.dateComponents(in: TimeZone(identifier: "UTC")!, from: date!)
        XCTAssertEqual(components.year, 2024)
        XCTAssertEqual(components.month, 1)
        XCTAssertEqual(components.day, 15)
    }

    func testTimestampRoundTrip() {
        let original = Timestamp("2024-06-15T12:00:00Z")
        guard let date = original.date() else {
            XCTFail("Failed to parse timestamp")
            return
        }
        let roundTripped = Timestamp(date: date)
        XCTAssertEqual(original.rawValue, roundTripped.rawValue)
    }

    func testTimestampNow() {
        let ts = Timestamp.now()
        XCTAssertFalse(ts.isZero)
        XCTAssertNotNil(ts.date())
    }

    func testIDCodableRoundTrip() throws {
        let original = ContentID("test-ulid-123")
        let data = try JSONEncoder().encode(original)
        let decoded = try JSONDecoder().decode(ContentID.self, from: data)
        XCTAssertEqual(original, decoded)
    }

    func testStringLiteral() {
        let id: ContentID = "literal-id"
        XCTAssertEqual(id.rawValue, "literal-id")
    }

    func testSlugType() {
        let slug = Slug("my-page")
        XCTAssertEqual(slug.rawValue, "my-page")
        XCTAssertFalse(slug.isZero)
    }

    func testEmailType() {
        let email = Email("user@example.com")
        XCTAssertEqual(email.rawValue, "user@example.com")
    }

    func testURLValueType() {
        let url = URLValue("https://example.com/image.png")
        XCTAssertEqual(url.rawValue, "https://example.com/image.png")
    }
}
