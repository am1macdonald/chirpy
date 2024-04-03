package database

import (
	"os"
	"sync"

	"github.com/am1macdonald/chirpy/internal/chirps"
)

const (
	dbPath = "../../database.json"
)

type DB struct {
	path string
	mu   *sync.Mutex
}

type DBStructure struct {
	Chirps map[int]chirps.Chirp
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(dbPath)
	if err != nil {
		_, err = os.Create(dbPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewDB() (*DB, error) {
	db := DB{}
	err := db.ensureDB()
	if err != nil {
		return nil, err
	}
	return &db, nil
}
