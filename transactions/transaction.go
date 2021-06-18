package transactions

import (
	"bytes"
	"crypto/sha256"
	"digitalWallet/utils"
	"encoding/gob"
	"fmt"
)

// Transaction defines transaction model
type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxOutput struct {
	Value int
	// Value would be representative of the amount of coins in a transaction
	PubKey string
	// The Pubkey is needed to "unlock" any coins within an Output. This indicated that YOU are the one that sent it.
	// You are indentifiable by your PubKey
	// PubKey in this iteration will be very straightforward, however in an actual application this is a more complex algorithm
}

// TxInput is representative of a reference to a previous TxOutput
type TxInput struct {
	ID []byte
	// ID will find the Transaction that a specific output is inside of
	Out int
	// Out will be the index of the specific output we found within a transaction.
	// For example if a transaction has 4 outputs, we can use this "Out" field to specify which output we are looking for
	Sig string
	// This would be a script that adds data to an outputs' PubKey
	// however for this tutorial the Sig will be indentical to the PubKey.
}

// Defines constants
const (
	Reward = 100
)

// Initializes initial transaction
func CoinbaseTxn(toAddress, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", toAddress)
	}
	//Since this is the "first" transaction of the block, it has no previous output to reference.
	//This means that we initialize it with no ID, and it's OutputIndex is -1
	txIn := TxInput{[]byte{}, -1, data}

	txOut := TxOutput{Reward, toAddress}

	tx := Transaction{nil, []TxInput{txIn}, []TxOutput{txOut}}

	return &tx
}

// Encodes and hash Transactions ID
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	utils.HandleError(err)

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]

}

// Checks that the Sig is correct
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

// Checks that the PubKey is correct
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}

// Checks if it was a coinbase transaction
func (tx *Transaction) IsCoinbase() bool {
	//This checks a transaction and will only return true if it is a newly minted "coin"
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

//func NewTransaction(from, to string, amount int, c *blockchain.BlockChain) *Transaction {
//	var inputs []TxInput
//	var outputs []TxOutput
//
//	// Finds Spendable Outputs
//	acc, validOutputs := c.FindSpendableOutputs(from, amount)
//
//	// Checks if there is enough money to send the amount
//	if acc < amount {
//		log.Panic("Error: Not enough funds!")
//	}
//
//	// Makes inputs that point to the outputs being spent
//	for txid, outs := range validOutputs {
//		txID, err := hex.DecodeString(txid)
//		utils.HandleError(err)
//
//		for _, out := range outs {
//			input := TxInput{txID, out, from}
//			inputs = append(inputs, input)
//		}
//	}
//
//	outputs = append(outputs, TxOutput{amount, to})
//
//	// Make new outputs from the difference
//	if acc > amount {
//		outputs = append(outputs, TxOutput{acc - amount, from})
//	}
//
//	// Initializes a new transaction with all the new inputs and outputs
//	tx := Transaction{nil, inputs, outputs}
//
//	// Sets a new ID, and returns it
//	tx.SetID()
//
//	return &tx
//}
