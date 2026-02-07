package state

import (
	"crypto/sha256"
	"sort"

	"github.com/georgecane/opencoin/pkg/types"
)

// ComputeStateRoot computes a deterministic Merkle root over account state.
func ComputeStateRoot(store *Store) (types.Hash, error) {
	type kv struct {
		key []byte
		val []byte
	}
	var items []kv
	err := store.IterateAccounts(func(key []byte, acct *types.Account) error {
		val, err := marshalAccount(acct)
		if err != nil {
			return err
		}
		items = append(items, kv{key: append([]byte(nil), key...), val: val})
		return nil
	})
	if err != nil {
		return types.Hash{}, err
	}
	sort.Slice(items, func(i, j int) bool {
		return string(items[i].key) < string(items[j].key)
	})
	leaves := make([][]byte, 0, len(items))
	for _, it := range items {
		h := sha256.New()
		h.Write([]byte{0x00})
		h.Write(it.key)
		h.Write(it.val)
		leaves = append(leaves, h.Sum(nil))
	}
	if len(leaves) == 0 {
		return types.Hash{}, nil
	}
	root := merkleRoot(leaves)
	var out types.Hash
	copy(out[:], root)
	return out, nil
}

func merkleRoot(leaves [][]byte) []byte {
	if len(leaves) == 0 {
		return nil
	}
	nodes := leaves
	for len(nodes) > 1 {
		var next [][]byte
		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			right := left
			if i+1 < len(nodes) {
				right = nodes[i+1]
			}
			h := sha256.New()
			h.Write([]byte{0x01})
			h.Write(left)
			h.Write(right)
			next = append(next, h.Sum(nil))
		}
		nodes = next
	}
	return nodes[0]
}
