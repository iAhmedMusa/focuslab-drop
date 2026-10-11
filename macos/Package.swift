// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "FocusLabDrop",
    platforms: [.macOS(.v13)],
    targets: [
        // Pure logic (addresses, URLs, server arguments, port check) so it can be unit tested.
        .target(name: "FocusLabCore", path: "Sources/FocusLabCore"),
        .executableTarget(name: "FocusLabDrop", dependencies: ["FocusLabCore"], path: "Sources/FocusLabDrop"),
        .testTarget(name: "FocusLabCoreTests", dependencies: ["FocusLabCore"], path: "Tests/FocusLabCoreTests"),
    ]
)
