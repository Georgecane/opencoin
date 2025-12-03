# opencoin

Opencoin is an experimental, open-source blockchain framework implemented in Go.
It is designed as a research / development platform for a gas-less, post-quantum
blockchain using DPoS + CometBFT, a DAG-based state model, and WASM smart
contracts. This repository includes reference C implementations for post-quantum
primitives (Dilithium and Kyber) and Go bindings for development.

This README helps you build the project, explains CGO vs non-CGO behavior,
and gives quick notes about running the node on constrained devices (phones).

Quick build (development)
-------------------------
- Ensure you have Go (recommended 1.24+) installed.
- For a simple, fast build (no CGO native PQ crypto):

```bash
# disable CGO (uses non-CGO fallbacks which currently return explicit errors)
env CGO_ENABLED=0 go build ./...
```

- To build with native PQ crypto (Dilithium & Kyber) you need a C toolchain:

```bash
# enable CGO and build (requires gcc/clang and standard C toolchain)
export CGO_ENABLED=1
go build ./...
```

Notes about CGO and the crypto packages
--------------------------------------
- The repository contains reference implementations in `dilithium/ref/` and
	`kyber/ref/` and CGO wrappers under `pkg/crypto` that compile those C sources
	into the Go package when CGO is enabled.
- When CGO is disabled the package exposes clear fallback errors (`errDilithiumCGODisabled`,
	`errKyberCGODisabled`) so higher layers can detect the absence of native
	implementations and fail gracefully or choose alternate flows (e.g. remote
	signing).
- The package exposes a compile-time constant `pkg/crypto.CGOEnabled` (true/false)
	so code can branch at runtime or inform the user.

Editor (gopls / VS Code) configuration
-------------------------------------
- This workspace contains `.vscode/settings.json` that sets `CGO_ENABLED=1` for
	the Go tools and configures the editor to avoid noisy cgo diagnostics. If
	you use a different editor, add the equivalent `gopls` build flags or
	environment variables to let the language server process cgo preambles.

Running on a phone (Android) — practical notes
---------------------------------------------
- Full node / validator on a phone is generally NOT recommended for production
	(resource, security, and networking constraints). Consider using a phone as
	a light client or remote signer instead.
- Quick experimental path (Android + Termux):

	1. Build a non-CGO ARM64 binary on your development machine:

	```bash
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o opencoin-arm64 ./cmd/opencoin
	```

	2. Copy binary to the phone and run under Termux (or native shell):

	```bash
	adb push opencoin-arm64 /data/data/com.termux/files/home/
	# on the phone (Termux):
	chmod +x ./opencoin-arm64
	./opencoin-arm64 start
	```

- If you require native PQ crypto (Dilithium/Kyber) on the phone you must
	either compile with CGO for the phone's platform (requires Android NDK/toolchain)
	or build on-device with Termux's toolchain. These flows are more advanced.

Security & production guidance
-----------------------------
- Do not store validator private keys on general-purpose phones; use remote
	signers, hardware signers (HSM/YubiKey), or dedicated secure signing
	infrastructure for validators.
- Phones under sustained load will throttle CPU and potentially overheat. A
	charger helps but does not remove thermal and reliability concerns.

Tests and smoke checks
----------------------
- The repo has non-CGO fallbacks so you can build without a C toolchain, but
	PQ crypto tests / demos require CGO. Consider adding CI jobs that run the
	crypto smoke tests on runners with a C toolchain.

Contributing
------------
- Please open issues or PRs for improvements. If you add alternative PQ
	implementations (pure-Go, or different C backends), document them here and
	add CI matrix entries for cross-platform builds.

License
-------
See `LICENSE` for licensing information.

