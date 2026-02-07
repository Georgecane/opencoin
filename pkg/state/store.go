package state

import (
	"encoding/binary"
	"fmt"
	"path/filepath"

	"github.com/cockroachdb/pebble"

	"github.com/georgecane/opencoin/pkg/encoding"
	"github.com/georgecane/opencoin/pkg/types"
)

const (
	accountPrefix              = "acct/"
	contractPrefix             = "contract/"
	blockPrefix                = "block/"
	blockHeightPrefix          = "block_height/"
	metaPrefix                 = "meta/"
	metaLastTimestamps         = "meta/last_timestamps"
	metaConsensusHeight        = "meta/consensus_height"
	metaConsensusRound         = "meta/consensus_round"
	metaConsensusLastFinalized = "meta/consensus_last_finalized"
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

// NewIndexedBatch creates a new indexed batch for previewing state changes.
func (s *Store) NewIndexedBatch() *pebble.Batch {
	return s.db.NewIndexedBatch()
}

// GetAccount returns the account state for an address.
func (s *Store) GetAccount(addr types.Address) (*types.Account, error) {
	return getAccountFromReader(s.db, addr)
}

// SetAccount persists an account state.
func (s *Store) SetAccount(acct *types.Account) error {
	return setAccountWithWriter(s.db, acct, pebble.Sync)
}

// IterateAccounts iterates over all account entries.
func (s *Store) IterateAccounts(fn func(key []byte, acct *types.Account) error) error {
	return iterateAccounts(s.db, fn)
}

func getAccountFromReader(reader pebble.Reader, addr types.Address) (*types.Account, error) {
	key := []byte(accountPrefix + string(addr))
	val, closer, err := reader.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get account: %w", err)
	}
	defer closer.Close()
	return unmarshalAccount(val)
}

func setAccountWithWriter(writer pebble.Writer, acct *types.Account, opts *pebble.WriteOptions) error {
	if acct == nil {
		return fmt.Errorf("account is nil")
	}
	key := []byte(accountPrefix + string(acct.Address))
	val, err := marshalAccount(acct)
	if err != nil {
		return err
	}
	return writer.Set(key, val, opts)
}

func iterateAccounts(reader pebble.Reader, fn func(key []byte, acct *types.Account) error) error {
	iter, err := reader.NewIter(&pebble.IterOptions{
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

func setLastTimestampsWithWriter(writer pebble.Writer, ts []int64) error {
	val := encodeTimestamps(ts)
	return writer.Set([]byte(metaLastTimestamps), val, nil)
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

// SetBlock persists a block by hash and height.
func (s *Store) SetBlock(block *types.Block) (types.Hash, error) {
	if block == nil {
		return types.Hash{}, fmt.Errorf("block is nil")
	}
	hash, err := encoding.HashBlock(block)
	if err != nil {
		return types.Hash{}, err
	}
	if err := setBlockWithWriter(s.db, block, hash); err != nil {
		return types.Hash{}, err
	}
	return hash, nil
}

// GetBlockByHeight retrieves a block by height.
func (s *Store) GetBlockByHeight(height uint64) (*types.Block, error) {
	key := append([]byte(blockHeightPrefix), encoding.MarshalUint64(height)...)
	val, closer, err := s.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get block height: %w", err)
	}
	defer closer.Close()
	if len(val) != len(types.Hash{}) {
		return nil, fmt.Errorf("invalid block hash length")
	}
	var hash types.Hash
	copy(hash[:], val)
	return s.GetBlockByHash(hash)
}

// GetBlockByHash retrieves a block by header hash.
func (s *Store) GetBlockByHash(hash types.Hash) (*types.Block, error) {
	key := append([]byte(blockPrefix), hash[:]...)
	val, closer, err := s.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get block: %w", err)
	}
	defer closer.Close()
	return encoding.UnmarshalBlock(val)
}

func setBlockWithWriter(writer pebble.Writer, block *types.Block, hash types.Hash) error {
	blockBytes, err := encoding.MarshalBlock(block)
	if err != nil {
		return err
	}
	blockKey := append([]byte(blockPrefix), hash[:]...)
	heightKey := append([]byte(blockHeightPrefix), encoding.MarshalUint64(block.Height)...)
	if err := writer.Set(blockKey, blockBytes, nil); err != nil {
		return err
	}
	return writer.Set(heightKey, hash[:], nil)
}

// SetConsensusState persists consensus metadata.
func (s *Store) SetConsensusState(height, round uint64, lastFinalized types.Hash) error {
	batch := s.db.NewBatch()
	defer batch.Close()
	if err := batch.Set([]byte(metaConsensusHeight), encoding.MarshalUint64(height), nil); err != nil {
		return err
	}
	if err := batch.Set([]byte(metaConsensusRound), encoding.MarshalUint64(round), nil); err != nil {
		return err
	}
	if err := batch.Set([]byte(metaConsensusLastFinalized), lastFinalized[:], nil); err != nil {
		return err
	}
	return batch.Commit(pebble.Sync)
}

// GetConsensusState loads consensus metadata; returns zero values if not found.
func (s *Store) GetConsensusState() (uint64, uint64, types.Hash, error) {
	var height uint64
	var round uint64
	var lastFinalized types.Hash
	if val, closer, err := s.db.Get([]byte(metaConsensusHeight)); err == nil {
		if len(val) == 8 {
			height = binary.BigEndian.Uint64(val)
		}
		closer.Close()
	} else if err != pebble.ErrNotFound {
		return 0, 0, types.Hash{}, fmt.Errorf("get consensus height: %w", err)
	}
	if val, closer, err := s.db.Get([]byte(metaConsensusRound)); err == nil {
		if len(val) == 8 {
			round = binary.BigEndian.Uint64(val)
		}
		closer.Close()
	} else if err != pebble.ErrNotFound {
		return 0, 0, types.Hash{}, fmt.Errorf("get consensus round: %w", err)
	}
	if val, closer, err := s.db.Get([]byte(metaConsensusLastFinalized)); err == nil {
		if len(val) == len(lastFinalized) {
			copy(lastFinalized[:], val)
		}
		closer.Close()
	} else if err != pebble.ErrNotFound {
		return 0, 0, types.Hash{}, fmt.Errorf("get consensus last finalized: %w", err)
	}
	return height, round, lastFinalized, nil
}
