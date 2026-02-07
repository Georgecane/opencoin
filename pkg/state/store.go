package state

import (
	"encoding/binary"
	"fmt"
	"path/filepath"

	"github.com/cockroachdb/pebble"

	"github.com/georgecane/opencoin/pkg/types"
)

const (
	accountPrefix = "acct/"
	contractPrefix = "contract/"
	metaPrefix    = "meta/"
	metaLastTimestamps = "meta/last_timestamps"
)

// Store is the persistent state store backed by Pebble.
type Store struct {
	db *pebble.DB
}

// OpenStore opens or creates a Pebble store at the given path.
func OpenStore(home string) (*Store, error) {
	path := filepath.Join(home, "state")
	db, err := pebble.Open(path, &pebble.Options{})
	if err != nil {
		return nil, fmt.Errorf("open pebble: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the store.
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// GetAccount returns the account state for an address.
func (s *Store) GetAccount(addr types.Address) (*types.Account, error) {
	key := []byte(accountPrefix + string(addr))
	val, closer, err := s.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get account: %w", err)
	}
	defer closer.Close()
	return unmarshalAccount(val)
}

// SetAccount persists an account state.
func (s *Store) SetAccount(acct *types.Account) error {
	if acct == nil {
		return fmt.Errorf("account is nil")
	}
	key := []byte(accountPrefix + string(acct.Address))
	val, err := marshalAccount(acct)
	if err != nil {
		return err
	}
	return s.db.Set(key, val, pebble.Sync)
}

// IterateAccounts iterates over all account entries.
func (s *Store) IterateAccounts(fn func(key []byte, acct *types.Account) error) error {
	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(accountPrefix),
		UpperBound: []byte(accountPrefix + string([]byte{0xFF})),
	})
	if err != nil {
		return err
	}
	defer iter.Close()
	for iter.First(); iter.Valid(); iter.Next() {
		key := append([]byte(nil), iter.Key()...)
		acct, err := unmarshalAccount(iter.Value())
		if err != nil {
			return err
		}
		if err := fn(key, acct); err != nil {
			return err
		}
	}
	return iter.Error()
}

// GetLastTimestamps returns the last N block timestamps stored.
func (s *Store) GetLastTimestamps() ([]int64, error) {
	val, closer, err := s.db.Get([]byte(metaLastTimestamps))
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get last timestamps: %w", err)
	}
	defer closer.Close()
	return decodeTimestamps(val)
}

// SetLastTimestamps stores the last N block timestamps.
func (s *Store) SetLastTimestamps(ts []int64) error {
	val := encodeTimestamps(ts)
	return s.db.Set([]byte(metaLastTimestamps), val, pebble.Sync)
}

func encodeTimestamps(ts []int64) []byte {
	buf := make([]byte, 0, 4+len(ts)*8)
	tmp := make([]byte, 8)
	binary.BigEndian.PutUint32(tmp[:4], uint32(len(ts)))
	buf = append(buf, tmp[:4]...)
	for _, t := range ts {
		binary.BigEndian.PutUint64(tmp, uint64(t))
		buf = append(buf, tmp...)
	}
	return buf
}

func decodeTimestamps(b []byte) ([]int64, error) {
	if len(b) < 4 {
		return nil, fmt.Errorf("invalid timestamp encoding")
	}
	n := binary.BigEndian.Uint32(b[:4])
	b = b[4:]
	if len(b) < int(n)*8 {
		return nil, fmt.Errorf("invalid timestamp length")
	}
	out := make([]int64, 0, n)
	for i := 0; i < int(n); i++ {
		v := binary.BigEndian.Uint64(b[i*8 : (i+1)*8])
		out = append(out, int64(v))
	}
	return out, nil
}
