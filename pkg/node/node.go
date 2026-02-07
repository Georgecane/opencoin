package node

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/georgecane/opencoin/pkg/config"
	"github.com/georgecane/opencoin/pkg/consensus"
	"github.com/georgecane/opencoin/pkg/contracts"
	"github.com/georgecane/opencoin/pkg/crypto"
	"github.com/georgecane/opencoin/pkg/genesis"
	"github.com/georgecane/opencoin/pkg/mempool"
	"github.com/georgecane/opencoin/pkg/p2p"
	"github.com/georgecane/opencoin/pkg/rc"
	"github.com/georgecane/opencoin/pkg/state"
	"github.com/georgecane/opencoin/pkg/tx"
	"github.com/georgecane/opencoin/pkg/types"
)

// Node is the main runtime.
type Node struct {
	cfg       *config.NodeConfig
	store     *state.Store
	dag       *state.DAG
	state     *state.State
	contracts *contracts.ContractEngine
	mempool   *mempool.Mempool
	dpos      *consensus.DPoS
	consensus *consensus.Engine
	p2p       *p2p.P2P
	httpSrv   *http.Server
	genesis   *genesis.Genesis
}

// New creates a new node.
func New(cfg *config.NodeConfig) (*Node, error) {
	return &Node{cfg: cfg}, nil
}

// Start initializes and starts the node components.
func (n *Node) Start(ctx context.Context) error {
	store, err := state.OpenStore(n.cfg.HomeDir)
	if err != nil {
		return err
	}
	n.store = store
	n.dag = state.NewDAG()

	genPath := filepath.Join(n.cfg.HomeDir, "config", "genesis.json")
	gen, err := genesis.Load(genPath)
	if err != nil {
		return fmt.Errorf("load genesis: %w", err)
	}
	n.genesis = gen

	rcParams := rc.Params{
		Alpha:      gen.RCParams.Alpha,
		Beta:       gen.RCParams.Beta,
		CSize:      gen.RCParams.CSize,
		CCompute:   gen.RCParams.CCompute,
		CStorage:   gen.RCParams.CStorage,
		MaxSkewSec: gen.RCParams.MaxSkewSec,
		WindowN:    gen.RCParams.WindowN,
	}
	n.state = state.NewState(store, n.dag, rcParams)
	n.contracts = contracts.NewContractEngine()
	n.dpos = consensus.NewDPoS(n.cfg.Consensus.MinStake, n.cfg.Consensus.MaxValidators)

	if err := n.applyGenesis(); err != nil {
		return err
	}

	coster := &tx.Coster{Params: rcParams, Contracts: n.contracts}
	n.mempool = mempool.New(n.state, coster)

	if n.cfg.Validator.Enabled {
		keyPath := filepath.Join(n.cfg.HomeDir, n.cfg.Validator.PrivateKeyFile)
		kp, err := crypto.LoadKeyPair(keyPath)
		if err != nil {
			return err
		}
		signer := crypto.NewDilithiumSigner(kp.PublicKey, kp.PrivateKey)
		verifier := crypto.NewDilithiumVerifier()
		engine, err := consensus.NewEngine(consensus.Config{
			EpochLength:   n.cfg.Consensus.EpochLength,
			MaxValidators: n.cfg.Consensus.MaxValidators,
			BlockMaxTxs:   1000,
			MinStake:      n.cfg.Consensus.MinStake,
			SlashDouble:   n.cfg.Consensus.SlashDoubleBps,
			JailDouble:    n.cfg.Consensus.JailDouble,
			SlashOffline:  n.cfg.Consensus.SlashOfflineBps,
			JailOffline:   n.cfg.Consensus.JailOffline,
		}, n.state, n.dpos, n.mempool, signer, verifier, nil)
		if err != nil {
			return err
		}
		n.consensus = engine
	}

	p2pNode, err := p2p.New(ctx, p2p.Config{
		ListenAddrs:    []string{toMultiaddr(n.cfg.P2P.ListenAddr)},
		BootstrapPeers: n.cfg.P2P.BootstrapPeers,
	})
	if err != nil {
		return err
	}
	n.p2p = p2pNode

	n.httpSrv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", n.cfg.RPC.Addr, n.cfg.RPC.Port),
		Handler: n.httpHandler(),
	}
	go n.httpSrv.ListenAndServe()
	return nil
}

// Stop stops node services.
func (n *Node) Stop(ctx context.Context) error {
	if n.httpSrv != nil {
		_ = n.httpSrv.Shutdown(ctx)
	}
	if n.store != nil {
		_ = n.store.Close()
	}
	return nil
}

func (n *Node) httpHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/consensus", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

func (n *Node) applyGenesis() error {
	// Initialize accounts.
	for _, acct := range n.genesis.Accounts {
		stateAcct := &types.Account{
			Address: acct.Address,
			Balance: acct.Balance,
			Stake:   acct.Stake,
			RC:      0,
			RCMax:   n.state.RCParams().RCMax(acct.Stake),
		}
		if err := n.store.SetAccount(stateAcct); err != nil {
			return err
		}
	}
	// Initialize validators.
	for _, v := range n.genesis.Validators {
		if err := n.dpos.RegisterValidator(v.Address, v.PublicKey, v.Stake, v.Commission); err != nil {
			return err
		}
	}
	if ts, err := n.store.GetLastTimestamps(); err == nil && len(ts) == 0 {
		_ = n.store.SetLastTimestamps([]int64{n.genesis.GenesisTime.Unix()})
	}
	return nil
}

func toMultiaddr(addr string) string {
	if strings.HasPrefix(addr, "/") {
		return addr
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "/ip4/0.0.0.0/tcp/26656"
	}
	return fmt.Sprintf("/ip4/%s/tcp/%s", host, port)
}
