package blockchain

import (
	_ "digitalWallet/utils"
	"github.com/dgraph-io/badger/v3"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
)

var (
	db     *badger.DB
	dbFile = "/MANIFEST"
)

func init() {
	// Load .env file
	e := godotenv.Load()
	if e != nil {
		log.Println(e)
	}
	// Ensures directory path
	EnsureDir(os.Getenv("BADGE_DB"))

	// Opens database connection
	//OpenDB()
}

// Ensure Dir Checks if directory(ies) exists
// If not creates new directory(ies)
func EnsureDir(fileName string) {

	dirName := filepath.Dir(fileName)
	if _, err := os.Stat(dirName); err != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			log.Panic(merr)
		}
	}
}

// OpenDB opens database connection
//func OpenDB() {
//
//	opts := badger.DefaultOptions(os.Getenv("BADGE_DB"))
//
//	// Opens the Badger database located in the ./tmp/blocks directory.
//	// It creates if it doesn't exist.
//	conn, err := badger.Open(opts)
//	utils.HandleError(err)
//
//	db = conn
//
//	defer conn.Close()
//
//}

//func GetDB()  *badger.DB{
//	return db
//}

// DBExists checks to see if database has been initialized
func DBExists() bool {
	if _, err := os.Stat(os.Getenv("BADGE_DB") + dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}
