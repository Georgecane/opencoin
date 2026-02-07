package state

import (
	"fmt"
	"sync"

	"github.com/georgecane/opencoin/pkg/types"
)

// DAG manages state versioning nodes.
type DAG struct {
	mu       sync.RWMutex
	nodes    map[types.Hash]*types.StateNode
	children map[types.Hash][]types.Hash
	tips     []types.Hash
}

// NewDAG creates a new DAG.
func NewDAG() *DAG {
	return &DAG{
		nodes:    make(map[types.Hash]*types.StateNode),
		children: make(map[types.Hash][]types.Hash),
		tips:     make([]types.Hash, 0),
	}
}

// AddNode inserts a new StateNode.
func (d *DAG) AddNode(node *types.StateNode) error {
	if node == nil {
		return fmt.Errorf("state node is nil")
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.nodes[node.RootHash]; exists {
		return fmt.Errorf("state node already exists: %s", node.RootHash.String())
	}

	d.nodes[node.RootHash] = node
	for _, parent := range node.Parents {
		d.children[parent] = append(d.children[parent], node.RootHash)
	}

	d.updateTips(node.RootHash, node.Parents)
	return nil
}

// updateTips updates the DAG frontier tips.
func (d *DAG) updateTips(newNode types.Hash, parents []types.Hash) {
	parentSet := make(map[types.Hash]struct{}, len(parents))
	for _, p := range parents {
		parentSet[p] = struct{}{}
	}
	next := make([]types.Hash, 0, len(d.tips)+1)
	for _, tip := range d.tips {
		if _, ok := parentSet[tip]; !ok {
			next = append(next, tip)
		}
	}
	next = append(next, newNode)
	d.tips = next
}

// GetNode returns a state node by root hash.
func (d *DAG) GetNode(hash types.Hash) *types.StateNode {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nodes[hash]
}

// Tips returns the current DAG frontier.
func (d *DAG) Tips() []types.Hash {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]types.Hash, len(d.tips))
	copy(out, d.tips)
	return out
}

// PruneNonFinal removes nodes not reachable from the finalized root.
func (d *DAG) PruneNonFinal(finalRoot types.Hash) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// Mark all reachable nodes from finalRoot.
	seen := map[types.Hash]struct{}{finalRoot: {}}
	var walk func(h types.Hash)
	walk = func(h types.Hash) {
		for _, child := range d.children[h] {
			if _, ok := seen[child]; ok {
				continue
			}
			seen[child] = struct{}{}
			walk(child)
		}
	}
	walk(finalRoot)
	for h := range d.nodes {
		if _, ok := seen[h]; !ok {
			delete(d.nodes, h)
			delete(d.children, h)
		}
	}
	// Recompute tips.
	d.tips = d.tips[:0]
	for h, node := range d.nodes {
		if len(node.Parents) == 0 {
			continue
		}
		// Tip if no children remain.
		if len(d.children[h]) == 0 {
			d.tips = append(d.tips, h)
		}
	}
}
