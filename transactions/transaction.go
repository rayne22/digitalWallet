package transactions

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"digitalWallet/utils"
	"digitalWallet/wallet"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

// Transaction defines transaction model
type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

// TxInput is representative of a reference to a previous TxOutput
type TxInput struct {
	ID        []byte
	Out       int
	Signature []byte
	PubKey    []byte
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
	txIn := TxInput{[]byte{}, -1, nil, []byte(data)}

	txOut := NewTXOutput(Reward, toAddress)

	tx := Transaction{nil, []TxInput{txIn}, []TxOutput{*txOut}}

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

// UsesKey hashes input public key
// Compares public key
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	// hashes the public key
	lockingHash := wallet.PublicKeyHash(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

// Lock locks the address
func (out *TxOutput) Lock(address []byte) {
	// Decodes address
	pubKeyHash := utils.Base58Decode(address)

	// Removes the Versioned hash and the Checksum hash from the public key hash
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey checks if the output is locked with a key
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTXOutput converts address into bytes
// Populates the transaction out put with a public key hash
func NewTXOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	// Locks the address
	txo.Lock([]byte(address))

	return txo
}

// Checks if it was a coinbase transaction
func (tx *Transaction) IsCoinbase() bool {
	//This checks a transaction and will only return true if it is a newly minted "coin"
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

// Hash hashes a transaction copy
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Sign signs the transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	// checks if it's coin base
	if tx.IsCoinbase() {
		return
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	// Creates a transaction copy
	txCopy := tx.TrimmedCopy()

	for inId, in := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		utils.HandleError(err)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inId].Signature = signature

	}
}

// Verify verifies the transaction
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("Previous transaction not correct")
		}
	}

	// creates a transaction copy
	txCopy := tx.TrimmedCopy()

	curve := elliptic.P256()

	for inId, in := range tx.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r := big.Int{}
		s := big.Int{}

		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}

//TrimmedCopy creates a transaction copy
func (tx Transaction) TrimmedCopy() Transaction {

	var inputs []TxInput
	var outputs []TxOutput

	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{out.Value, out.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy

}

// String creates strings for the CLI
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:     %x", input.ID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
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
