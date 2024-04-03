package database

import (
	"encoding/json"
	"os"
	"sync"
)

const (
	dbPath = "../../database.json"
)

type Chirp struct {
	Body string `json:"body"`
}

type DB struct {
	path string
	mu   sync.Mutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(dbPath)
	if err != nil {
		_, err = os.Create(dbPath)
		if err != nil {
			return err
		}
		err = db.writeDB(DBStructure{
			Chirps: map[int]Chirp{},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) loadDB() (*DBStructure, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	bytes, err := os.ReadFile(db.path)
	if err != nil {
		return nil, err
	}
	dbs := DBStructure{}
	err = json.Unmarshal(bytes, &dbs)
	if err != nil {
		return nil, err
	}
	return &dbs, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	bytes, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, bytes, os.FileMode(int(0777)))
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) CreateChirp(body string) (*Chirp, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	chirp := Chirp{
		Body: body,
	}
	dbs.Chirps[len(dbs.Chirps)+1] = chirp
	err = db.writeDB(*dbs)
	if err != nil {
		return nil, err
	}
	return &chirp, nil
}

func NewDB() (*DB, error) {
	db := DB{
		path: dbPath,
	}
	err := db.ensureDB()
	if err != nil {
		return nil, err
	}
	return &db, nil
}
