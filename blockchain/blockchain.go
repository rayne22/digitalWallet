package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"digitalWallet/transactions"
	"digitalWallet/utils"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"os"
	"runtime"
)

// BlockChain defines a blockchain
type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

// Defines blockchain iterator
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// Defines constants
const (

	// This is arbitrary data for our genesis block
	genesisData = "First Transaction from Genesis"
)

// AddBlock adds a new block to the block chain
func (c *BlockChain) AddBlock(transactions []*transactions.Transaction) {
	var lastHash []byte

	// Views last hash in the blockchain
	err := c.Database.View(func(txn *badger.Txn) error {
		// Gets Item based on key
		item, err := txn.Get([]byte("lh"))
		utils.HandleError(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		utils.HandleError(err)
		return err
	})
	utils.HandleError(err)

	// Creates new block
	newBlock := CreateBlock(transactions, lastHash)

	// Updates transaction
	err = c.Database.Update(func(transaction *badger.Txn) error {

		// Adds new hash to transaction
		err := transaction.Set(newBlock.Hash, newBlock.Serialize())
		utils.HandleError(err)

		// Adds last hash to transaction
		err = transaction.Set([]byte("lh"), newBlock.Hash)

		c.LastHash = newBlock.Hash
		return err
	})
	utils.HandleError(err)
}

// InitBlockChain initializes initial block chain
func InitBlockChain(address string) *BlockChain {
	//var lastHash []byte

	//if DBExists() {
	//	fmt.Println("blockchain already exists")
	//	runtime.Goexit()
	//}
	//
	//
	//// Updates transaction
	//_ = GetDB().Update(func(txn *badger.Txn) error {
	//
	//	// Initializes initial transaction
	//	cbtx := transactions.CoinbaseTxn(address, genesisData)
	//	genesis := Genesis(cbtx)
	//	fmt.Println("Genesis Created")
	//
	//	// Adds initial hash to transaction
	//	err := txn.Set(genesis.Hash, genesis.Serialize())
	//	utils.HandleError(err)
	//
	//	// Adds last hash to transaction
	//	err = txn.Set([]byte("lh"), genesis.Hash)
	//
	//	lastHash = genesis.Hash
	//
	//	return err
	//
	//})
	//
	//blockchain := BlockChain{lastHash, GetDB()}
	//return &blockchain

	var lastHash []byte

	if DBExists() {
		fmt.Println("blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(os.Getenv("BADGE_DB"))
	db, err := badger.Open(opts)
	utils.HandleError(err)

	err = db.Update(func(txn *badger.Txn) error {

		cbtx := transactions.CoinbaseTxn(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis Created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		utils.HandleError(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err

	})

	utils.HandleError(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

// Adds to an existing blockchain
func ContinueBlockChain(address string) *BlockChain {
	//if !DBExists() {
	//	fmt.Println("No blockchain found, please create one first")
	//	runtime.Goexit()
	//}
	//
	//
	//var lastHash []byte
	//
	//
	//opts := badger.DefaultOptions(os.Getenv("BADGE_DB"))
	//db, err := badger.Open(opts)
	//utils.HandleError(err)
	//
	//// Updates transaction
	//err = db.Update(func(txn *badger.Txn) error {
	//
	//	// Gets last hash from transaction
	//	item, err := txn.Get([]byte("lh"))
	//	fmt.Println("ALALALALALA")
	//	utils.HandleError(err)
	//	err = item.Value(func(val []byte) error {
	//		lastHash = val
	//		return nil
	//	})
	//	utils.HandleError(err)
	//	return err
	//})
	//
	//
	//utils.HandleError(err)
	//
	//chain := BlockChain{lastHash, GetDB()}
	//return &chain

	if DBExists() == false {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(os.Getenv("BADGE_DB"))

	db, err := badger.Open(opts)
	utils.HandleError(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		utils.HandleError(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		return err
	})
	utils.HandleError(err)

	chain := BlockChain{lastHash, db}

	return &chain
}

// Initiates blockchain iterator
func (c *BlockChain) Iterator() *BlockChainIterator {
	iterator := BlockChainIterator{c.LastHash, c.Database}

	return &iterator
}

// Next calls the next block in the chain
func (iterator *BlockChainIterator) Next() *Block {
	var block *Block

	// Views current hash in the blockchain
	err := iterator.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iterator.CurrentHash)
		utils.HandleError(err)

		err = item.Value(func(val []byte) error {
			block = Deserialize(val)
			return nil
		})
		utils.HandleError(err)
		return err
	})
	utils.HandleError(err)

	iterator.CurrentHash = block.PrevHash

	return block
}

// Finds unspent all transactions
func (c *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []transactions.Transaction {
	var unspentTxs []transactions.Transaction

	spentTXNOs := make(map[string][]int)

	// Initiates blockchain iterator
	iter := c.Iterator()

	for {
		// calls the next block in the chain
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXNOs[txID] != nil {
					for _, spentOut := range spentTXNOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXNOs[inTxID] = append(spentTXNOs[inTxID], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspentTxs
}

// Finds all transaction outputs
func (c *BlockChain) FindUTXO(pubKeyHash []byte) []transactions.TxOutput {
	var UTXOs []transactions.TxOutput

	// Finds unspent all transactions
	unspentTransactions := c.FindUnspentTransactions(pubKeyHash)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// FindSpendableOutputs finds all spendable outputs from transactions
func (c *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)

	// Finds unspent all transactions
	unspentTxs := c.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOuts
}

// FindTransaction finds transaction based on ID
func (c *BlockChain) FindTransaction(ID []byte) (transactions.Transaction, error) {
	iter := c.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return transactions.Transaction{}, errors.New("Transaction does not exist")
}

// SignTransaction signs the transaction
func (c *BlockChain) SignTransaction(tx *transactions.Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]transactions.Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := c.FindTransaction(in.ID)
		utils.HandleError(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies transaction
func (c *BlockChain) VerifyTransaction(tx *transactions.Transaction) bool {
	prevTXs := make(map[string]transactions.Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := c.FindTransaction(in.ID)
		utils.HandleError(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
