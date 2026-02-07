package contracts

import (
	"context"
	"fmt"
	"sync"

	"github.com/georgecane/opencoin/pkg/types"
	"github.com/tetratelabs/wazero"
	wazeroapi "github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
)

// ContractEngine manages smart contract execution
type ContractEngine struct {
	mu        sync.RWMutex
	contracts map[string]*Contract
	code      map[string][]byte
	state     map[string]map[string][]byte // contract addr -> key -> value
	// wasm runtime
	ctx      context.Context
	runtime  wazero.Runtime
	compiled map[string]wazero.CompiledModule
	modules  map[string]wazeroapi.Module
	maxCallDepth int
	metrics      map[string]contractMetrics
}

// Contract represents a deployed smart contract
type Contract struct {
	Address  string
	Owner    string
	Code     []byte // WASM bytecode
	Balance  uint64
	Storage  map[string][]byte
	Deployed int64
}

// ExecutionContext provides context for contract execution
type ExecutionContext struct {
	Caller       string
	ContractAddr string
	Value        uint64
	Gas          uint64
	Block        *types.Block
	State        map[string][]byte
}

// ExecutionResult contains execution metadata for RC accounting.
type ExecutionResult struct {
	Output       []byte
	Instructions uint64
	StateWrites  uint64
}

// NewContractEngine creates a new contract engine
func NewContractEngine() *ContractEngine {
	ctx := context.Background()
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithMemoryLimitPages(1024))
	return &ContractEngine{
		contracts: make(map[string]*Contract),
		code:      make(map[string][]byte),
		state:     make(map[string]map[string][]byte),
		ctx:       ctx,
		runtime:   r,
		modules:   make(map[string]wazeroapi.Module),
		maxCallDepth: 32,
		metrics:  make(map[string]contractMetrics),
	}
}

// DeployContract deploys a new WASM contract
func (ce *ContractEngine) DeployContract(owner string, wasmCode []byte, address string) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	if _, exists := ce.contracts[address]; exists {
		return fmt.Errorf("contract already exists at address: %s", address)
	}

	if len(wasmCode) == 0 {
		return fmt.Errorf("empty contract code")
	}

	contract := &Contract{
		Address: address,
		Owner:   owner,
		Code:    wasmCode,
		Balance: 0,
		Storage: make(map[string][]byte),
	}

	ce.contracts[address] = contract
	ce.code[address] = wasmCode
	ce.state[address] = make(map[string][]byte)

	if err := ValidateWasmCode(wasmCode); err != nil {
		return err
	}

	ce.metrics[address] = contractMetrics{
		InstructionEstimate: estimateInstructions(wasmCode),
		StateWriteEstimate:  1,
	}

	// Try to compile/instantiate the module so errors surface early
	if len(wasmCode) > 0 {
		compiled, err := ce.runtime.CompileModule(ce.withCallDepthListener(), wasmCode)
		if err != nil {
			return fmt.Errorf("failed to compile wasm module: %w", err)
		}
		mod, err := ce.runtime.InstantiateModule(ce.withCallDepthListener(), compiled, wazero.NewModuleConfig())
		if err != nil {
			return fmt.Errorf("failed to instantiate wasm module: %w", err)
		}
		ce.compiled[address] = compiled
		ce.modules[address] = mod
	}

	return nil
}

// ExecuteContract executes a contract call
func (ce *ContractEngine) ExecuteContract(ctx *ExecutionContext) ([]byte, error) {
	ce.mu.Lock()
	contract, exists := ce.contracts[ctx.ContractAddr]
	ce.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("contract not found: %s", ctx.ContractAddr)
	}

	if len(contract.Code) == 0 {
		return nil, fmt.Errorf("contract has no code")
	}

	// Execute the exported `handle` function (no params) if available.
	// Use previously instantiated module if present.
	ce.mu.Lock()
	mod, ok := ce.modules[ctx.ContractAddr]
	ce.mu.Unlock()

	if !ok {
		// compile and instantiate on demand
		compiled, err := ce.runtime.CompileModule(ce.withCallDepthListener(), contract.Code)
		if err != nil {
			return nil, fmt.Errorf("failed to compile wasm module: %w", err)
		}
		m, err := ce.runtime.InstantiateModule(ce.withCallDepthListener(), compiled, wazero.NewModuleConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate wasm module: %w", err)
		}
		mod = m
		ce.mu.Lock()
		ce.compiled[ctx.ContractAddr] = compiled
		ce.modules[ctx.ContractAddr] = mod
		ce.mu.Unlock()
	}

	// Look up `handle` export
	fn := mod.ExportedFunction("handle")
	if fn == nil {
		// no handle function; nothing to do
		return nil, nil
	}

	// Call the function with no parameters
	var results []uint64
	var callErr error
	defer func() {
		if r := recover(); r != nil {
			callErr = fmt.Errorf("wasm trap: %v", r)
		}
	}()
	results, callErr = fn.Call(ce.withCallDepthListener())
	if callErr != nil {
		return nil, fmt.Errorf("wasm contract handle failed: %w", callErr)
	}

	// Return raw results as bytes (if any uint64 results exist, encode them)
	if len(results) > 0 {
		// convert first uint64 result to bytes (little-endian)
		v := results[0]
		b := []byte{
			byte(v),
			byte(v >> 8),
			byte(v >> 16),
			byte(v >> 24),
			byte(v >> 32),
			byte(v >> 40),
			byte(v >> 48),
			byte(v >> 56),
		}
		return b, nil
	}

	return nil, nil
}

// ExecuteContract returns execution result with instruction and state write estimates.
func (ce *ContractEngine) ExecuteContractWithResult(ctx *ExecutionContext) (*ExecutionResult, error) {
	output, err := ce.ExecuteContract(ctx)
	if err != nil {
		return nil, err
	}
	ce.mu.RLock()
	metrics := ce.metrics[ctx.ContractAddr]
	ce.mu.RUnlock()
	return &ExecutionResult{
		Output:       output,
		Instructions: metrics.InstructionEstimate,
		StateWrites:  metrics.StateWriteEstimate,
	}, nil
}

// GetContract returns a contract by address
func (ce *ContractEngine) GetContract(address string) *Contract {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	if contract, exists := ce.contracts[address]; exists {
		contractCopy := *contract
		return &contractCopy
	}
	return nil
}

// SetContractStorage sets a contract's storage value
func (ce *ContractEngine) SetContractStorage(contractAddr string, key, value []byte) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	storage, exists := ce.state[contractAddr]
	if !exists {
		return fmt.Errorf("contract not found: %s", contractAddr)
	}

	storage[string(key)] = value
	return nil
}

// GetContractStorage gets a contract's storage value
func (ce *ContractEngine) GetContractStorage(contractAddr string, key []byte) ([]byte, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	storage, exists := ce.state[contractAddr]
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddr)
	}

	if value, ok := storage[string(key)]; ok {
		return value, nil
	}

	return nil, nil
}

// TransferToContract transfers funds to a contract
func (ce *ContractEngine) TransferToContract(contractAddr string, amount uint64) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	contract, exists := ce.contracts[contractAddr]
	if !exists {
		return fmt.Errorf("contract not found: %s", contractAddr)
	}

	contract.Balance += amount
	return nil
}

// GetContractBalance returns the balance of a contract
func (ce *ContractEngine) GetContractBalance(contractAddr string) (uint64, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	contract, exists := ce.contracts[contractAddr]
	if !exists {
		return 0, fmt.Errorf("contract not found: %s", contractAddr)
	}

	return contract.Balance, nil
}

// ValidateWasmCode validates WASM bytecode
func ValidateWasmCode(code []byte) error {
	if len(code) == 0 {
		return fmt.Errorf("empty code")
	}

	// WASM magic number: 0x00 0x61 0x73 0x6d
	if len(code) < 4 || code[0] != 0x00 || code[1] != 0x61 || code[2] != 0x73 || code[3] != 0x6d {
		return fmt.Errorf("invalid WASM magic number")
	}

	// Reject floating-point opcodes deterministically.
	if containsFloatOpcodes(code) {
		return fmt.Errorf("wasm contains floating-point opcodes")
	}
	return nil
}

type contractMetrics struct {
	InstructionEstimate uint64
	StateWriteEstimate  uint64
}

// EstimateInstructions returns a deterministic instruction estimate for a WASM module.
func (ce *ContractEngine) EstimateInstructions(wasmCode []byte) uint64 {
	return estimateInstructions(wasmCode)
}

// EstimateContractCall returns a cached instruction estimate.
func (ce *ContractEngine) EstimateContractCall(address string) uint64 {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.metrics[address].InstructionEstimate
}

// EstimateStateWrites returns a cached write estimate.
func (ce *ContractEngine) EstimateStateWrites(address string) uint64 {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.metrics[address].StateWriteEstimate
}

func (ce *ContractEngine) withCallDepthListener() context.Context {
	factory := experimental.FunctionListenerFactoryFunc(func(def wazeroapi.FunctionDefinition) experimental.FunctionListener {
		return experimental.FunctionListenerFunc(func(ctx context.Context, mod wazeroapi.Module, def wazeroapi.FunctionDefinition, params []uint64, stack experimental.StackIterator) {
			depth := 0
			for stack.Next() {
				depth++
			}
			if depth > ce.maxCallDepth {
				panic(fmt.Errorf("wasm max call depth exceeded: %d", depth))
			}
		})
	})
	return experimental.WithFunctionListenerFactory(ce.ctx, factory)
}

func estimateInstructions(wasmCode []byte) uint64 {
	// Deterministic upper-bound estimate based on code section size.
	return uint64(len(wasmCode))
}

// containsFloatOpcodes scans the code section for float opcode bytes.
// This is a conservative check that may reject some valid code, but is deterministic.
func containsFloatOpcodes(wasmCode []byte) bool {
	// float-related opcodes in WASM (subset)
	floatOpcodes := map[byte]struct{}{
		0x43: {}, // f32.const
		0x44: {}, // f64.const
		0x8b: {}, // f32.add
		0x8c: {}, // f32.sub
		0x8d: {}, // f32.mul
		0x8e: {}, // f32.div
		0x99: {}, // f64.add
		0x9a: {}, // f64.sub
		0x9b: {}, // f64.mul
		0x9c: {}, // f64.div
	}
	for _, b := range wasmCode {
		if _, ok := floatOpcodes[b]; ok {
			return true
		}
	}
	return false
}
