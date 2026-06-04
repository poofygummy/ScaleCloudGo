import ProjectDescription

let project = Project(
    name: "ScaleCloudGo",
    settings: .settings(base: [
        "SKIP_INSTALL": "NO",
        "BUILD_LIBRARY_FOR_DISTRIBUTION": "YES",
        "DEFINES_MODULE": "NO", // Stops Xcode from looking for native headers to build its own modulemap
        "PRODUCT_MODULE_NAME": "ScaleCloudGo"
    ]),
    targets: [
        .target(
            name: "ScaleCloudGo",
            destinations: .iOS,
            product: .framework, // Aligns perfectly with the .framework output below
            bundleId: "com.scalecloud.ScaleCloudGo",
            deploymentTargets: .iOS("14.0"),
            infoPlist: .default,
            sources: [],
            scripts: [
                .pre(
                    script: """
                    export PATH="$PATH:$(go env GOPATH)/bin"
                    # Clear out any empty placeholder directory initialized by the build coordinator
                    rm -rf "$BUILT_PRODUCTS_DIR/ScaleCloudGo.framework"
                    # Compile the Go code directly into the active build products path as a standard framework
                    gomobile bind -target=ios/arm64 -o "$BUILT_PRODUCTS_DIR/ScaleCloudGo.framework" "$PROJECT_DIR"
                    """,
                    name: "Build Go Framework"
                )
            ]
        )
    ]
)
