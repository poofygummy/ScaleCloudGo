import ProjectDescription

let project = Project(
    name: "ScaleCloudGo",
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
                .post(
                    script: """
                    export PATH="$PATH:$(go env GOPATH)/bin"
                    
                    TARGET_FRAMEWORK="$BUILT_PRODUCTS_DIR/ScaleCloudGo.framework"
                    TMP_XCFRAMEWORK="$TEMP_DIR/ScaleCloudGo.xcframework"
                    
                    # Wipe any stale files or Xcode placeholders
                    rm -rf "$TARGET_FRAMEWORK"
                    rm -rf "$TMP_XCFRAMEWORK"
                    
                    # Force Go to generate the xcframework package structure
                    gomobile bind -target=ios/arm64 -o "$TMP_XCFRAMEWORK" "$PROJECT_DIR"
                    
                    # Extract the standalone arm64 framework bundle directly into Xcode's target product path
                    cp -R "$TMP_XCFRAMEWORK/ios-arm64/ScaleCloudGo.framework" "$TARGET_FRAMEWORK"
                    """,
                    name: "Build Go Framework"
                )
            ],
            settings: .settings(base: [
                "SKIP_INSTALL": "NO",
                "BUILD_LIBRARY_FOR_DISTRIBUTION": "YES",
                "DEFINES_MODULE": "NO",
                "PRODUCT_MODULE_NAME": "ScaleCloudGo",
                "DEBUG_INFORMATION_FORMAT": "dwarf",
                "EAGER_LINKING": "NO"
            ])
        )
    ]
)
