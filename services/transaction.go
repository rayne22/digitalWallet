package services

import (
	"digitalWallet/blockchain"
	"digitalWallet/transactions"
	"digitalWallet/utils"
	"encoding/hex"
	"log"
)

// newTransactionServiceInterface interface keeps the new transaction function
type newTransactionServiceInterface interface {
	NewTransaction(from, to string, amount int, c *blockchain.BlockChain) *transactions.Transaction
}

// Instantiates type transactionService
type newTransactionService struct{}

func (n newTransactionService) NewTransaction(from, to string, amount int, c *blockchain.BlockChain) *transactions.Transaction {
	var inputs []transactions.TxInput
	var outputs []transactions.TxOutput

	// Finds Spendable Outputs
	acc, validOutputs := c.FindSpendableOutputs(from, amount)

	// Checks if there is enough money to send the amount
	if acc < amount {
		log.Panic("Error: Not enough funds!")
	}

	// Makes inputs that point to the outputs being spent
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		utils.HandleError(err)

		for _, out := range outs {
			input := transactions.TxInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, transactions.TxOutput{amount, to})

	// Make new outputs from the difference
	if acc > amount {
		outputs = append(outputs, transactions.TxOutput{acc - amount, from})
	}

	// Initializes a new transaction with all the new inputs and outputs
	tx := transactions.Transaction{nil, inputs, outputs}

	// Sets a new ID, and returns it
	tx.SetID()

	return &tx
}

var (
	// Instantiates the transaction services
	Txn newTransactionServiceInterface = &newTransactionService{}
)
