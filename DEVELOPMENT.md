# OpenCoin Development Guide

## Project Overview

OpenCoin is a next-generation blockchain that combines three key innovations:

1. **Post-Quantum Security** - Uses NIST-standardized Dilithium for quantum-resistant signatures
2. **Gas-less Design** - DPoS consensus eliminates transaction fees
3. **DAG Architecture** - Enables higher throughput through parallel block production
4. **Smart Contracts** - WASM-based programmable transactions

## Prerequisites

- Go 1.21 or later
- GCC/Clang (for compiling Dilithium C code)
- CMake 3.10+ (optional, for building Dilithium)

## Project Structure

```
opencoin/
├── cmd/                          # Executables
│   └── opencoin/
│       ├── main.go              # Entry point
│       └── cmd/                 # CLI command definitions
│
├── pkg/                          # Core libraries
│   ├── types/                   # Data structures
│   │   └── types.go            # Block, Tx, DAGVertex, etc.
│   │
│   ├── crypto/                  # Post-quantum cryptography
│   │   ├── dilithium.go        # Dilithium wrapper (TODO: CGO)
│   │   └── dilithium_cgo.go    # CGO bindings (TODO)
│   │
│   ├── state/                   # DAG state management
│   │   ├── dag.go              # DAG structure and operations
│   │   └── merkle.go           # Merkle tree for state roots (TODO)
│   │
│   ├── consensus/               # DPoS implementation
│   │   ├── dpos.go             # Validator and delegation logic
│   │   └── rewards.go          # Reward distribution (TODO)
│   │
│   ├── contracts/               # Smart contract engine
│   │   ├── engine.go           # Contract execution
│   │   └── runtime.go          # WASM runtime (TODO)
│   │
│   ├── network/                 # Network and consensus
│   │   ├── abci.go             # CometBFT ABCI interface
│   │   └── node.go             # Node implementation (TODO)
│   │
│   └── config/                  # Configuration management
│       └── config.go           # Config structures
│
├── dilithium/                   # Post-quantum crypto reference impl
│   ├── ref/                    # Reference implementation
│   └── avx2/                   # AVX2-optimized implementation
│
├── proto/                       # Protocol buffer definitions
│   └── opencoin.proto          # Message definitions
│
├── go.mod                       # Go module definition
├── go.sum                       # Go dependencies checksums
└── ARCHITECTURE.md              # Architecture documentation
```

## Key Implementation Areas

### 1. Dilithium Integration (Crypto Layer)

**Current Status**: TODO - CGO bindings needed

**What to do**:
```go
// Create pkg/crypto/dilithium_cgo.go with:
// #cgo CFLAGS: -I${SRCDIR}/../../dilithium/ref
// #cgo LDFLAGS: -lm
// #include "sign.h"

// Implement:
// - crypto_sign_keypair() -> go KeyPair
// - crypto_sign() -> go signature
// - crypto_sign_open() -> go verification
```

**Files involved**:
- `pkg/crypto/dilithium.go` - Go wrapper (skeleton exists)
- `dilithium/ref/` - C implementation to link
- Need to create `dilithium/ref/cgo_build.go` for compilation hints

### 2. DAG State Management

**Current Status**: DONE - Core structure implemented

**Files**:
- `pkg/state/dag.go` - Complete DAG operations
- `pkg/types/types.go` - Data structures

**Features**:
- [x] Add blocks with multiple parents
- [x] Maintain frontier (tips)
- [x] Topological ordering
- [x] Account state management
- [ ] State root computation (Merkle tree)
- [ ] Block finality rules

**Next steps**:
```go
// Implement merkle tree for state roots
// Add finality rules:
// - Blocks are final after confirmation from 2/3 of validators
// - Or after X blocks on top of them
```

### 3. DPoS Consensus

**Current Status**: DONE - Core logic implemented

**Files**:
- `pkg/consensus/dpos.go` - Validator registration and delegation

**Features**:
- [x] Validator registration with minimum stake
- [x] Delegation and undelegation
- [x] Validator set management
- [x] Reward distribution skeleton
- [ ] Slashing for Byzantine behavior
- [ ] Unbonding period enforcement
- [ ] Delegation history tracking

**Next steps**:
```go
// In dpos.go, implement:
// - Slashing penalties for double-signing
// - Unbonding queue for delegations
// - Historical reward tracking
// - Validator jailing for downtime
```

### 4. Smart Contract Engine

**Current Status**: Skeleton - Ready for WASM integration

**Files**:
- `pkg/contracts/engine.go` - Contract manager

**What to implement**:
```bash
go get github.com/wasmerio/wasmer-go

# Then in contracts/engine.go:
# - Use wasmer.NewInstance() to load WASM
# - Implement memory isolation per contract
# - Add gas metering via instruction hooks
# - Implement contract storage ops via WASM imports
```

**Contract Interface (WASM side)**:
```rust
// Contracts will export these functions:
#[no_mangle]
pub extern "C" fn init(params: *const u8, len: usize) -> u32 { ... }

#[no_mangle]
pub extern "C" fn handle(msg: *const u8, len: usize) -> u32 { ... }

#[no_mangle]
pub extern "C" fn query(msg: *const u8, len: usize) -> u32 { ... }
```

**Imports (provided by host)**:
```rust
extern "C" {
    fn storage_get(key: *const u8, key_len: usize, out: *mut u8, out_cap: usize) -> usize;
    fn storage_set(key: *const u8, key_len: usize, val: *const u8, val_len: usize);
    fn balance(addr: *const u8, addr_len: usize) -> u64;
    fn emit_log(msg: *const u8, len: usize);
}
```

### 5. CometBFT Integration

**Current Status**: ABCI skeleton - Ready for node implementation

**Files**:
- `pkg/network/abci.go` - ABCI application (complete)
- `pkg/network/node.go` - TODO: Node startup

**What to implement in node.go**:
```go
// Initialize CometBFT node:
// - Load config
// - Create privval signer
// - Create node with NewNode()
// - Start socket RPC server
// - Connect to peers
// - Run forever

func NewNode(config *NodeConfig) (*cmtnode.Node, error) {
    // 1. Load or generate node keys
    // 2. Create application
    // 3. Initialize database
    // 4. Create CometBFT node
    // 5. Return node
}
```

## Building the Project

### Step 1: Setup Go Module
```bash
cd /home/george/Documents/Projects/opencoin
go mod tidy
```

### Step 2: Build Dilithium C Library (CGO)
```bash
# This should happen automatically via CGO when you compile
# But first, create the CGO wrapper:
cat > pkg/crypto/dilithium_cgo.go << 'EOF'
//go:build cgo
package crypto

// #cgo CFLAGS: -I${SRCDIR}/../../dilithium/ref
// #include "sign.h"
// #include <string.h>
// #include <stdlib.h>
//
// int dilithium_generate_keypair(unsigned char *pk, unsigned char *sk) {
//     return crypto_sign_keypair(pk, sk);
// }
//
// int dilithium_sign(unsigned char *sig, unsigned long long *siglen,
//                    const unsigned char *m, unsigned long long mlen,
//                    const unsigned char *sk) {
//     return crypto_sign(sig, siglen, m, mlen, sk);
// }
//
// int dilithium_verify(const unsigned char *sig, unsigned long long siglen,
//                      const unsigned char *m, unsigned long long mlen,
//                      const unsigned char *pk) {
//     return crypto_sign_open(NULL, NULL, sig, siglen, m, mlen, pk);
// }
import "C"

// GenerateCGoKeyPair generates a key pair using C implementation
func GenerateCGoKeyPair() (*KeyPair, error) {
    pk := make([]byte, 1312)
    sk := make([]byte, 2544)
    
    if ret := C.dilithium_generate_keypair((*C.uchar)(&pk[0]), (*C.uchar)(&sk[0])); ret != 0 {
        return nil, fmt.Errorf("key generation failed: %d", ret)
    }
    
    return &KeyPair{PublicKey: pk, PrivateKey: sk}, nil
}

// CGoSign signs using C implementation
func (ds *DilithiumSigner) CGoSign(message []byte) ([]byte, error) {
    sig := make([]byte, 2420)
    var siglen C.ulonglong
    
    if ret := C.dilithium_sign(
        (*C.uchar)(&sig[0]),
        &siglen,
        (*C.uchar)(&message[0]),
        C.ulonglong(len(message)),
        (*C.uchar)(&ds.privateKey[0]),
    ); ret != 0 {
        return nil, fmt.Errorf("signing failed: %d", ret)
    }
    
    return sig[:siglen], nil
}
EOF
```

### Step 3: Compile
```bash
go build -o opencoin ./cmd/opencoin
```

### Step 4: Generate Protobuf Code (Optional)
```bash
# Install protoc compiler
go install github.com/grpc/grpc-go/cmd/protoc-gen-go-grpc@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Generate Go code from .proto files
protoc --go_out=. --go-grpc_out=. proto/opencoin.proto
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./pkg/state/...
go test ./pkg/consensus/...

# Run with coverage
go test -cover ./...
```

## Testing the Blockchain Locally

### Start a single node:
```bash
# Initialize
./opencoin init --moniker node1

# Start node
./opencoin start
```

### Deploy a smart contract:
```bash
# Compile WASM contract (Rust example):
# wasm-pack build --target wasm32-unknown-unknown

# Deploy contract
./opencoin tx contract deploy path/to/contract.wasm
```

### Submit transactions:
```bash
# Create transfer transaction
./opencoin tx transfer addr_to 1000 --from addr_from

# Check account balance
./opencoin query account addr_from
```

## Performance Considerations

### DAG Optimization
- Limit parents per block to 2-4 for better propagation
- Implement phantom protocol for conflict resolution
- Consider speculative block execution for parallelism

### Consensus Tuning
- Adjust timeout parameters based on network latency
- Increase max validators for decentralization (with performance trade-off)
- Implement validator rotation for security

### Contract Optimization
- Cache compiled WASM modules
- Implement contract state snapshots
- Batch contract calls when possible

## Security Checklist

- [ ] Dilithium signatures verified on all blocks
- [ ] Nonce checking to prevent replay attacks
- [ ] Balance checking to prevent double-spending
- [ ] Contract memory isolation enforced
- [ ] Gas metering prevents infinite loops
- [ ] Stake slashing for Byzantine validators
- [ ] Unbonding period enforces exit delays

## Contributing Guidelines

1. **Format**: Use `gofmt` and `golint`
2. **Tests**: All features must have unit tests
3. **Docs**: Document public APIs with comments
4. **Commits**: Use clear, descriptive commit messages
5. **PRs**: Include test coverage improvements

## Resources

- [Dilithium Specification](https://pq-crystals.org/dilithium/)
- [CometBFT ABCI](https://docs.cometbft.com/main/spec/abci/)
- [WASM Specification](https://webassembly.org/)
- [DPoS Overview](https://en.wikipedia.org/wiki/Proof_of_stake#Delegated_proof_of_stake)

## Troubleshooting

### CGO Build Issues
```bash
# Set CGO_ENABLED explicitly
export CGO_ENABLED=1
export CC=gcc  # or clang

go build ./cmd/opencoin
```

### Import Errors
```bash
# Ensure proto files are generated
go generate ./...

# Regenerate go.mod
go mod tidy
```

### Test Failures
```bash
# Check dependencies are installed
go mod download

# Run verbose tests
go test -v ./...
```
