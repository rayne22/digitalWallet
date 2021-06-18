package blockchain

import (
	"bytes"
	"digitalWallet/utils"
	"encoding/gob"
)

// Serialize converts struct to byte
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	utils.HandleError(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)

	utils.HandleError(err)

	return &block
}
