import XCTest
@testable import Modula

final class TypesCodableTests: XCTestCase {

    func testContentDataDecodesFromSnakeCase() throws {
        let json = """
        {
            "content_data_id": "cd1",
            "parent_id": null,
            "first_child_id": null,
            "next_sibling_id": null,
            "prev_sibling_id": null,
            "route_id": "r1",
            "datatype_id": "dt1",
            "author_id": "u1",
            "status": "published",
            "published_at": null,
            "published_by": null,
            "publish_at": null,
            "revision": 1,
            "date_created": "2024-01-01T00:00:00Z",
            "date_modified": "2024-01-02T00:00:00Z"
        }
        """.data(using: .utf8)!

        let content = try JSON.decoder.decode(ContentData.self, from: json)
        XCTAssertEqual(content.contentDataID, ContentID("cd1"))
        XCTAssertNil(content.parentID)
        XCTAssertEqual(content.routeID, RouteID("r1"))
        XCTAssertEqual(content.datatypeID, DatatypeID("dt1"))
        XCTAssertEqual(content.authorID, UserID("u1"))
        XCTAssertEqual(content.status, .published)
    }

    func testCreateParamsEncodesToSnakeCase() throws {
        let params = CreateRoleParams(label: "Admin")
        let data = try JSON.encoder.encode(params)
        let dict = try JSONSerialization.jsonObject(with: data) as! [String: Any]
        XCTAssertEqual(dict["label"] as? String, "Admin")
    }

    func testMediaOptionalFieldsDecodeNull() throws {
        let json = """
        {
            "media_id": "m1",
            "name": null,
            "display_name": null,
            "alt": null,
            "caption": null,
            "description": null,
            "class": null,
            "mimetype": "image/png",
            "dimensions": null,
            "url": "https://example.com/img.png",
            "srcset": null,
            "focal_x": null,
            "focal_y": null,
            "author_id": null,
            "date_created": "2024-01-01T00:00:00Z",
            "date_modified": "2024-01-01T00:00:00Z"
        }
        """.data(using: .utf8)!

        let media = try JSON.decoder.decode(Media.self, from: json)
        XCTAssertEqual(media.mediaID, MediaID("m1"))
        XCTAssertNil(media.name)
        XCTAssertNil(media.displayName)
        XCTAssertNil(media.alt)
        XCTAssertEqual(media.mimetype, "image/png")
        XCTAssertEqual(media.url, URLValue("https://example.com/img.png"))
    }

    func testJSONValueVariants() throws {
        let json = """
        {
            "str": "hello",
            "num": 42.5,
            "bool": true,
            "null_val": null,
            "arr": [1, 2],
            "obj": {"key": "val"}
        }
        """.data(using: .utf8)!

        let dict = try JSONDecoder().decode([String: JSONValue].self, from: json)
        XCTAssertEqual(dict["str"], .string("hello"))
        XCTAssertEqual(dict["num"], .number(42.5))
        XCTAssertEqual(dict["bool"], .bool(true))
        XCTAssertEqual(dict["null_val"], .null)
        XCTAssertEqual(dict["arr"], .array([.number(1), .number(2)]))
        XCTAssertEqual(dict["obj"], .object(["key": .string("val")]))
    }

    func testUpdateRouteParamsSlug2Encoding() throws {
        let params = UpdateRouteParams(
            slug: Slug("new-slug"),
            title: "Title",
            status: 1,
            authorID: UserID("u-author-1"),
            slug2: Slug("old-slug")
        )
        let data = try JSON.encoder.encode(params)
        let dict = try JSONSerialization.jsonObject(with: data) as! [String: Any]
        XCTAssertEqual(dict["slug"] as? String, "new-slug")
        XCTAssertEqual(dict["slug_2"] as? String, "old-slug")
    }

    func testChangeEventOptionalFields() throws {
        let json = """
        {
            "event_id": "e1",
            "hlc_timestamp": 12345,
            "wall_timestamp": "2024-01-01T00:00:00Z",
            "node_id": "node1",
            "table_name": "content_data",
            "record_id": "r1",
            "operation": "INSERT",
            "action": null,
            "user_id": null,
            "old_values": null,
            "new_values": {"name": "test"},
            "metadata": null,
            "request_id": null,
            "ip": null,
            "synced_at": null,
            "consumed_at": null
        }
        """.data(using: .utf8)!

        let event = try JSON.decoder.decode(ChangeEvent.self, from: json)
        XCTAssertEqual(event.eventID, EventID("e1"))
        XCTAssertEqual(event.hlcTimestamp, 12345)
        XCTAssertNil(event.action)
        XCTAssertNil(event.userID)
        XCTAssertNil(event.oldValues)
        XCTAssertEqual(event.newValues, .object(["name": .string("test")]))
        XCTAssertNil(event.syncedAt)
    }
}
