// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "ModulaCMS",
    platforms: [.iOS(.v16), .macOS(.v13), .tvOS(.v16), .watchOS(.v9)],
    products: [.library(name: "ModulaCMS", targets: ["ModulaCMS"])],
    targets: [
        .target(name: "ModulaCMS", path: "Sources/ModulaCMS"),
        .testTarget(name: "ModulaCMSTests", dependencies: ["ModulaCMS"], path: "Tests/ModulaCMSTests"),
    ]
)
