package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"digitalWallet/utils"
	"golang.org/x/crypto/ripemd160"
	"log"
)

// Wallet defines wallet model
type Wallet struct {
	//ecdsa = Elipitic  curve digital signature algorithm
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// Defines constants
const (
	checksumLength = 4
	//hexadecimal representation of 0
	version = byte(0x00)

	walletFile = "./tmp/wallets.data"
)

// NewKeyPair generates new key pair
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	// Generates key
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pub
}

// PublicKeyHash hashes the public key
func PublicKeyHash(publicKey []byte) []byte {
	hashedPublicKey := sha256.Sum256(publicKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(hashedPublicKey[:])
	if err != nil {
		log.Panic(err)
	}
	publicRipeMd := hasher.Sum(nil)

	return publicRipeMd
}

// Checksum runs sha256 on the versioned hash twice
// To create a checksum
func Checksum(ripeMdHash []byte) []byte {
	firstHash := sha256.Sum256(ripeMdHash)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}

// Address creates wallet address
func (w *Wallet) Address() []byte {

	// Hashes the public key
	pubHash := PublicKeyHash(w.PublicKey)

	// Creates versioned hash
	versionedHash := append([]byte{version}, pubHash...)

	// Creates Checksum
	checksum := Checksum(versionedHash)

	// Creates final hash
	finalHash := append(versionedHash, checksum...)

	// Creates a wallet address
	address := utils.Base58Encode(finalHash)

	return address
}

// MakeWallet creates a wallet
func MakeWallet() *Wallet {
	privateKey, publicKey := NewKeyPair()
	wallet := Wallet{privateKey, publicKey}
	return &wallet
}

// ValidateAddress validates the address...
func ValidateAddress(address string) bool {
	// Decodes address
	pubKeyHash := utils.Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLength]

	// Runs sha256 on the versioned hash twice To create a checksum
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}
