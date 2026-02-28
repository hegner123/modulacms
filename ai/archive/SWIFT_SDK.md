 Ready to code?

 Here is Claude's plan:
╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
 Plan: Swift SDK for ModulaCMS

 Context

 ModulaCMS now has TypeScript SDKs (read-only + admin) and a Go SDK. The Swift SDK targets Apple platform developers — iOS apps displaying CMS content, macOS admin tools, etc. It mirrors the
  Go SDK's single-package architecture, translating its generic Resource pattern into Swift's type system with async/await, Codable, and branded ID types via protocols.

 Decisions

 - Apple-only: iOS 16+, macOS 13+, tvOS 16+, watchOS 9+ — uses Foundation/URLSession, zero dependencies
 - Single SPM package: One library product ModulaCMS, matching the Go SDK pattern
 - Swift 5.9+: async/await, Sendable, generics — widely available on all target platforms
 - Never for Media create: Resource<Media, Never, UpdateMediaParams, MediaID> makes create() uncallable at compile time
 - URLProtocol mocking: Standard Swift test infrastructure, no test server needed
 - Data for raw JSON returns: Variable-format responses (admin tree, content delivery, import, batch) return Data — callers decode as needed

 Directory Structure

 sdks/swift/
   Package.swift
   Sources/
     ModulaCMS/
       Client.swift                    -- ModulaCMSClient, ClientConfig
       Resource.swift                  -- Generic Resource<E, C, U, ID> CRUD
       HTTPClient.swift                -- Internal URLSession transport
       IDs.swift                       -- ResourceID protocol + 28 branded ID types
       Enums.swift                     -- ContentStatus, FieldType, RouteType
       Types.swift                     -- All entity structs + Create/Update params
       Errors.swift                    -- APIError, isNotFound(), isUnauthorized()
       Pagination.swift                -- PaginationParams, PaginatedResponse<T>
       JSON.swift                      -- Shared JSONDecoder/Encoder with snake_case strategy + JSONValue enum
       AuthResource.swift              -- Login, logout, me, register, reset
       MediaUploadResource.swift       -- Multipart file upload
       AdminTreeResource.swift         -- Admin content tree
       ContentDeliveryResource.swift   -- Public slug-based content delivery
       SSHKeysResource.swift           -- SSH key management
       SessionsResource.swift          -- Session management
       ImportResource.swift            -- Import from various CMS formats
       ContentBatchResource.swift      -- Batch content updates
   Tests/
     ModulaCMSTests/
       MockURLProtocol.swift           -- URLProtocol subclass for test mocking
       ClientTests.swift               -- Client construction tests
       ResourceTests.swift             -- Generic resource CRUD tests
       HTTPClientTests.swift           -- HTTP layer tests
       ErrorsTests.swift               -- Error type tests
       IDsTests.swift                  -- ID type + Timestamp tests
       TypesCodableTests.swift         -- Codable round-trip tests

 18 source files + 7 test files + Package.swift = 26 files total.

 Implementation Phases

 Phase 1 — Foundation (sequential)

 1. Package.swift

 // swift-tools-version: 5.9
 import PackageDescription
 let package = Package(
     name: "ModulaCMS",
     platforms: [.iOS(.v16), .macOS(.v13), .tvOS(.v16), .watchOS(.v9)],
     products: [.library(name: "ModulaCMS", targets: ["ModulaCMS"])],
     targets: [
         .target(name: "ModulaCMS", path: "Sources/ModulaCMS"),
         .testTarget(name: "ModulaCMSTests", dependencies: ["ModulaCMS"], path: "Tests/ModulaCMSTests")
     ]
 )

 2. IDs.swift — ResourceID protocol with default implementations for Codable, CustomStringConvertible, ExpressibleByStringLiteral, isZero. Then 28 minimal structs. Timestamp gets extra
 date() -> Date?, init(date:), static func now() using ISO8601DateFormatter. Value type URLValue (not URL) avoids Foundation collision.

 3. Enums.swift — String-backed enums: ContentStatus, FieldType, RouteType. Note: RouteType.static uses backtick (`static`) since static is a keyword.

 4. Errors.swift — APIError struct conforming to Error + LocalizedError + Sendable. Free functions isNotFound(_:) and isUnauthorized(_:).

 5. Pagination.swift — PaginationParams and PaginatedResponse<T: Decodable & Sendable>.

 6. JSON.swift — Shared JSON.decoder / JSON.encoder with .convertFromSnakeCase / .convertToSnakeCase strategies. Also contains JSONValue enum (string/number/bool/object/array/null) as the
 Swift equivalent of Go's json.RawMessage for struct fields.

 Phase 2 — HTTP + Types (parallel)

 Agent A: HTTPClient.swift

 Sendable final class (all state immutable after init, URLSession is thread-safe). Methods:
 - get<T: Decodable>(path:queryItems:) async throws -> T
 - post<T: Decodable, B: Encodable>(path:body:) async throws -> T
 - put<T: Decodable, B: Encodable>(path:body:) async throws -> T
 - delete(path:queryItems:) async throws
 - executeRaw(_ request: URLRequest) async throws -> (Data, HTTPURLResponse)
 - Void overloads: post<B: Encodable>(path:body:) async throws, postNoBody(path:) async throws

 URL construction: string concatenation baseURL.absoluteString + path (preserves trailing slashes like Go SDK). Auth: Bearer {apiKey} header when non-empty. Error:
 buildError(statusCode:data:) extracts message or error from JSON body.

 Agent B: Types.swift

 Translate Go SDK types.go 1:1. All entity structs conform to Codable & Sendable. Create params conform to Encodable & Sendable. Nullable Go pointers → Swift optionals. All get public
 init(...) with default nil for optionals.

 Structs needing explicit CodingKeys: UpdateRouteParams and UpdateAdminRouteParams (to map slug2 → slug_2 JSON key, or use a clearer name like identifierSlug).

 Full entity list (22 entities, ~20 Create params, ~21 Update params, plus auth/import/read-only types) — reference Go SDK types.go.

 Phase 3 — Resource (after Phase 2)

 Resource.swift — Generic class:
 public final class Resource<
     Entity: Decodable & Sendable,
     CreateParams: Encodable & Sendable,
     UpdateParams: Encodable & Sendable,
     ID: ResourceID
 >: Sendable

 Methods: list(), get(id:), create(params:), update(params:), delete(id:), listPaginated(params:), count(), rawList(queryItems:). Internal CountResponse struct.

 Phase 4 — Specialized Resources (parallel after Phase 3)

 8 files, each a Sendable final class:

 ┌───────────────────────────────┬───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │             File              │                                                                      Methods                                                                      │
 ├───────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ AuthResource.swift            │ login(params:), logout(), me(), register(params:), resetPassword(params:)                                                                         │
 ├───────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ MediaUploadResource.swift     │ upload(data:filename:mimeType:) — multipart/form-data with UUID boundary                                                                          │
 ├───────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ AdminTreeResource.swift       │ get(slug:format:) -> Data                                                                                                                         │
 ├───────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ ContentDeliveryResource.swift │ getPage(slug:format:) -> Data                                                                                                                     │
 ├───────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ SSHKeysResource.swift         │ list(), create(params:), delete(id:)                                                                                                              │
 ├───────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ SessionsResource.swift        │ update(params:), remove(id:)                                                                                                                      │
 ├───────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ ImportResource.swift          │ contentful(data:), sanity(data:), strapi(data:), wordpress(data:), clean(data:), bulk(data:) — all generic <T: Encodable & Sendable>, return Data │
 ├───────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ ContentBatchResource.swift    │ update(request:) -> Data                                                                                                                          │
 └───────────────────────────────┴───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

 Phase 5 — Client Assembly

 Client.swift — ModulaCMSClient (not just Client — avoids namespace collisions in Swift):
 - ClientConfig struct: baseURL: String, apiKey: String?, urlSession: URLSession?
 - init(config:) throws — validates URL, trims trailing slashes, creates HTTPClient, wires 14 standard + 6 admin + 8 specialized resources
 - Property importResource (not import — Swift keyword)
 - Media resource: Resource<Media, Never, UpdateMediaParams, MediaID>

 Phase 6 — Tests (parallel)

 ┌─────────────────────────┬────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │          File           │                                                                     Tests                                                                      │
 ├─────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ MockURLProtocol.swift   │ URLProtocol subclass + makeTestClient() helper                                                                                                 │
 ├─────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ ClientTests.swift       │ testInit_requiresBaseURL, testInit_trimsTrailingSlash, testInit_defaultURLSession, testInit_customURLSession, testInit_allResourcesInitialized │
 ├─────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ ResourceTests.swift     │ testList, testGet, testCreate, testUpdate, testDelete, testListPaginated, testCount, testNotFoundError                                         │
 ├─────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ HTTPClientTests.swift   │ testGet, testPost, testPut, testDelete, testAuthHeader, testErrorResponse, testErrorWithJSONMessage, testCancellation                          │
 ├─────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ ErrorsTests.swift       │ testAPIError_errorDescription, testIsNotFound, testIsUnauthorized                                                                              │
 ├─────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ IDsTests.swift          │ testID_description, testID_isZero, testTimestamp_date, testTimestamp_roundTrip, testTimestampNow, testID_codableRoundTrip                      │
 ├─────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ TypesCodableTests.swift │ testContentData_decodesFromSnakeCase, testCreateParams_encodesToSnakeCase, testMedia_optionalFieldsDecodeNull, testJSONValue_variants          │
 └─────────────────────────┴────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

 Phase 7 — Integration

 justfile — add after sdk-go-vet:
 sdk-swift-build:
     cd sdks/swift && swift build
 sdk-swift-test:
     cd sdks/swift && swift test
 sdk-swift-clean:
     cd sdks/swift && swift package clean

 CI — add sdks/swift/** to path triggers, new swift-sdk job on macos-14 runner (has Xcode 15+ / Swift 5.9+):
 swift-sdk:
   runs-on: macos-14
   steps:
     - uses: actions/checkout@v4
     - name: Build
       working-directory: sdks/swift
       run: swift build
     - name: Test
       working-directory: sdks/swift
       run: swift test

 CLAUDE.md — add Swift SDK section after Go SDK section, add sdks/swift/ to Package Map.

 Key Design Details

 ┌──────────────────┬─────────────────────────────────────────┬─────────────────────────────────┐
 │      Aspect      │             Swift Approach              │          Go Equivalent          │
 ├──────────────────┼─────────────────────────────────────────┼─────────────────────────────────┤
 │ ID types         │ ResourceID protocol + 28 structs        │ type ContentID string           │
 ├──────────────────┼─────────────────────────────────────────┼─────────────────────────────────┤
 │ Nullable fields  │ String?, ContentID?                     │ *string, *ContentID             │
 ├──────────────────┼─────────────────────────────────────────┼─────────────────────────────────┤
 │ JSON coding      │ .convertFromSnakeCase on shared decoder │ Explicit json:"snake_case" tags │
 ├──────────────────┼─────────────────────────────────────────┼─────────────────────────────────┤
 │ Raw JSON fields  │ JSONValue enum (Codable)                │ json.RawMessage                 │
 ├──────────────────┼─────────────────────────────────────────┼─────────────────────────────────┤
 │ Raw JSON returns │ Data                                    │ json.RawMessage                 │
 ├──────────────────┼─────────────────────────────────────────┼─────────────────────────────────┤
 │ Async            │ async throws                            │ context.Context + error return  │
 ├──────────────────┼─────────────────────────────────────────┼─────────────────────────────────┤
 │ Media create     │ Never (uncallable)                      │ any                             │
 ├──────────────────┼─────────────────────────────────────────┼─────────────────────────────────┤
 │ HTTP client      │ Sendable final class + URLSession       │ struct + *http.Client           │
 ├──────────────────┼─────────────────────────────────────────┼─────────────────────────────────┤
 │ Test mocking     │ URLProtocol subclass                    │ httptest.Server                 │
 ├──────────────────┼─────────────────────────────────────────┼─────────────────────────────────┤
 │ import keyword   │ Property named importResource           │ Import (Go has no conflict)     │
 ├──────────────────┼─────────────────────────────────────────┼─────────────────────────────────┤
 │ URL type         │ URLValue                                │ URL (no conflict in Go)         │
 ├──────────────────┼─────────────────────────────────────────┼─────────────────────────────────┤
 │ Static enum case │ `static` (backtick)                     │ "static" string const           │
 └──────────────────┴─────────────────────────────────────────┴─────────────────────────────────┘

 Critical Reference Files

 - sdks/go/types.go — Primary reference for all entity/param structs (translate 1:1)
 - sdks/go/modulacms.go — Client assembly and resource path mapping
 - sdks/go/resource.go — Generic Resource pattern
 - sdks/go/httpclient.go — HTTP transport pattern
 - sdks/go/ids.go — All ID types and value types
 - .github/workflows/sdks.yml — CI workflow to extend
 - justfile — Build targets to extend

 Verification

 1. cd sdks/swift && swift build — compiles without errors
 2. cd sdks/swift && swift test — all tests pass
 3. just sdk-swift-test — justfile target works
 4. Existing Go and TypeScript SDK tests unaffected
 5. CI workflow validates on push

 Multi-Agent Implementation Order

 Phase 1 — Foundation: Package.swift, IDs.swift, Enums.swift, Errors.swift, Pagination.swift, JSON.swift (sequential, written directly)

 Phase 2 — Parallel:
 - Agent A: HTTPClient.swift
 - Agent B: Types.swift (large, reads Go SDK types.go as reference)

 Phase 3 — After Phase 2: Resource.swift (depends on HTTPClient and entity types)

 Phase 4 — Parallel after Phase 3: All 8 specialized resource files

 Phase 5 — After Phase 4: Client.swift (wires everything)

 Phase 6 — Parallel: All 7 test files

 Phase 7 — Integration: justfile, CI, CLAUDE.md
