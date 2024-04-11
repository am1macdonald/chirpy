package database

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	dbPath = "./database.json"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *User) Validate(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) GetAccessToken(secret string) (string, error) {
	expiry := time.Now().Add(time.Duration(time.Hour * 1))
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiry),
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   strconv.Itoa(u.ID),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func (u *User) GetRefreshToken(secret string) (string, error) {
	expiry := time.Now().Add(time.Duration(time.Hour * 24 * 60))
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiry),
		Issuer:    "chirpy-refresh",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   strconv.Itoa(u.ID),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func (u *User) UpdatePassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

type DB struct {
	path string
	mu   sync.Mutex
}

type DBStructure struct {
	Chirps        map[int]Chirp        `json:"chirps"`
	Users         map[int]User         `json:"users"`
	RevokedTokens map[string]time.Time `json:"revoked_tokens"`
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(dbPath)
	if err != nil {
		_, err = os.Create(dbPath)
		if err != nil {
			return err
		}
		err = db.writeDB(DBStructure{
			Chirps:        map[int]Chirp{},
			Users:         map[int]User{},
			RevokedTokens: map[string]time.Time{},
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

func (db *DB) CreateUser(email string, password string) (*User, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	user, err := db.GetUserByEmail(email)
	if err == nil && user != nil {
		return nil, errors.New("User already exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	if err != nil {
		return nil, err
	}
	user = &User{
		Email:    email,
		ID:       len(dbs.Users) + 1,
		Password: string(hash),
	}
	dbs.Users[user.ID] = *user
	err = db.writeDB(*dbs)
	if err != nil {
		return nil, err
	}
	return user, nil
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

func (db *DB) GetUserByEmail(email string) (*User, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	for _, v := range dbs.Users {
		if v.Email == email {
			return &v, nil
		}
	}
	return nil, errors.New("User not found")
}

func (db *DB) UpdateUser(id int, u *User) (*User, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	dbs.Users[id] = *u
	err = db.writeDB(*dbs)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (db *DB) RevokeToken(token string) error {
	dbs, err := db.loadDB()
	if err != nil {
		return err
	}
	dbs.RevokedTokens[token] = time.Now()
	err = db.writeDB(*dbs)
	if err != nil {
		return err
	}
	return nil

}

func (db *DB) ValidateToken(token string) (bool, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return false, err
	}
	_, ok := dbs.RevokedTokens[token]
	if ok {
		return false, nil
	}
	return true, nil
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
