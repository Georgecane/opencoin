package mempool

import (
	"container/heap"
	"fmt"
	"sync"

	"github.com/georgecane/opencoin/pkg/encoding"
	"github.com/georgecane/opencoin/pkg/types"
)

// StateView provides account state to the mempool.
type StateView interface {
	GetAccount(addr types.Address) (*types.Account, error)
}

// Coster computes RC cost for a transaction.
type Coster interface {
	Cost(tx *types.Transaction) (uint64, error)
}

// Mempool stores pending transactions with deterministic ordering rules.
type Mempool struct {
	mu     sync.RWMutex
	state  StateView
	coster Coster
	pool   map[types.Address][]*types.Transaction
}

// New creates a new mempool.
func New(state StateView, coster Coster) *Mempool {
	return &Mempool{
		state:  state,
		coster: coster,
		pool:   make(map[types.Address][]*types.Transaction),
	}
}

// AddTx adds a transaction to the mempool with RC/nonce checks.
func (m *Mempool) AddTx(tx *types.Transaction) error {
	if tx == nil {
		return fmt.Errorf("tx is nil")
	}
	if tx.From == "" || tx.To == "" {
		return fmt.Errorf("invalid sender or recipient")
	}
	acct, err := m.state.GetAccount(tx.From)
	if err != nil {
		return err
	}
	if acct == nil {
		return fmt.Errorf("account not found")
	}
	if tx.Nonce < acct.Nonce {
		return fmt.Errorf("stale nonce")
	}
	cost, err := m.coster.Cost(tx)
	if err != nil {
		return err
	}
	if acct.RC < cost {
		return fmt.Errorf("insufficient rc")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	queue := m.pool[tx.From]
	// Insert by nonce order.
	inserted := false
	for i, existing := range queue {
		if tx.Nonce == existing.Nonce {
			return fmt.Errorf("duplicate nonce")
		}
		if tx.Nonce < existing.Nonce {
			queue = append(queue[:i], append([]*types.Transaction{tx}, queue[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		queue = append(queue, tx)
	}
	m.pool[tx.From] = queue
	return nil
}

// SelectForBlock returns up to max transactions in deterministic order:
// nonce ascending per sender, then RC_cost descending, tie-break by tx hash.
func (m *Mempool) SelectForBlock(max int) ([]*types.Transaction, error) {
	if max <= 0 {
		return nil, nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	type senderState struct {
		acct   *types.Account
		cursor int
		queue  []*types.Transaction
	}
	senders := make(map[types.Address]*senderState)
	for addr, queue := range m.pool {
		acct, err := m.state.GetAccount(addr)
		if err != nil {
			return nil, err
		}
		if acct == nil || len(queue) == 0 {
			continue
		}
		senders[addr] = &senderState{acct: acct, queue: queue}
	}

	h := &txHeap{}
	heap.Init(h)

	// Seed heap with first valid tx per sender.
	for addr, st := range senders {
		tx := st.queue[0]
		if tx.Nonce != st.acct.Nonce {
			continue
		}
		cost, err := m.coster.Cost(tx)
		if err != nil {
			return nil, err
		}
		if st.acct.RC < cost {
			continue
		}
		txHash, err := encoding.HashTransaction(tx)
		if err != nil {
			return nil, err
		}
		heap.Push(h, &txItem{tx: tx, cost: cost, hash: txHash, sender: addr})
	}

	var out []*types.Transaction
	for h.Len() > 0 && len(out) < max {
		item := heap.Pop(h).(*txItem)
		out = append(out, item.tx)

		st := senders[item.sender]
		// Apply tx locally to update sender state for selection.
		st.acct.Nonce++
		if st.acct.RC >= item.cost {
			st.acct.RC -= item.cost
		} else {
			st.acct.RC = 0
		}
		st.cursor++
		if st.cursor >= len(st.queue) {
			continue
		}
		next := st.queue[st.cursor]
		if next.Nonce != st.acct.Nonce {
			continue
		}
		cost, err := m.coster.Cost(next)
		if err != nil {
			return nil, err
		}
		if st.acct.RC < cost {
			continue
		}
		hash, err := encoding.HashTransaction(next)
		if err != nil {
			return nil, err
		}
		heap.Push(h, &txItem{tx: next, cost: cost, hash: hash, sender: item.sender})
	}
	return out, nil
}

type txItem struct {
	tx     *types.Transaction
	cost   uint64
	hash   types.Hash
	sender types.Address
}

type txHeap []*txItem

func (h txHeap) Len() int { return len(h) }
func (h txHeap) Less(i, j int) bool {
	if h[i].cost != h[j].cost {
		return h[i].cost > h[j].cost
	}
	return h[i].hash.String() < h[j].hash.String()
}
func (h txHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *txHeap) Push(x interface{}) {
	*h = append(*h, x.(*txItem))
}
func (h *txHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}
