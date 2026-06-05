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

# Create Headers directory and umbrella header
mkdir -p "$TARGET_FRAMEWORK/Headers"
UMBRELLA="$TARGET_FRAMEWORK/Headers/ScaleCloudGo.h"

# Create a minimal umbrella header (gomobile sometimes puts the header at framework root)
if [ -f "$TARGET_FRAMEWORK/ScaleCloudGo.h" ]; then
    cp "$TARGET_FRAMEWORK/ScaleCloudGo.h" "$UMBRELLA"
else
    echo '// ScaleCloudGo umbrella header' > "$UMBRELLA"
fi

# Create module map
mkdir -p "$TARGET_FRAMEWORK/Modules"
cat > "$TARGET_FRAMEWORK/Modules/module.modulemap" << EOF
framework module ScaleCloudGo {
    umbrella header "ScaleCloudGo.h"
    export *
}
EOF
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
