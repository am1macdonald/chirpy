package database_test

import (
	"os"
	"testing"

	"github.com/am1macdonald/chirpy/internal/database"
)

var (
	db *database.DB
)

func beforeEach() (*database.DB, error) {
	os.Remove("../../database.json")
	return database.NewDB()
}

// creates a new database_test
func TestCreateDatabase(t *testing.T) {
	db, err := beforeEach()
	if db == nil || err != nil {
		t.Fatal("Failed: create method")
	}
	_, err = os.Open("../../database.json")
	if err != nil {
		t.Fatalf("Test 'CreateDatabase' failed: %s", err.Error())
	}
}

// gets a new chirp from the create chirp function & tests reading chirps
func TestCreateChirp(t *testing.T) {
	db, err := beforeEach()
	chirp, err := db.CreateChirp("wow a chirp!")
	if chirp == nil || err != nil {
		t.Fatalf("Test 'CreateChirp' failed: %s", err.Error())
	}
	chirps, err := db.GetChirps()
	if err != nil {
		t.Fatalf("Test 'CreateChirp' failed: %s", err.Error())
	}
	if len(chirps) < 1 {
		t.Fatalf("Test 'CreateChirp' failed: expected one chirp, got %d", len(chirps))
	}
}
