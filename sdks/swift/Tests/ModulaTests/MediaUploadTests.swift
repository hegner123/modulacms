import XCTest
@testable import Modula

final class MediaUploadTests: XCTestCase {

    func testUploadSendsMultipartFormData() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertEqual(request.httpMethod, "POST")
            let contentType = request.value(forHTTPHeaderField: "Content-Type") ?? ""
            XCTAssertTrue(contentType.hasPrefix("multipart/form-data; boundary="))

            let bodyData = readBody(from: request) ?? Data()
            // Use .isoLatin1 because body contains binary file data (PNG magic bytes)
            // that makes .utf8 return nil
            let bodyStr = String(data: bodyData, encoding: .isoLatin1) ?? ""
            XCTAssertTrue(bodyStr.contains("name=\"file\""), "Body should contain form field name")
            XCTAssertTrue(bodyStr.contains("test.png"), "Body should contain filename")

            return mockResponse(json: """
            {
                "media_id": "m1",
                "name": "test.png",
                "display_name": "test.png",
                "alt": null,
                "caption": null,
                "description": null,
                "class": null,
                "mimetype": "image/png",
                "dimensions": "100x100",
                "url": "https://cdn.example.com/test.png",
                "srcset": null,
                "focal_x": null,
                "focal_y": null,
                "author_id": "u1",
                "date_created": "2024-01-01T00:00:00Z",
                "date_modified": "2024-01-01T00:00:00Z"
            }
            """)
        }

        let client = try makeTestClient()
        let fileData = Data([0x89, 0x50, 0x4E, 0x47]) // PNG magic bytes
        let media: Media = try awaitAsync {
            try await client.mediaUpload.upload(data: fileData, filename: "test.png")
        }
        XCTAssertEqual(media.mediaID, MediaID("m1"))
        XCTAssertEqual(media.name, "test.png")
        XCTAssertEqual(media.url, URLValue("https://cdn.example.com/test.png"))
    }

    func testUploadAuthorizationHeader() throws {
        MockURLProtocol.requestHandler = { request in
            XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer test-key")
            return mockResponse(json: """
            {
                "media_id": "m1",
                "name": "file.jpg",
                "display_name": null,
                "alt": null,
                "caption": null,
                "description": null,
                "class": null,
                "mimetype": "image/jpeg",
                "dimensions": null,
                "url": "https://cdn.example.com/file.jpg",
                "srcset": null,
                "focal_x": null,
                "focal_y": null,
                "author_id": null,
                "date_created": "2024-01-01T00:00:00Z",
                "date_modified": "2024-01-01T00:00:00Z"
            }
            """)
        }

        let client = try makeTestClient()
        let _: Media = try awaitAsync {
            try await client.mediaUpload.upload(data: Data([0xFF, 0xD8]), filename: "file.jpg")
        }
    }

    func testUploadErrorResponse() throws {
        MockURLProtocol.requestHandler = { _ in
            mockResponse(statusCode: 413, json: """
            {"message":"file too large"}
            """)
        }

        let client = try makeTestClient()
        XCTAssertThrowsError(try awaitAsync {
            try await client.mediaUpload.upload(data: Data(repeating: 0, count: 100), filename: "big.bin")
        }) { error in
            let apiError = error as? APIError
            XCTAssertNotNil(apiError)
            XCTAssertEqual(apiError?.statusCode, 413)
        }
    }

    func testUploadNoMimeTypeParameter() throws {
        // Verify the upload signature matches Go SDK (data + filename only, no mimeType)
        MockURLProtocol.requestHandler = { _ in
            return mockResponse(json: """
            {
                "media_id": "m1",
                "name": "doc.pdf",
                "display_name": null,
                "alt": null,
                "caption": null,
                "description": null,
                "class": null,
                "mimetype": "application/pdf",
                "dimensions": null,
                "url": "https://cdn.example.com/doc.pdf",
                "srcset": null,
                "focal_x": null,
                "focal_y": null,
                "author_id": null,
                "date_created": "2024-01-01T00:00:00Z",
                "date_modified": "2024-01-01T00:00:00Z"
            }
            """)
        }

        let client = try makeTestClient()
        // Signature is upload(data:filename:) — no mimeType parameter
        let media: Media = try awaitAsync {
            try await client.mediaUpload.upload(data: Data([0x25, 0x50, 0x44, 0x46]), filename: "doc.pdf")
        }
        XCTAssertEqual(media.mediaID, MediaID("m1"))
    }
}
