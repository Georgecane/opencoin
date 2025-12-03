package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
		fmt.Println("Starting OpenCoin node...")
		// TODO: Initialize and start the node
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
		fmt.Println("Initializing node...")
		// TODO: Create config, home directory, generate keys
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
		fmt.Printf("Adding key: %s\n", args[0])
		// TODO: Generate Dilithium key pair and save
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
		fmt.Printf("Querying account: %s\n", args[0])
		// TODO: Query account state
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
		fmt.Printf("Transferring %s to %s\n", args[1], args[0])
		// TODO: Create and broadcast transfer transaction
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(genesisCmd)
	RootCmd.AddCommand(keysCmd)
	RootCmd.AddCommand(queryCmd)
	RootCmd.AddCommand(txCmd)

	keysCmd.AddCommand(keysAddCmd)

	queryCmd.AddCommand(queryAccountCmd)

	txCmd.AddCommand(txTransferCmd)
}
