package consensus

import (
	"fmt"
	"sync"
	"time"

	"github.com/georgecane/opencoin/pkg/contracts"
	"github.com/georgecane/opencoin/pkg/crypto"
	"github.com/georgecane/opencoin/pkg/encoding"
	"github.com/georgecane/opencoin/pkg/mempool"
	"github.com/georgecane/opencoin/pkg/state"
	"github.com/georgecane/opencoin/pkg/types"
)

// Network abstracts consensus message dissemination.
type Network interface {
	BroadcastProposal(*types.Proposal) error
	BroadcastPrecommit(*types.PrecommitVote) error
	BroadcastQC(*types.QuorumCertificate) error
}

// Config defines consensus parameters.
type Config struct {
	EpochLength   uint64
	MaxValidators uint32
	BlockMaxTxs   int
	MinStake      uint64
	SlashDouble   uint64
	JailDouble    uint64
	SlashOffline  uint64
	JailOffline   uint64
}

// Engine implements a HotStuff-like linear consensus with DPoS validator sets.
type Engine struct {
	mu            sync.Mutex
	cfg           Config
	state         *state.State
	dpos          *DPoS
	mempool       *mempool.Mempool
	verifier      crypto.Verifier
	signer        crypto.Signer
	network       Network
	validatorSet  *types.ValidatorSet
	height        uint64
	round         uint64
	lastFinalized types.Hash
	votes         map[types.Hash]map[types.Address]*types.PrecommitVote
	validatorAddr types.Address
}

// NewEngine creates a new consensus engine.
func NewEngine(cfg Config, st *state.State, dpos *DPoS, mp *mempool.Mempool, signer crypto.Signer, verifier crypto.Verifier, net Network) (*Engine, error) {
	addr, err := crypto.AddressFromPubKey(signer.PublicKey())
	if err != nil {
		return nil, err
	}
	return &Engine{
		cfg:          cfg,
		state:        st,
		dpos:         dpos,
		mempool:      mp,
		signer:       signer,
		verifier:     verifier,
		network:      net,
		validatorSet: dpos.ValidatorSet(),
		votes:        make(map[types.Hash]map[types.Address]*types.PrecommitVote),
		validatorAddr: types.Address(addr),
	}, nil
}

// ProposeBlock builds and broadcasts a new proposal if this node is proposer.
func (e *Engine) ProposeBlock() (*types.Proposal, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isProposerLocked() {
		return nil, fmt.Errorf("not proposer")
	}
	txs, err := e.mempool.SelectForBlock(e.cfg.BlockMaxTxs)
	if err != nil {
		return nil, err
	}
	block := &types.Block{
		Height:       e.height + 1,
		PrevHash:     e.lastFinalized,
		StateRoot:    types.Hash{},
		Timestamp:    time.Now().Unix(),
		Proposer:     e.validatorAddress(),
		Transactions: txs,
		ValidatorSigs: make([][]byte, len(e.validatorSet.Validators)),
	}
	blockHash, err := encoding.HashBlock(block)
	if err != nil {
		return nil, err
	}
	block.StateRoot = blockHash // placeholder until state applied
	prop := &types.Proposal{
		Block: block,
		Round: e.round,
	}
	propBytes, err := ProposalSignBytes(prop)
	if err != nil {
		return nil, err
	}
	sig, err := e.signer.Sign(propBytes)
	if err != nil {
		return nil, err
	}
	prop.ProposerSig = sig
	if e.network != nil {
		if err := e.network.BroadcastProposal(prop); err != nil {
			return nil, err
		}
	}
	return prop, nil
}

// HandleProposal validates and votes for a proposal.
func (e *Engine) HandleProposal(prop *types.Proposal) (*types.PrecommitVote, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if prop == nil || prop.Block == nil {
		return nil, fmt.Errorf("invalid proposal")
	}
	if prop.Block.Height != e.height+1 {
		return nil, fmt.Errorf("unexpected height")
	}
	if !e.isExpectedProposerLocked(prop.Block.Proposer) {
		return nil, fmt.Errorf("unexpected proposer")
	}
	propBytes, err := ProposalSignBytes(prop)
	if err != nil {
		return nil, err
	}
	pk, ok := e.validatorPubKey(prop.Block.Proposer)
	if !ok {
		return nil, fmt.Errorf("unknown proposer")
	}
	if !e.verifier.Verify(propBytes, prop.ProposerSig, pk) {
		return nil, fmt.Errorf("invalid proposer signature")
	}
	vote := &types.PrecommitVote{
		BlockHash: mustHashBlock(prop.Block),
		Height:    prop.Block.Height,
		Round:     prop.Round,
		Validator: e.validatorAddress(),
	}
	voteBytes, err := PrecommitSignBytes(vote)
	if err != nil {
		return nil, err
	}
	sig, err := e.signer.Sign(voteBytes)
	if err != nil {
		return nil, err
	}
	vote.Signature = sig
	if e.network != nil {
		if err := e.network.BroadcastPrecommit(vote); err != nil {
			return nil, err
		}
	}
	return vote, nil
}

// HandlePrecommitVote records a vote and finalizes if quorum reached.
func (e *Engine) HandlePrecommitVote(vote *types.PrecommitVote) (*types.QuorumCertificate, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if vote == nil {
		return nil, fmt.Errorf("nil vote")
	}
	if vote.Height != e.height+1 {
		return nil, fmt.Errorf("unexpected vote height")
	}
	voteBytes, err := PrecommitSignBytes(vote)
	if err != nil {
		return nil, err
	}
	pk, ok := e.validatorPubKey(vote.Validator)
	if !ok {
		return nil, fmt.Errorf("unknown validator")
	}
	if !e.verifier.Verify(voteBytes, vote.Signature, pk) {
		return nil, fmt.Errorf("invalid vote signature")
	}

	vmap := e.votes[vote.BlockHash]
	if vmap == nil {
		vmap = make(map[types.Address]*types.PrecommitVote)
		e.votes[vote.BlockHash] = vmap
	}
	vmap[vote.Validator] = vote

	qc, ok := e.tryBuildQC(vote.BlockHash, vmap)
	if ok {
		if e.network != nil {
			_ = e.network.BroadcastQC(qc)
		}
		return qc, nil
	}
	return nil, nil
}

// HandleViewChange processes a view change message.
func (e *Engine) HandleViewChange(vc *types.ViewChange) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if vc == nil {
		return fmt.Errorf("nil view change")
	}
	if vc.Height != e.height+1 {
		return fmt.Errorf("unexpected view change height")
	}
	pk, ok := e.validatorPubKey(vc.Validator)
	if !ok {
		return fmt.Errorf("unknown validator")
	}
	msg, err := ViewChangeSignBytes(vc)
	if err != nil {
		return err
	}
	if !e.verifier.Verify(msg, vc.Signature, pk) {
		return fmt.Errorf("invalid view change signature")
	}
	if vc.Round > e.round {
		e.round = vc.Round
	}
	return nil
}

// OnTimeout advances the round for view change.
func (e *Engine) OnTimeout() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.round++
}

// FinalizeBlock finalizes the block once QC achieved.
func (e *Engine) FinalizeBlock(block *types.Block, qc *types.QuorumCertificate, contracts *contracts.ContractEngine) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if block == nil || qc == nil {
		return fmt.Errorf("invalid finalize arguments")
	}
	if len(qc.Signatures) == len(e.validatorSet.Validators) {
		block.ValidatorSigs = qc.Signatures
	}
	root, err := e.state.ApplyBlock(block, contracts)
	if err != nil {
		return err
	}
	block.StateRoot = root
	e.height = block.Height
	e.lastFinalized = mustHashBlock(block)
	e.validatorSet = e.dpos.ValidatorSet()
	e.round = 0
	delete(e.votes, qc.BlockHash)
	return nil
}

func (e *Engine) tryBuildQC(blockHash types.Hash, votes map[types.Address]*types.PrecommitVote) (*types.QuorumCertificate, bool) {
	totalPower := e.validatorSet.TotalPower
	if totalPower == 0 {
		return nil, false
	}
	var signedPower uint64
	signatures := make([][]byte, len(e.validatorSet.Validators))
	bitmap := make([]byte, (len(e.validatorSet.Validators)+7)/8)
	for addr, vote := range votes {
		idx, ok := e.validatorSet.IndexByAddr[addr]
		if !ok {
			continue
		}
		signatures[idx] = vote.Signature
		bitmap[idx/8] |= 1 << (idx % 8)
		signedPower += e.validatorSet.Validators[idx].Power
	}
	if signedPower*3 <= totalPower*2 {
		return nil, false
	}
	qc := &types.QuorumCertificate{
		BlockHash:  blockHash,
		Height:     e.height + 1,
		Round:      e.round,
		SigBitmap:  bitmap,
		Signatures: signatures,
	}
	return qc, true
}

func (e *Engine) isProposerLocked() bool {
	return e.isExpectedProposerLocked(e.validatorAddr)
}

func (e *Engine) isExpectedProposerLocked(addr types.Address) bool {
	if e.validatorSet == nil || len(e.validatorSet.Validators) == 0 {
		return false
	}
	totalPower := e.validatorSet.TotalPower
	if totalPower == 0 {
		return false
	}
	seed := (e.height + e.round) % totalPower
	var acc uint64
	for _, v := range e.validatorSet.Validators {
		acc += v.Power
		if seed < acc {
			return v.Address == addr
		}
	}
	return false
}

func mustHashBlock(block *types.Block) types.Hash {
	h, _ := encoding.HashBlock(block)
	return h
}

func (e *Engine) validatorAddress() types.Address {
	return e.validatorAddr
}

func (e *Engine) validatorPubKey(addr types.Address) ([]byte, bool) {
	if e.validatorSet == nil {
		return nil, false
	}
	idx, ok := e.validatorSet.IndexByAddr[addr]
	if !ok {
		return nil, false
	}
	if int(idx) >= len(e.validatorSet.Validators) {
		return nil, false
	}
	return e.validatorSet.Validators[idx].PublicKey, true
}
