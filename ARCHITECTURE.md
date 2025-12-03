# OpenCoin - Gas-less Post-Quantum Blockchain

A delegated proof-of-stake blockchain with post-quantum cryptography (Dilithium) and smart contract support.

## Architecture

```
opencoin/
├── cmd/                    # CLI applications
│   └── opencoin/          # Main node binary
├── pkg/
│   ├── types/            # Core data structures (Block, Tx, DAGVertex)
│   ├── crypto/           # Post-quantum cryptography (Dilithium wrapper)
│   ├── state/            # DAG-based state management
│   ├── consensus/        # DPoS validator logic
│   ├── contracts/        # WASM smart contract engine
│   └── network/          # CometBFT ABCI interface
└── proto/                # Protocol buffer definitions
```

## Core Components

### 1. **DAG State Management** (`pkg/state/dag.go`)
- Maintains a directed acyclic graph of blocks
- Supports multiple parent blocks (unlike traditional linear blockchain)
- Computes topological ordering for deterministic transaction ordering
- Manages account balances and nonces

### 2. **Post-Quantum Cryptography** (`pkg/crypto/dilithium.go`)
- Integrates NIST-standardized Dilithium signature scheme
- Provides `Signer` and `Verifier` interfaces
- TODO: Add CGO bindings to C implementation in `dilithium/`

### 3. **DPoS Consensus** (`pkg/consensus/dpos.go`)
- Validators stake tokens to participate in consensus
- Delegators can delegate stake to validators
- Rewards distribution with configurable commission
- No transaction fees needed (gas-less design)

### 4. **Smart Contracts** (`pkg/contracts/engine.go`)
- WASM bytecode execution
- Isolated contract storage (key-value)
- Gas metering for resource limits
- Contract deployment and invocation
- TODO: Integrate wasmer-go for full WASM execution

### 5. **CometBFT Integration** (`pkg/network/abci.go`)
- Implements ABCI interface for CometBFT consensus
- Handles transaction validation in `CheckTx`
- Applies transactions in `DeliverTx`
- Provides RPC query endpoints

## Implementation Roadmap

### Phase 1: Foundation
- [x] Type definitions
- [x] DAG state structure
- [x] DPoS validator logic
- [x] Smart contract engine skeleton
- [x] ABCI application
- [ ] CGO bindings for Dilithium
- [ ] CometBFT node startup

### Phase 2: Cryptography
- [ ] Compile Dilithium C code with CGO
- [ ] Implement key generation
- [ ] Implement signing/verification
- [ ] Add key serialization

### Phase 3: Smart Contracts
- [ ] Integrate wasmer-go runtime
- [ ] Implement contract storage operations
- [ ] Add gas metering
- [ ] Create contract deployment transaction type
- [ ] Build contract debugging tools

### Phase 4: Networking & Testing
- [ ] Full CometBFT integration
- [ ] Multi-node testing
- [ ] State sync implementation
- [ ] RPC client library
- [ ] Unit tests for all components

### Phase 5: Optimization
- [ ] DAG finality rules
- [ ] Batch transaction verification
- [ ] Contract caching
- [ ] State pruning

## Next Steps

1. **Setup CGO for Dilithium**:
   ```bash
   # Create cgo_dilithium.go wrapper
   # Link to dilithium/ref or dilithium/avx2
   ```

2. **Build CometBFT Node**:
   ```bash
   # Initialize CometBFT home directory
   # Create node.go with NewNode() function
   # Implement node startup
   ```

3. **Add WASM Integration**:
   ```bash
   go get github.com/wasmerio/wasmer-go
   # Implement contract execution in contracts/engine.go
   ```

4. **CLI Functionality**:
   - Implement key generation
   - Add transaction creation/signing
   - Implement RPC queries

## Dependencies

```
CometBFT v0.37.2       - Byzantine Fault Tolerant consensus
WASM (wasmer-go)       - Smart contract execution
Dilithium (C)          - Post-quantum signatures
Protocol Buffers       - Message serialization
```

## Design Decisions

### Why DAG instead of linear blockchain?
- Higher throughput: multiple parents allow parallel blocks
- Better latency: no need to wait for finality between blocks
- Reduced synchronization: validators can propose simultaneously

### Why DPoS?
- Gas-less: no transaction fees required
- Scalable: fixed number of validators
- Democratic: token holders can change validators via delegation

### Why WASM contracts?
- Deterministic: same code = same result everywhere
- Secure: sandboxed execution
- Efficient: compiled bytecode
- Language-agnostic: write in Rust, C, etc.

## Building

```bash
# Requires Go 1.21+
go mod download
go build -o opencoin ./cmd/opencoin
```

## Running a Node

```bash
# Initialize home directory
./opencoin init

# Start the node
./opencoin start
```

## Contributing

- Post-quantum cryptography expertise
- DAG consensus optimization
- Smart contract library development
- Performance optimization
