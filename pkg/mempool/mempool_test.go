package mempool

import (
	"bytes"
	"crypto/ed25519"
	"testing"

	"github.com/georgecane/opencoin/pkg/crypto"
	"github.com/georgecane/opencoin/pkg/tx"
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

type mockCoster struct {
	priority types.Address
}

func (c *mockCoster) Cost(tx *types.Transaction) (uint64, error) {
	if tx.From == c.priority && tx.Nonce == 0 {
		return 10, nil
	}
	return 5, nil
}

func TestMempoolOrdering(t *testing.T) {
	kpA := keyFromSeed(0x01)
	kpB := keyFromSeed(0x02)
	addrA, _ := crypto.AddressFromPubKey(kpA.PublicKey)
	addrB, _ := crypto.AddressFromPubKey(kpB.PublicKey)
	state := &mockState{
		accounts: map[types.Address]*types.Account{
			types.Address(addrA): {Address: types.Address(addrA), Nonce: 0, RC: 100},
			types.Address(addrB): {Address: types.Address(addrB), Nonce: 0, RC: 100},
		},
	}
	coster := &mockCoster{priority: types.Address(addrB)}
	mp := New(state, coster)

	txA0 := mustSignedTransfer(t, kpA, types.Address(addrA), "x", 0)
	txA1 := mustSignedTransfer(t, kpA, types.Address(addrA), "x", 1)
	txB0 := mustSignedTransfer(t, kpB, types.Address(addrB), "y", 0)

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
	if selected[0].From != types.Address(addrB) || selected[0].Nonce != 0 {
		t.Fatalf("unexpected first tx")
	}
	if selected[1].From != types.Address(addrA) || selected[1].Nonce != 0 {
		t.Fatalf("unexpected second tx")
	}
	if selected[2].From != types.Address(addrA) || selected[2].Nonce != 1 {
		t.Fatalf("unexpected third tx")
	}
}

func keyFromSeed(b byte) *crypto.Ed25519KeyPair {
	seed := bytes.Repeat([]byte{b}, ed25519.SeedSize)
	priv := ed25519.NewKeyFromSeed(seed)
	return &crypto.Ed25519KeyPair{
		PublicKey:  priv.Public().(ed25519.PublicKey),
		PrivateKey: priv,
	}
}

func mustSignedTransfer(t *testing.T, kp *crypto.Ed25519KeyPair, from types.Address, to string, nonce uint64) *types.Transaction {
	t.Helper()
	payload, err := tx.EncodePayload(tx.Transfer{To: types.Address(to), Amount: 1}, kp.PublicKey)
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}
	txn := &types.Transaction{
		From:    from,
		To:      types.Address(to),
		Nonce:   nonce,
		Payload: payload,
	}
	signBytes, err := tx.SigningBytes(txn)
	if err != nil {
		t.Fatalf("sign bytes: %v", err)
	}
	sig, err := crypto.SignEd25519(kp.PrivateKey, signBytes)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	txn.Signature = sig
	return txn
}
