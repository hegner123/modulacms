// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "Modula",
    platforms: [.iOS(.v16), .macOS(.v13), .tvOS(.v16), .watchOS(.v9)],
    products: [.library(name: "Modula", targets: ["Modula"])],
    targets: [
        .target(name: "Modula", path: "Sources/Modula"),
        .testTarget(name: "ModulaTests", dependencies: ["Modula"], path: "Tests/ModulaTests"),
    ]
)
