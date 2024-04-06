package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

const (
	dbPath = "./database.json"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type DB struct {
	path string
	mu   sync.Mutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
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
			Users:  map[int]User{},
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
		ID:   len(dbs.Chirps) + 1,
	}
	dbs.Chirps[chirp.ID] = chirp
	err = db.writeDB(*dbs)
	if err != nil {
		return nil, err
	}
	return &chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	chirps := []Chirp{}
	for _, val := range dbs.Chirps {
		chirps = append(chirps, val)
	}
	return chirps, nil
}

func (db *DB) GetChirp(id int) (*Chirp, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	val, ok := dbs.Chirps[id]
	if !ok {
		return nil, errors.New("Chirp not found in database")
	}

	return &val, nil
}

func (db *DB) CreateUser(email string) (*User, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	user := User{
		Email: email,
		ID:    len(dbs.Users) + 1,
	}
	dbs.Users[user.ID] = user
	err = db.writeDB(*dbs)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) GetUser(id int) (*User, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	val, ok := dbs.Users[id]
	if !ok {
		return nil, errors.New("User not found in database")
	}
	return &val, nil
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
