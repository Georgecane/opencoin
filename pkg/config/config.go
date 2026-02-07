package config

import (
    "time"
)

// NodeConfig represents the configuration for a node
type NodeConfig struct {
    Moniker     string        `mapstructure:"moniker"`
    ChainID     string        `mapstructure:"chain_id"`
    HomeDir     string        `mapstructure:"home_dir"`
    LogLevel    string        `mapstructure:"log_level"`
    LogFormat   string        `mapstructure:"log_format"`

    P2P         P2PConfig     `mapstructure:"p2p"`
    RPC         RPCConfig     `mapstructure:"rpc"`
    Consensus   ConsensusConfig `mapstructure:"consensus"`
    Validator   ValidatorConfig `mapstructure:"validator"`
    RC          RCConfig      `mapstructure:"rc"`
    Governance  GovernanceConfig `mapstructure:"governance"`
}

// P2PConfig represents P2P network configuration
type P2PConfig struct {
    ListenAddr     string `mapstructure:"listen_addr"`
    MaxConnInbound uint16 `mapstructure:"max_conn_inbound"`
    MaxConnOutbound uint16 `mapstructure:"max_conn_outbound"`
    MaxPeers       uint16 `mapstructure:"max_peers"`
    PersistentPeers string `mapstructure:"persistent_peers"`
    Seeds          string `mapstructure:"seeds"`
    PrivateKeyFile string `mapstructure:"private_key_file"`
    BootstrapPeers []string `mapstructure:"bootstrap_peers"`
}

// RPCConfig represents RPC server configuration
type RPCConfig struct {
    Addr             string `mapstructure:"addr"`
    Port             uint16 `mapstructure:"port"`
    MaxBodyBytes     int64  `mapstructure:"max_body_bytes"`
    MaxOpenConnections int  `mapstructure:"max_open_connections"`
}

// ConsensusConfig represents consensus parameters
type ConsensusConfig struct {
    TimeoutPropose   time.Duration `mapstructure:"timeout_propose"`
    TimeoutPrevote   time.Duration `mapstructure:"timeout_prevote"`
    TimeoutPrecommit time.Duration `mapstructure:"timeout_precommit"`
    TimeoutCommit    time.Duration `mapstructure:"timeout_commit"`

    MinStake         uint64  `mapstructure:"min_stake"`
    MaxValidators    uint32  `mapstructure:"max_validators"`
    EpochLength      uint64  `mapstructure:"epoch_length"`
    BlockReward      uint64  `mapstructure:"block_reward"`
    UnbondingPeriod  uint64  `mapstructure:"unbonding_period"`
    SlashDoubleBps      uint64  `mapstructure:"slash_double_bps"`
    JailDouble          uint64  `mapstructure:"jail_double"`
    SlashOfflineBps     uint64  `mapstructure:"slash_offline_bps"`
    JailOffline         uint64  `mapstructure:"jail_offline"`
}

// ValidatorConfig represents validator configuration
type ValidatorConfig struct {
    Enabled        bool   `mapstructure:"enabled"`
    PrivateKeyFile string `mapstructure:"private_key_file"`
    Stake          uint64 `mapstructure:"stake"`
    Commission     uint16 `mapstructure:"commission"`
}

// RCConfig represents RC parameters.
type RCConfig struct {
    Alpha    uint64 `mapstructure:"alpha"`
    Beta     uint64 `mapstructure:"beta"`
    CSize    uint64 `mapstructure:"c_size"`
    CCompute uint64 `mapstructure:"c_compute"`
    CStorage uint64 `mapstructure:"c_storage"`
    MaxSkewSec int64 `mapstructure:"max_skew_sec"`
    WindowN  int   `mapstructure:"window_n"`
}

// GovernanceConfig represents governance parameters.
type GovernanceConfig struct {
    VotingPeriodEpochs uint64 `mapstructure:"voting_period_epochs"`
    QuorumPercent      uint64 `mapstructure:"quorum_percent"`
    ThresholdPercent   uint64 `mapstructure:"threshold_percent"`
    TimelockEpochs     uint64 `mapstructure:"timelock_epochs"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *NodeConfig {
    return &NodeConfig{
        Moniker:   "opencoin-node",
        ChainID:   "opencoin-1",
        HomeDir:   "$HOME/.opencoin",
        LogLevel:  "info",
        LogFormat: "json",

        P2P: P2PConfig{
            ListenAddr:      "0.0.0.0:26656",
            MaxConnInbound:  100,
            MaxConnOutbound: 32,
            MaxPeers:        200,
            PrivateKeyFile:  "config/node_key.json",
            BootstrapPeers:  []string{},
        },

        RPC: RPCConfig{
            Addr:               "0.0.0.0",
            Port:               26657,
            MaxBodyBytes:       1000000,
            MaxOpenConnections: 900,
        },

        Consensus: ConsensusConfig{
            TimeoutPropose:   3 * time.Second,
            TimeoutPrevote:   1 * time.Second,
            TimeoutPrecommit: 1 * time.Second,
            TimeoutCommit:    1 * time.Second,

            MinStake:         100_000_000, // 100 OCN
            MaxValidators:    100,
            EpochLength:      10_000,
            BlockReward:      1_000_000,   // 1 OCN per block
            UnbondingPeriod:  259200,      // 3 days in seconds
            SlashDoubleBps:      500,       // 5%
            JailDouble:          10,        // 10 epochs
            SlashOfflineBps:     10,        // 0.1%
            JailOffline:         2,         // 2 epochs
        },

        Validator: ValidatorConfig{
            Enabled:        false,
            PrivateKeyFile: "config/validator_key.json",
            Stake:          1_000_000_000, // 1000 OCN
            Commission:     1000,          // 10%
        },
        RC: RCConfig{
            Alpha:    1000,
            Beta:     1,
            CSize:    1,
            CCompute: 1,
            CStorage: 50,
            MaxSkewSec: 30,
            WindowN:  11,
        },
        Governance: GovernanceConfig{
            VotingPeriodEpochs: 2,
            QuorumPercent:      33,
            ThresholdPercent:   50,
            TimelockEpochs:     1,
        },
    }
}
