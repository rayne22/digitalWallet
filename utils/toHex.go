package utils

import (
	"bytes"
	"encoding/binary"
)

// ToHex converts number to byte
func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)

	err := binary.Write(buff, binary.BigEndian, num)

	HandleError(err)

	return buff.Bytes()
}
