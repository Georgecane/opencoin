package governance

import (
	"fmt"
	"sync"

	"github.com/georgecane/opencoin/pkg/types"
)

// Params defines governance parameters.
type Params struct {
	VotingPeriodEpochs uint64
	QuorumPercent      uint64
	ThresholdPercent   uint64
	TimelockEpochs     uint64
}

// Manager handles governance proposals and votes.
type Manager struct {
	mu        sync.RWMutex
	params    Params
	proposals map[uint64]*types.GovernanceProposal
	votes     map[uint64]map[types.Address]types.VoteOption
	nextID    uint64
}

// New creates a new governance manager.
func New(params Params) *Manager {
	return &Manager{
		params:    params,
		proposals: make(map[uint64]*types.GovernanceProposal),
		votes:     make(map[uint64]map[types.Address]types.VoteOption),
		nextID:    1,
	}
}

// SubmitProposal registers a new governance proposal.
func (g *Manager) SubmitProposal(p types.GovernanceProposal) (uint64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	p.ID = g.nextID
	g.nextID++
	g.proposals[p.ID] = &p
	return p.ID, nil
}

// Vote records a vote.
func (g *Manager) Vote(v types.GovernanceVote) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.proposals[v.ProposalID]; !ok {
		return fmt.Errorf("proposal not found")
	}
	if g.votes[v.ProposalID] == nil {
		g.votes[v.ProposalID] = make(map[types.Address]types.VoteOption)
	}
	g.votes[v.ProposalID][v.Voter] = v.Option
	return nil
}

// GetProposal returns a proposal by id.
func (g *Manager) GetProposal(id uint64) (*types.GovernanceProposal, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	p, ok := g.proposals[id]
	if !ok {
		return nil, fmt.Errorf("proposal not found")
	}
	cp := *p
	return &cp, nil
}
