import ProjectDescription

let project = Project(
    name: "ScaleCloudGo",
    targets: [
        .target(
            name: "ScaleCloudGo",
            destinations: .iOS,
            product: .staticFramework,
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

rm -rf "$TARGET_FRAMEWORK"
rm -rf "$TMP_XCFRAMEWORK"

gomobile bind -target=ios/arm64 -o "$TMP_XCFRAMEWORK" "$PROJECT_DIR"

cp -R "$TMP_XCFRAMEWORK/ios-arm64/ScaleCloudGo.framework" "$TARGET_FRAMEWORK"

MODULEMAP="$TARGET_FRAMEWORK/Modules/module.modulemap"
if [ ! -f "$MODULEMAP" ]; then
    mkdir -p "$TARGET_FRAMEWORK/Modules"
    echo 'framework module ScaleCloudGo {' > "$MODULEMAP"
    echo '  umbrella header "ScaleCloudGo.h"' >> "$MODULEMAP"
    echo '  export *' >> "$MODULEMAP"
    echo '}' >> "$MODULEMAP"
fi
""",
                    name: "Build Go Framework"
                )
            ],
            settings: .settings(base: [
                "SKIP_INSTALL": "NO",
                "BUILD_LIBRARY_FOR_DISTRIBUTION": "YES",
                "DEFINES_MODULE": "YES",
                "PRODUCT_MODULE_NAME": "ScaleCloudGo",
                "DEBUG_INFORMATION_FORMAT": "dwarf",
                "EAGER_LINKING": "NO",
                "SUPPORTS_TEXT_BASED_API": "NO"
            ])
        )
    ]
)
