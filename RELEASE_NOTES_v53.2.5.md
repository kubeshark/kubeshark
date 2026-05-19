# Kubeshark release v53.2.5 (05-01-2026)

## Release Highlights

Kubeshark 53.2.5 adds native **MySQL** and **PostgreSQL** protocol dissection — both enabled out of the box — fixes server-side TLS decryption for pre-fork servers like PostgreSQL, and introduces a configurable dashboard entries limit. The release also includes stream-history playback, a 4× heap memory reduction in the dashboard, and several reliability fixes across the stack.

## New Features

- **MySQL protocol dissector** — Full wire-protocol dissection for MySQL traffic on port `3306`, enabled by default. Includes worker dissector, hub dependency update, front-end UI support, and e2e tests (worker#1126, hub#745, front#1196, registry-go-tls#72)
- **PostgreSQL protocol dissector** — Full wire-protocol dissection for PostgreSQL traffic on port `5432`, enabled by default. Includes worker dissector, hub dependency update, front-end UI support, and e2e tests (worker#1129, hub#746, front#1200, registry-go-tls#74)
- **Stream history** — New API and UI for streaming historical (raw capture) traffic, enabling retrospective browsing of previously captured data with a time-picker range control (hub#722, worker#1110, front#1173)
- **Dashboard entries limit** — New `tap.dashboard.entriesLimit` Helm value (default `300000`) controls the maximum number of entries the dashboard holds in memory

## Improvements

- **Columnar entry storage — 4× heap memory reduction** — Replaces sparse per-entry maps with dictionary-backed columnar storage, dramatically reducing browser memory usage (front#1161)
- **Decode multi-message gRPC payloads** — Request and response views now decode concatenated gRPC frames instead of showing raw bytes (front#1197)
- **Surface `grpc_method` / `grpc_status` as KFL queries in UI** — Clicking the Path row on gRPC entries emits `grpc_method == "..."` and clicking Grpc-Status emits `grpc_status == N` (front#1174)
- **Show both entry namespaces by default** — Source and destination namespaces are now always visible in the entry list (front#1178)
- **Copy button for snapshot failure reason** — Adds a clipboard copy button to snapshot error messages for easier debugging (front#1179)
- **Document gRPC KFL fields in MCP schema** — `grpc`, `grpc_method`, `grpc_status` are now documented in the MCP KFL schema (hub#726)
- **Adjust KFL input boxes** — Visual refinements to KFL filter input areas (front#1203)
- **Release pipeline overhaul** — Split the monolithic `release-pr` Makefile target into three independent, idempotent targets (`release-siblings`, `release-pr-kubeshark`, `release-pr-helm`) that can be rerun individually without side effects
- **Fix Chart.yaml sync to kubeshark.github.io** — The helm PR target was switching back to master before copying the chart, shipping the pre-bump version

## Bug Fixes

- **Fix server-side TLS decryption for pre-fork servers (PostgreSQL)** — Adds dual-key `connection_context` and accept-symbol fallback for servers that fork before accepting connections (tracer2#54)
- **Fix TLS stop-capturing** — Prevent closing uprobe hooks for TLS workloads that are still being traced (tracer2#56)
- **Fix Istio one-leg HTTP** — Corrects single-leg HTTP capture in Istio service mesh environments (tracer2#55)
- **Extract Go TLS offsets from stripped binaries at runtime** — Enables Go TLS decryption for binaries without debug symbols by extracting offsets at runtime (tracer2#49)
- **Fix HTTP api-server one-leg** — Resolves single-leg API server traffic capture (worker#1138)
- **Remove sliding-window TLS heuristic from all L7 dissectors** — Eliminates false-positive TLS detection that could misclassify plaintext traffic (worker#1137)
- **Fix request/response matcher in MongoDB and Kafka** — Corrects pairing logic in the MongoDB and Kafka protocol dissectors (worker#1114)
- **Fix Pebble use-after-close** — Resolves a crash from accessing Pebble DB after it has been closed (worker#1114)
- **Fix re-running dissection on failure** — Snapshot dissection can now be retried after a failure without manual intervention (front#1192)
- **Skip incomplete dissections from cloud upload** — Prevents partially dissected snapshots from being uploaded to cloud storage (hub#739)
- **Ensure auth credentials on all API requests** — Mirrors auth headers to cookies for consistent authentication across all request types (hub#750, front#1205)

## Infrastructure & Dependencies

- **Revert time-boundaries display above snapshots table** — Reverted pending a redesign (front#1202)
- **Update Go to 1.26.2** (worker#1116)
- **Update build environment** to latest version (front#1180)
- **Bump deps to close Scout CVEs** — Updates moby v2, go-jose, jsonparser, and OpenTelemetry dependencies (worker#1117)
- **Update spdystream** — Addresses CVEs (worker#1118)
- **Bump KFL2 and lock in gRPC trailer merge** (worker#1112)
- **Drop arm64 race builds** due to toolchain limitations (worker#1125)
- **Update registry offsets** — Refreshes embedded Go TLS offset bundle (worker#1124)
- **Reduce noisy parse-packet log messages** (worker#1114)
- **Update gRPC in e2e tests** to address Dependabot issues (hub#729)
- **Update hub dependencies** for Docker Scout compliance (hub)

### registry-go-tls

- Add bifrost-1.4.23 via go-stripped strategy
- Add bifrost-1.4.24 via go-stripped strategy
- Add envoy-1.38.0
- Add PostgreSQL and MySQL test infrastructure and e2e tests
- Add Ollama e2e test infrastructure with TLS support
- Istio: require cross-node placement and detect missing ambient components

### console-v3

- Update Pro plan pricing
- Update docs-agent knowledge base for v53.2.0

## Download Kubeshark for your platform

**Mac** (x86-64/Intel)
```
curl -Lo kubeshark https://github.com/kubeshark/kubeshark/releases/download/v53.2.5/kubeshark_darwin_amd64 && chmod 755 kubeshark
```

**Mac** (AArch64/Apple M1 silicon)
```
curl -Lo kubeshark https://github.com/kubeshark/kubeshark/releases/download/v53.2.5/kubeshark_darwin_arm64 && chmod 755 kubeshark
```

**Linux** (x86-64)
```
curl -Lo kubeshark https://github.com/kubeshark/kubeshark/releases/download/v53.2.5/kubeshark_linux_amd64 && chmod 755 kubeshark
```

**Linux** (AArch64)
```
curl -Lo kubeshark https://github.com/kubeshark/kubeshark/releases/download/v53.2.5/kubeshark_linux_arm64 && chmod 755 kubeshark
```

**Windows** (x86-64)
```
curl -LO https://github.com/kubeshark/kubeshark/releases/download/v53.2.5/kubeshark.exe
```

### Checksums
SHA256 checksums available for compiled binaries.
Run `shasum -a 256 -c kubeshark_OS_ARCH.sha256` to verify.
