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
    BlockReward      uint64  `mapstructure:"block_reward"`
    UnbondingPeriod  uint64  `mapstructure:"unbonding_period"`
}

// ValidatorConfig represents validator configuration
type ValidatorConfig struct {
    Enabled        bool   `mapstructure:"enabled"`
    PrivateKeyFile string `mapstructure:"private_key_file"`
    Stake          uint64 `mapstructure:"stake"`
    Commission     uint16 `mapstructure:"commission"`
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
            BlockReward:      1_000_000,   // 1 OCN per block
            UnbondingPeriod:  259200,      // 3 days in seconds
        },

        Validator: ValidatorConfig{
            Enabled:        false,
            PrivateKeyFile: "config/validator_key.json",
            Stake:          1_000_000_000, // 1000 OCN
            Commission:     1000,          // 10%
        },
    }
}
