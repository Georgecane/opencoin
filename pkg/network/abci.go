//go:build cometbft

package network

import (
	"encoding/json"
	"fmt"

	abcitypes "github.com/cometbft/cometbft/abci/types"

	"github.com/georgecane/opencoin/pkg/consensus"
	"github.com/georgecane/opencoin/pkg/contracts"
	"github.com/georgecane/opencoin/pkg/crypto"
	"github.com/georgecane/opencoin/pkg/state"
	"github.com/georgecane/opencoin/pkg/types"
)

// ABCIApp implements the CometBFT ABCI interface
type ABCIApp struct {
	dagState       *state.DAGState
	dposEngine     *consensus.DPoSEngine
	contractEngine *contracts.ContractEngine
	verifier       crypto.Verifier
	mempool        []*types.Transaction
	blockHeight    uint64
}

// NewABCIApp creates a new ABCI application
func NewABCIApp(dagState *state.DAGState, dposEngine *consensus.DPoSEngine, contractEngine *contracts.ContractEngine) *ABCIApp {
	return &ABCIApp{
		dagState:       dagState,
		dposEngine:     dposEngine,
		contractEngine: contractEngine,
		verifier:       crypto.NewDilithiumVerifier(),
		mempool:        make([]*types.Transaction, 0),
		blockHeight:    0,
	}
}

// Info implements ABCI Info
func (app *ABCIApp) Info(req abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{
		Data:             "opencoin",
		Version:          "0.1.0",
		AppVersion:       1,
		LastBlockHeight:  int64(app.blockHeight),
		LastBlockAppHash: []byte("opencoin"),
	}
}

// InitChain implements ABCI InitChain
func (app *ABCIApp) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	// Initialize validator set from chain params
	return abcitypes.ResponseInitChain{}
}

// PrepareProposal implements ABCI PrepareProposal
func (app *ABCIApp) PrepareProposal(req abcitypes.RequestPrepareProposal) abcitypes.ResponsePrepareProposal {
	txs := make([][]byte, 0)

	// Include pending transactions from mempool
	for _, tx := range app.mempool {
		if len(txs) >= 1000 { // Max 1000 txs per block
			break
		}
		txBytes, _ := json.Marshal(tx)
		txs = append(txs, txBytes)
	}

	return abcitypes.ResponsePrepareProposal{
		Txs: txs,
	}
}

// ProcessProposal implements ABCI ProcessProposal
func (app *ABCIApp) ProcessProposal(req abcitypes.RequestProcessProposal) abcitypes.ResponseProcessProposal {
	// Validate all transactions in the proposal
	for _, txBytes := range req.Txs {
		var tx types.Transaction
		if err := json.Unmarshal(txBytes, &tx); err != nil {
			return abcitypes.ResponseProcessProposal{}
		}

		if err := app.validateTransaction(&tx); err != nil {
			return abcitypes.ResponseProcessProposal{}
		}
	}

	return abcitypes.ResponseProcessProposal{}
}

// CheckTx implements ABCI CheckTx
func (app *ABCIApp) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	var tx types.Transaction
	if err := json.Unmarshal(req.Tx, &tx); err != nil {
		return abcitypes.ResponseCheckTx{Code: 1, Log: "invalid transaction"}
	}

	if err := app.validateTransaction(&tx); err != nil {
		return abcitypes.ResponseCheckTx{Code: 1, Log: err.Error()}
	}

	app.mempool = append(app.mempool, &tx)
	return abcitypes.ResponseCheckTx{Code: 0}
}

// DeliverTx implements ABCI DeliverTx (execute transaction in block)
func (app *ABCIApp) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	var tx types.Transaction
	if err := json.Unmarshal(req.Tx, &tx); err != nil {
		return abcitypes.ResponseDeliverTx{Code: 1, Log: "invalid transaction"}
	}

	if err := app.dagState.ApplyBlock(&types.Block{Txs: []*types.Transaction{&tx}}); err != nil {
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error()}
	}

	// Execute contract if present
	if tx.Contract != nil {
		ctx := &contracts.ExecutionContext{
			Caller:       tx.From,
			ContractAddr: tx.Contract.Address,
			Value:        tx.Amount,
			Gas:          tx.Contract.Gas,
		}
		if _, err := app.contractEngine.ExecuteContract(ctx); err != nil {
			return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error()}
		}
	}

	return abcitypes.ResponseDeliverTx{Code: 0}
}

// EndBlock implements ABCI EndBlock
func (app *ABCIApp) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	// Distribute block rewards
	return abcitypes.ResponseEndBlock{}
}

// Commit implements ABCI Commit
func (app *ABCIApp) Commit() abcitypes.ResponseCommit {
	// Commit the state
	app.blockHeight++
	app.mempool = make([]*types.Transaction, 0) // Clear mempool
	return abcitypes.ResponseCommit{}
}

// Query implements ABCI Query
func (app *ABCIApp) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	// Handle queries for account balance, contract state, etc.
	if len(req.Path) == 0 {
		return abcitypes.ResponseQuery{Code: 1, Log: "empty path"}
	}

	switch req.Path {
	case "account":
		address := string(req.Data)
		account := app.dagState.GetAccount(address)
		if account == nil {
			return abcitypes.ResponseQuery{Code: 1, Log: "account not found"}
		}
		data, _ := json.Marshal(account)
		return abcitypes.ResponseQuery{Code: 0, Value: data}

	case "validator":
		address := string(req.Data)
		validator := app.dposEngine.GetValidator(address)
		if validator == nil {
			return abcitypes.ResponseQuery{Code: 1, Log: "validator not found"}
		}
		data, _ := json.Marshal(validator)
		return abcitypes.ResponseQuery{Code: 0, Value: data}

	case "contract":
		address := string(req.Data)
		contract := app.contractEngine.GetContract(address)
		if contract == nil {
			return abcitypes.ResponseQuery{Code: 1, Log: "contract not found"}
		}
		data, _ := json.Marshal(contract)
		return abcitypes.ResponseQuery{Code: 0, Value: data}

	default:
		return abcitypes.ResponseQuery{Code: 1, Log: "unknown path"}
	}
}

// ListSnapshots implements ABCI ListSnapshots (for state sync)
func (app *ABCIApp) ListSnapshots(req abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{Snapshots: []*abcitypes.Snapshot{}}
}

// OfferSnapshot implements ABCI OfferSnapshot (for state sync)
func (app *ABCIApp) OfferSnapshot(req abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

// LoadSnapshotChunk implements ABCI LoadSnapshotChunk (for state sync)
func (app *ABCIApp) LoadSnapshotChunk(req abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

// ApplySnapshotChunk implements ABCI ApplySnapshotChunk (for state sync)
func (app *ABCIApp) ApplySnapshotChunk(req abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}

// validateTransaction validates a transaction
func (app *ABCIApp) validateTransaction(tx *types.Transaction) error {
	if tx.From == "" || tx.To == "" {
		return fmt.Errorf("invalid sender or recipient")
	}

	account := app.dagState.GetAccount(tx.From)
	if account == nil {
		return fmt.Errorf("account not found: %s", tx.From)
	}

	if account.Nonce != tx.Nonce {
		return fmt.Errorf("invalid nonce")
	}

	if account.Balance < tx.Amount {
		return fmt.Errorf("insufficient balance")
	}

	return nil
}
