package state

import (
	"fmt"
	"sort"
	"sync"

	"github.com/georgecane/opencoin/pkg/types"
)

// DAGState manages the DAG-based state of the blockchain
type DAGState struct {
	mu       sync.RWMutex
	vertices map[string]*types.DAGVertex
	tips     []*types.DAGVertex // frontier (recent blocks)
	accounts map[string]*types.AccountState
	order    uint64 // counter for topological order
}

// NewDAGState creates a new DAG state manager
func NewDAGState() *DAGState {
	return &DAGState{
		vertices: make(map[string]*types.DAGVertex),
		tips:     make([]*types.DAGVertex, 0),
		accounts: make(map[string]*types.AccountState),
		order:    0,
	}
}

// AddBlock adds a new block to the DAG
func (ds *DAGState) AddBlock(block *types.Block) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if _, exists := ds.vertices[block.Hash]; exists {
		return fmt.Errorf("block %s already exists", block.Hash)
	}

	vertex := &types.DAGVertex{
		BlockHash: block.Hash,
		Block:     block,
		Parents:   make([]*types.DAGVertex, 0),
		Children:  make([]*types.DAGVertex, 0),
	}

	// Link to parents
	for _, parentHash := range block.Parents {
		if parentVertex, ok := ds.vertices[parentHash]; ok {
			vertex.Parents = append(vertex.Parents, parentVertex)
			parentVertex.Children = append(parentVertex.Children, vertex)
		}
	}

	ds.vertices[block.Hash] = vertex

	// Update tips (remove parents from tips, add this vertex)
	ds.updateTips(vertex, block.Parents)

	// Compute topological order
	ds.order++
	vertex.Order = ds.order

	return nil
}

// updateTips updates the frontier of the DAG
func (ds *DAGState) updateTips(newVertex *types.DAGVertex, parentHashes []string) {
	// Remove parents from tips
	newTips := make([]*types.DAGVertex, 0)
	parentSet := make(map[string]bool)
	for _, h := range parentHashes {
		parentSet[h] = true
	}

	for _, tip := range ds.tips {
		if !parentSet[tip.BlockHash] {
			newTips = append(newTips, tip)
		}
	}

	newTips = append(newTips, newVertex)
	ds.tips = newTips
}

// GetTips returns the current frontier of the DAG
func (ds *DAGState) GetTips() []*types.DAGVertex {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	tips := make([]*types.DAGVertex, len(ds.tips))
	copy(tips, ds.tips)
	return tips
}

// ApplyBlock applies a block's transactions to the state
func (ds *DAGState) ApplyBlock(block *types.Block) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for _, tx := range block.Txs {
		if err := ds.applyTransaction(tx); err != nil {
			return fmt.Errorf("failed to apply transaction %s: %w", tx.ID, err)
		}
	}

	return nil
}

// applyTransaction applies a single transaction to account state
func (ds *DAGState) applyTransaction(tx *types.Transaction) error {
	// Get or create sender account
	sender, ok := ds.accounts[tx.From]
	if !ok {
		sender = &types.AccountState{
			Address: tx.From,
			Balance: 0,
			Nonce:   0,
		}
		ds.accounts[tx.From] = sender
	}

	// Check nonce
	if tx.Nonce != sender.Nonce {
		return fmt.Errorf("invalid nonce: expected %d, got %d", sender.Nonce, tx.Nonce)
	}

	// Update sender balance
	if sender.Balance < tx.Amount {
		return fmt.Errorf("insufficient balance")
	}
	sender.Balance -= tx.Amount
	sender.Nonce++

	// Handle contract call or simple transfer
	if tx.Contract != nil {
		// Contract execution happens in contract engine
		return nil
	}

	// Simple transfer to recipient
	recipient, ok := ds.accounts[tx.To]
	if !ok {
		recipient = &types.AccountState{
			Address: tx.To,
			Balance: 0,
			Nonce:   0,
		}
		ds.accounts[tx.To] = recipient
	}

	recipient.Balance += tx.Amount

	return nil
}

// GetAccount returns the current state of an account
func (ds *DAGState) GetAccount(address string) *types.AccountState {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	account, ok := ds.accounts[address]
	if !ok {
		return nil
	}

	// Return a copy to prevent external modification
	accountCopy := *account
	return &accountCopy
}

// GetTopologicalOrder returns blocks in topological order
func (ds *DAGState) GetTopologicalOrder() []*types.DAGVertex {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	vertices := make([]*types.DAGVertex, 0)
	for _, v := range ds.vertices {
		vertices = append(vertices, v)
	}

	// Sort by topological order
	sort.Slice(vertices, func(i, j int) bool {
		return vertices[i].Order < vertices[j].Order
	})

	return vertices
}

// GetBlock returns a block by hash
func (ds *DAGState) GetBlock(hash string) *types.Block {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if vertex, ok := ds.vertices[hash]; ok {
		return vertex.Block
	}
	return nil
}

// GetVertex returns a DAG vertex by hash
func (ds *DAGState) GetVertex(hash string) *types.DAGVertex {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.vertices[hash]
}
