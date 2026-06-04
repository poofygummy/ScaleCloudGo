import ProjectDescription

let project = Project(
    name: "ScaleCloudGo",
    settings: .settings(base: [
        "SKIP_INSTALL": "NO",
        "BUILD_LIBRARY_FOR_DISTRIBUTION": "YES",
        "DEFINES_MODULE": "YES",
        "PRODUCT_MODULE_NAME": "ScaleCloudGo"
    ]),
    targets: [
        .target(
            name: "ScaleCloudGo",
            destinations: .iOS,
            product: .framework,
            bundleId: "com.scalecloud.ScaleCloudGo",
            deploymentTargets: .iOS("14.0"),
            infoPlist: .default,
            sources: [],
            scripts: [
                .pre(
                    script: """
                    export PATH="$PATH:$(go env GOPATH)/bin"
                    gomobile bind -target=ios/arm64 -o "$BUILT_PRODUCTS_DIR/ScaleCloudGo.xcframework" "$PROJECT_DIR"
                    """,
                    name: "Build Go Framework"
                )
            ]
        )
    ]
)
