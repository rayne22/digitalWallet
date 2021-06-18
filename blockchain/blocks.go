package blockchain

import (
	"bytes"
	"crypto/sha256"
	"digitalWallet/transactions"
)

// Block defines a block model
type Block struct {
	Hash         []byte
	Transactions []*transactions.Transaction
	PrevHash     []byte
	Nonce        int
}

// CreateBlock creates new block
func CreateBlock(txs []*transactions.Transaction, prevHash []byte) *Block {

	block := &Block{[]byte{}, txs, prevHash, 0}

	// Executes creation of new proof of work
	pow := NewProofOfWork(block)

	// Executes hashing algorithm
	nonce, hash := pow.RunPoW()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// Genesis creates initial Block
func Genesis(coinbase *transactions.Transaction) *Block {
	return CreateBlock([]*transactions.Transaction{coinbase}, []byte{})
}

// HashTransactions hashes all transactions in a block
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}
