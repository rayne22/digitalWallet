package blockchain

import (
	"bytes"
	"crypto/sha256"
	"digitalWallet/utils"
	"fmt"
	"math"
	"math/big"
)

// ProofOfWork
type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

// Defines constants
const (
	Difficulty = 12
)

// NewProofOfWork initializes ProofOfWork struct
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty))

	pow := &ProofOfWork{b, target}

	return pow
}

// InitNonce initiates nonce
func (pow *ProofOfWork) InitNonce(nonce int) []byte {

	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.HashTransactions(),
			utils.ToHex(int64(nonce)),
			utils.ToHex(int64(Difficulty))},
		[]byte{},
	)

	return data
}

// RunPoW runs Proof of Work algorithm
func (pow *ProofOfWork) RunPoW() (int, []byte) {

	var intHash big.Int
	var hash [32]byte

	nonce := 0

	// This essentially is an infinite loop
	// Due to how large MaxInt64 is

	for nonce < math.MaxInt64 {
		data := pow.InitNonce(nonce)
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash)
		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}

	fmt.Println()

	return nonce, hash[:]

}

// Validates Proof of Work algorithm
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitNonce(pow.Block.Nonce)

	hash := sha256.Sum256(data)

	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}
