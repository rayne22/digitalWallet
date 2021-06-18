package main

import (
	"digitalWallet/cli"
	"os"
)

func main() {
	defer os.Exit(0)

	//blockchain.OpenDB()

	cmd := cli.CommandLine{}
	cmd.Run()
}
