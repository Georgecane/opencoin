
# Opencoin Blockchain Framework

Opencoin is a professional, open-source blockchain ecosystem and research platform written in Go. It is designed for gas-less, post-quantum secure applications, supporting DPoS consensus, DAG-based state, and WASM smart contracts.

## Architecture Overview

- **Consensus:** Delegated Proof-of-Stake (DPoS) with CometBFT integration for robust validator management and block finality.
- **State Model:** Directed Acyclic Graph (DAG) for scalable, parallel block processing and topological ordering.
- **Smart Contracts:** WASM runtime (wazero) for secure, isolated contract execution.
- **Cryptography:** Post-quantum primitives (Dilithium, Kyber) via reference C implementations and Go CGO bindings.

## Financial Model

- **Gas-less Transactions:** Opencoin is designed to minimize or eliminate transaction fees, enabling frictionless user and contract interactions.
- **DPoS Staking & Rewards:** Validators are elected by token holders, who delegate stake. Block rewards and transaction incentives are distributed to validators and delegators based on performance and stake.
- **Slashing & Security:** Misbehaving validators are penalized via slashing. Economic security is enforced through stake bonding and unbonding periods.
- **Tokenomics:** The system supports flexible token models, including inflationary or capped supply, with governance mechanisms for upgrades and parameter changes.

## Cryptography

- **Dilithium:** Lattice-based digital signature scheme (NIST PQC finalist). Used for block signing, transaction authentication, and validator identity. Integrated via reference C code (`dilithium/ref/`) and Go CGO wrappers (`pkg/crypto`).
- **Kyber:** Lattice-based key encapsulation (KEM) for secure key exchange and contract privacy. Integrated via reference C code (`kyber/ref/`) and Go CGO wrappers.
- **Fallbacks:** When CGO is disabled, the system provides clear error messages and disables PQ features. The `CGOEnabled` constant in `pkg/crypto` allows runtime detection.
- **Security Guidance:** Never store validator keys on insecure devices. Use hardware signers or remote signing infrastructure for production.

## Building & Development

### Quick Build (No Native PQ Crypto)

```bash
env CGO_ENABLED=0 go build ./...
```

### Full Build (With Native Dilithium/Kyber)

Requires GCC/Clang and standard C toolchain.

```bash
export CGO_ENABLED=1
go build ./...
```

### Cross-Compiling for ARM64 (Android/Termux)

```bash
env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o opencoin-arm64 ./cmd/opencoin
# Copy to phone and run in Termux:
# adb push opencoin-arm64 /data/data/com.termux/files/home/
# chmod +x ./opencoin-arm64
# ./opencoin-arm64 start
```

### Editor Configuration

- VS Code: `.vscode/settings.json` sets `CGO_ENABLED=1` for Go tools and disables noisy cgo diagnostics. For other editors, set equivalent build flags.

## Error Handling & Robustness

- All PQ crypto APIs return explicit errors when CGO/native code is unavailable. Check for `errDilithiumCGODisabled` and `errKyberCGODisabled` in your code.
- The system is structured for clear separation of concerns: consensus, state, crypto, contracts, and CLI are modular and testable.
- Fallbacks and runtime flags ensure builds are robust across platforms (desktop, server, phone).

## Security & Production Notes

- Do not run validators on general-purpose phones. Use remote signers, HSMs, or dedicated secure hardware.
- Phones may throttle or overheat under sustained load. Use light-client or wallet mode for mobile devices.
- Always backup keys and use encrypted storage.

## Contributing

- Open issues or PRs for improvements, new PQ primitives, or platform support.
- Document new cryptographic backends and add CI matrix entries for cross-platform builds.

## License

See `LICENSE` for details.

