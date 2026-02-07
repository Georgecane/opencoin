package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/georgecane/opencoin/pkg/config"
	"github.com/georgecane/opencoin/pkg/crypto"
	"github.com/georgecane/opencoin/pkg/genesis"
	"github.com/georgecane/opencoin/pkg/node"
	"github.com/georgecane/opencoin/pkg/state"
	"github.com/georgecane/opencoin/pkg/tx"
	"github.com/georgecane/opencoin/pkg/types"
)

var RootCmd = &cobra.Command{
	Use:   "opencoin",
	Short: "OpenCoin - A gas-less, post-quantum blockchain",
	Long: `OpenCoin is a delegated proof-of-stake blockchain with post-quantum cryptography
and smart contract support, designed to be gas-less.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to OpenCoin!")
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the OpenCoin node",
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := cmd.Flags().GetString("home")
		cfgPath := filepath.Join(home, "config", "config.json")
		cfg, err := config.Load(cfgPath)
		if err != nil {
			fmt.Println("failed to load config:", err)
			os.Exit(1)
		}
		n, err := node.New(cfg)
		if err != nil {
			fmt.Println("failed to create node:", err)
			os.Exit(1)
		}
		if err := n.Start(cmd.Context()); err != nil {
			fmt.Println("failed to start node:", err)
			os.Exit(1)
		}
		select {}
	},
}

var genesisCmd = &cobra.Command{
	Use:   "genesis",
	Short: "Manage genesis configuration",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new node",
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := cmd.Flags().GetString("home")
		if err := os.MkdirAll(filepath.Join(home, "config"), 0o700); err != nil {
			fmt.Println("failed to create home:", err)
			os.Exit(1)
		}
		cfg := config.DefaultConfig()
		cfg.HomeDir = home
		cfgPath := filepath.Join(home, "config", "config.json")
		if err := config.Save(cfgPath, cfg); err != nil {
			fmt.Println("failed to save config:", err)
			os.Exit(1)
		}
		gen := genesis.DefaultGenesis()
		genPath := filepath.Join(home, "config", "genesis.json")
		if err := gen.Save(genPath); err != nil {
			fmt.Println("failed to save genesis:", err)
			os.Exit(1)
		}
		fmt.Println("Initialized node at", home)
	},
}

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage account keys",
}

var keysAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new key",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := cmd.Flags().GetString("home")
		keyDir := filepath.Join(home, "config", "keys")
		if err := os.MkdirAll(keyDir, 0o700); err != nil {
			fmt.Println("failed to create key dir:", err)
			os.Exit(1)
		}
		kp, err := crypto.GenerateEd25519()
		if err != nil {
			fmt.Println("failed to generate key:", err)
			os.Exit(1)
		}
		path := filepath.Join(keyDir, args[0]+".json")
		if err := crypto.SaveEd25519(path, kp); err != nil {
			fmt.Println("failed to save key:", err)
			os.Exit(1)
		}
		addr, _ := crypto.AddressFromPubKey(kp.PublicKey)
		fmt.Printf("Created key %s address %s\n", args[0], addr)
	},
}

var keysValidatorCmd = &cobra.Command{
	Use:   "validator [name]",
	Short: "Add a new validator key (Dilithium)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := cmd.Flags().GetString("home")
		keyDir := filepath.Join(home, "config", "validator-keys")
		if err := os.MkdirAll(keyDir, 0o700); err != nil {
			fmt.Println("failed to create key dir:", err)
			os.Exit(1)
		}
		kp, err := crypto.GenerateKeyPair()
		if err != nil {
			fmt.Println("failed to generate validator key:", err)
			os.Exit(1)
		}
		path := filepath.Join(keyDir, args[0]+".json")
		if err := crypto.SaveKeyPair(path, kp); err != nil {
			fmt.Println("failed to save validator key:", err)
			os.Exit(1)
		}
		addr, _ := crypto.AddressFromPubKey(kp.PublicKey)
		fmt.Printf("Created validator key %s address %s\n", args[0], addr)
	},
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query blockchain state",
}

var queryAccountCmd = &cobra.Command{
	Use:   "account [address]",
	Short: "Query account balance and nonce",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := cmd.Flags().GetString("home")
		store, err := state.OpenStore(home)
		if err != nil {
			fmt.Println("failed to open state:", err)
			os.Exit(1)
		}
		defer store.Close()
		acct, err := store.GetAccount(types.Address(args[0]))
		if err != nil {
			fmt.Println("query failed:", err)
			os.Exit(1)
		}
		if acct == nil {
			fmt.Println("account not found")
			return
		}
		fmt.Printf("address=%s balance=%d nonce=%d stake=%d rc=%d\n", acct.Address, acct.Balance, acct.Nonce, acct.Stake, acct.RC)
	},
}

var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "Broadcast transactions",
}

var txTransferCmd = &cobra.Command{
	Use:   "transfer [to] [amount]",
	Short: "Transfer tokens",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := cmd.Flags().GetString("home")
		fromKey, _ := cmd.Flags().GetString("from")
		nonce, _ := cmd.Flags().GetUint64("nonce")
		if fromKey == "" {
			fmt.Println("missing --from")
			os.Exit(1)
		}
		keyPath := filepath.Join(home, "config", "keys", fromKey+".json")
		kp, err := crypto.LoadEd25519(keyPath)
		if err != nil {
			fmt.Println("failed to load key:", err)
			os.Exit(1)
		}
		to := args[0]
		amount, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			fmt.Println("invalid amount:", err)
			os.Exit(1)
		}
		payload, err := tx.EncodePayload(tx.Transfer{
			To:     types.Address(to),
			Amount: amount,
		})
		if err != nil {
			fmt.Println("failed to encode payload:", err)
			os.Exit(1)
		}
		fromAddr, _ := crypto.AddressFromPubKey(kp.PublicKey)
		txn := &types.Transaction{
			From:    types.Address(fromAddr),
			To:      types.Address(to),
			Nonce:   nonce,
			Payload: payload,
		}
		signBytes, err := tx.SigningBytes(txn)
		if err != nil {
			fmt.Println("failed to sign:", err)
			os.Exit(1)
		}
		sig, err := crypto.SignEd25519(kp.PrivateKey, signBytes)
		if err != nil {
			fmt.Println("failed to sign:", err)
			os.Exit(1)
		}
		txn.Signature = sig
		fmt.Printf("tx: from=%s to=%s nonce=%d size=%d\n", txn.From, txn.To, txn.Nonce, len(signBytes))
	},
}

func init() {
	RootCmd.PersistentFlags().String("home", filepath.Join(os.Getenv("USERPROFILE"), ".opencoin"), "node home directory")
	RootCmd.AddCommand(startCmd)
	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(genesisCmd)
	RootCmd.AddCommand(keysCmd)
	RootCmd.AddCommand(queryCmd)
	RootCmd.AddCommand(txCmd)

	keysCmd.AddCommand(keysAddCmd)
	keysCmd.AddCommand(keysValidatorCmd)

	queryCmd.AddCommand(queryAccountCmd)

	txCmd.AddCommand(txTransferCmd)

	txTransferCmd.Flags().String("from", "", "sender key name")
	txTransferCmd.Flags().Uint64("nonce", 0, "transaction nonce")
}
