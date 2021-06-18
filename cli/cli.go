package cli

import (
	"digitalWallet/blockchain"
	"digitalWallet/services"
	"digitalWallet/transactions"
	"digitalWallet/wallet"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
)

// CommandLine defines commandline
type CommandLine struct{}

// PrintUsage will display what options are available to the user
func (cli *CommandLine) PrintUsage() {
	fmt.Println("Usage: ")

	fmt.Println("getbalance -address ADDRESS - get balance for ADDRESS")

	fmt.Println("createblockchain -address ADDRESS creates a blockchain and rewards the mining fee")

	fmt.Println("printchain - Prints the blocks in the chain")

	fmt.Println("send -from FROM -to TO -amount AMOUNT - Send amount of coins from one address to another")

	fmt.Println("createwallet - Creates a new wallet")

	fmt.Println("listaddresses - Lists the addresses in the wallet file")
}

// ValidateArgs ensures the cli was given valid input
func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		//go exit will exit the application by shutting down the goroutine
		// if you were to use os.exit you might corrupt the data
		runtime.Goexit()
	}
}

// CreateBlockChain creates blockchain
func (cli *CommandLine) CreateBlockChain(address string) {
	// Initializes initial block chain
	newChain := blockchain.InitBlockChain(address)
	_ = newChain.Database.Close()
	fmt.Println("Finished creating chain")
}

// GetBalance gets account balance
func (cli *CommandLine) GetBalance(address string) {

	// Adds to an existing blockchain
	chain := blockchain.ContinueBlockChain(address)
	defer chain.Database.Close()

	balance := 0
	// Finds all transaction outputs
	UTXOs := chain.FindUTXO(address)
	fmt.Println("TEST", UTXOs)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

// Send sends money to address
func (cli *CommandLine) Send(from, to string, amount int) {

	// Adds to an existing blockchain
	chain := blockchain.ContinueBlockChain(from)
	defer chain.Database.Close()

	tx := services.Txn.NewTransaction(from, to, amount, chain)

	// Adds a new block to the block chain
	chain.AddBlock([]*transactions.Transaction{tx})
	fmt.Println("Success!")
}

// PrintChain will display the entire contents of the blockchain
func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()
	iterator := chain.Iterator()

	for {
		block := iterator.Next()
		fmt.Printf("Previous hash: %x\n", block.PrevHash)
		fmt.Printf("hash: %x\n", block.Hash)
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("Pow: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
		// This works because the Genesis block has no PrevHash to point to.
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

// ListAddresses lists all addresses
func (cli *CommandLine) ListAddresses() {
	wallets, _ := wallet.CreateWallets()
	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}

}

// CreateWallet creates wallet
func (cli *CommandLine) CreateWallet() {
	wallets, _ := wallet.CreateWallets()
	address := wallets.AddWallet()
	wallets.SaveFile()

	fmt.Printf("New address is: %s\n", address)

}

// Run will start up the command line
func (cli *CommandLine) Run() {
	cli.ValidateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)   // this Cmd is new
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError) // this Cmd is new

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses": // this case statement is new
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet": // this case statement is new
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.PrintUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.GetBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.CreateBlockChain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.Send(*sendFrom, *sendTo, *sendAmount)
	}
	if listAddressesCmd.Parsed() {
		cli.ListAddresses()
	}
	if createWalletCmd.Parsed() {
		cli.CreateWallet()
	}
}
