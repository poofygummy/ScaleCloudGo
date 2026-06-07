# ScaleCloudGo Prebuilt Framework

This directory contains the precompiled ScaleCloudGo.xcframework built with `gomobile bind`.

## Contents

```
ScaleCloudGo.xcframework/
├── Info.plist
├── ios-arm64/
│   └── ScaleCloudGo.framework/
└── ios-arm64_x86_64-simulator/
    └── ScaleCloudGo.framework/
```

## How to Build

Run the GitHub Actions workflow:
- Workflow: `testbuild.yml`
- Layer: `go`

Or build locally:
```bash
cd ScaleCloudGo
gomobile bind -target=ios -o prebuilt/ScaleCloudGo.xcframework .
```

## Used By

- ScaleCloudKit (depends on this framework)
- All higher layers transitively

## When to Rebuild

Rebuild when you modify:
- `ScaleCloudGo.go`
- `go.mod` / `go.sum`
- Any Go source files

After rebuilding, **you must also rebuild**:
- ScaleCloudKit
- ScaleCloudApp  
- ScaleCloudWrap
