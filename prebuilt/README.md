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
  - In the target workflow (SCKit/SCApp/SCWrap) supply the prior `go_run_id` as a `workflow_dispatch` input.
  - The workflow does `download-artifact` of `ScaleCloudGo-prebuilt` and unpacks it verbatim under `ScaleCloudGo/prebuilt/`.
- The xcodegen spec (`project.yml`) used inside the GitHub job, or any gomobile command executed inside that same job, will look under `prebuilt/` first. The job is what materializes the prebuilt (via artifact download when a prior run id was supplied) before it starts its own work.

**There is no supported local build for this layer.**

The only place `ScaleCloudGo` (and its `gomobile bind` step) is built is inside the official **Build ScaleCloudGo** GitHub Actions workflow (`testbuildSCGo.yml`).

To provide a prebuilt for a Kit / App / Wrap dispatch:
- Dispatch the Go workflow (or use the run id of a prior successful run).
- Download the `ScaleCloudGo-prebuilt` artifact from that run.
- Unpack it so its contents land under `ScaleCloudGo/prebuilt/` in a clone (or let the consuming layer's workflow do the `actions/download-artifact + materialize` step for you).
- Then dispatch the next layer's workflow (supplying the `go_run_id` if the higher workflow supports it).

The Go workflow is the only workflow permitted to install the Go toolchain and gomobile. Higher workflows that receive a valid prior Go prebuilt must not pay that cost.

## Used by

- ScaleCloudKit (via projectReference to ScaleCloudGo/ScaleCloudGo.xcodeproj + explicit prebuilt path resolution)
- Transitively ScaleCloudApp, ScaleCloudWrap

## Max upstream compatibility note

Nothing in this layer is inherited from Nextcloud repos; this is the private Tailscale/Go mobility shim. Changes here only affect the Go contract (symbols exported to the Swift consumer).
