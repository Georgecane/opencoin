package contracts

import (
	"context"
	"fmt"
	"sync"

	"github.com/georgecane/opencoin/pkg/types"
	"github.com/tetratelabs/wazero"
	wazeroapi "github.com/tetratelabs/wazero/api"
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

// NewContractEngine creates a new contract engine
func NewContractEngine() *ContractEngine {
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	return &ContractEngine{
		contracts: make(map[string]*Contract),
		code:      make(map[string][]byte),
		state:     make(map[string]map[string][]byte),
		ctx:       ctx,
		runtime:   r,
		modules:   make(map[string]wazeroapi.Module),
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

	// Try to compile/instantiate the module so errors surface early
	if len(wasmCode) > 0 {
		compiled, err := ce.runtime.CompileModule(ce.ctx, wasmCode)
		if err != nil {
			return fmt.Errorf("failed to compile wasm module: %w", err)
		}
		mod, err := ce.runtime.InstantiateModule(ce.ctx, compiled, wazero.NewModuleConfig())
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
		compiled, err := ce.runtime.CompileModule(ce.ctx, contract.Code)
		if err != nil {
			return nil, fmt.Errorf("failed to compile wasm module: %w", err)
		}
		m, err := ce.runtime.InstantiateModule(ce.ctx, compiled, wazero.NewModuleConfig())
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
	results, err := fn.Call(ce.ctx)
	if err != nil {
		return nil, fmt.Errorf("wasm contract handle failed: %w", err)
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

	// TODO: Full WASM validation using wasmer
	return nil
}
