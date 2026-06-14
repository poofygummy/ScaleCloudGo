# ScaleCloudGo Prebuilt

This directory holds the Go/Tailscale xcframework (or .framework) produced by the `gomobile bind` step in the ScaleCloudGo xcodegen project.

## Expected layout after a GitHub Actions download or git checkout

```
ScaleCloudGo/prebuilt/
└── ScaleCloudGo.xcframework/   (or ScaleCloudGo.framework)
    └── ...
```

## CI usage (independent layer build)

- Use workflow: **Build ScaleCloudGo** (testbuildSCGo.yml)
- On success it always uploads:
  - `ScaleCloudGo-xcarchive`
  - `ScaleCloudGo-prebuilt` (clean tree ready for Kit/App consumption)
- To build *only* Kit/App/Wrap without re-running Go+gomobile:
  - Download the `ScaleCloudGo-prebuilt` artifact from that successful run in the GitHub UI.
  - Manually unpack it into your clone so the contents land under `ScaleCloudGo/prebuilt/` (you should end up with `ScaleCloudGo/prebuilt/ScaleCloudGo.xcframework` or `.framework`).
  - Then dispatch the higher layer's workflow (`testbuildSCKit.yml`, etc.). The higher workflow will only check that the prebuilt tree is already there; it will not download anything itself.
- The xcodegen spec and any build-phase logic in a higher job simply look for the framework under `prebuilt/`. The human is responsible for making sure it is present in the tree at dispatch time.

**There is no supported local build for this layer.**

The only place `ScaleCloudGo` (and its `gomobile bind` step) is built is inside the official **Build ScaleCloudGo** GitHub Actions workflow (`testbuildSCGo.yml`).

To provide a prebuilt for a Kit / App / Wrap dispatch:
- Dispatch the Go workflow.
- After success, download the `ScaleCloudGo-prebuilt` artifact from the GitHub UI (Actions → the run → Artifacts).
- Unpack the artifact contents by hand (unzip/copy) into your clone so they land correctly under `ScaleCloudGo/prebuilt/`.
- Then dispatch the next layer's workflow. The higher workflow sees the prebuilt you placed there and skips Go work entirely.

The Go workflow is the only workflow permitted to install the Go toolchain and gomobile. Any higher workflow that finds a prebuilt already present in the tree must not pay that cost.

## Used by

- ScaleCloudKit (via projectReference to ScaleCloudGo/ScaleCloudGo.xcodeproj + explicit prebuilt path resolution)
- Transitively ScaleCloudApp, ScaleCloudWrap

## Max upstream compatibility note

Nothing in this layer is inherited from Nextcloud repos; this is the private Tailscale/Go mobility shim. Changes here only affect the Go contract (symbols exported to the Swift consumer).
