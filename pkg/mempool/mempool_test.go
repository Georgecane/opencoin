package mempool

import (
	"testing"

	"github.com/georgecane/opencoin/pkg/types"
)

type mockState struct {
	accounts map[types.Address]*types.Account
}

func (m *mockState) GetAccount(addr types.Address) (*types.Account, error) {
	if acct, ok := m.accounts[addr]; ok {
		cp := *acct
		return &cp, nil
	}
	return nil, nil
}

type mockCoster struct{}

func (c *mockCoster) Cost(tx *types.Transaction) (uint64, error) {
	if tx.From == "b" && tx.Nonce == 0 {
		return 10, nil
	}
	return 5, nil
}

func TestMempoolOrdering(t *testing.T) {
	state := &mockState{
		accounts: map[types.Address]*types.Account{
			"a": {Address: "a", Nonce: 0, RC: 100},
			"b": {Address: "b", Nonce: 0, RC: 100},
		},
	}
	coster := &mockCoster{}
	mp := New(state, coster)

	txA0 := &types.Transaction{From: "a", To: "x", Nonce: 0}
	txA1 := &types.Transaction{From: "a", To: "x", Nonce: 1}
	txB0 := &types.Transaction{From: "b", To: "y", Nonce: 0}

	if err := mp.AddTx(txA0); err != nil {
		t.Fatalf("add txA0: %v", err)
	}
	if err := mp.AddTx(txA1); err != nil {
		t.Fatalf("add txA1: %v", err)
	}
	if err := mp.AddTx(txB0); err != nil {
		t.Fatalf("add txB0: %v", err)
	}

	selected, err := mp.SelectForBlock(3)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if len(selected) != 3 {
		t.Fatalf("expected 3 selected got %d", len(selected))
	}
	// B0 has highest cost, then A0 (nonce order), then A1.
	if selected[0].From != "b" || selected[0].Nonce != 0 {
		t.Fatalf("unexpected first tx")
	}
	if selected[1].From != "a" || selected[1].Nonce != 0 {
		t.Fatalf("unexpected second tx")
	}
	if selected[2].From != "a" || selected[2].Nonce != 1 {
		t.Fatalf("unexpected third tx")
	}
}
