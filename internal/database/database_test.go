package database_test

import (
	"os"
	"testing"

	"github.com/am1macdonald/chirpy/internal/database"
)

// creates a new database_test
func TestCreateDatabase(t *testing.T) {
	os.Remove("../../database.json")
	db, err := database.NewDB()
	if db == nil || err != nil {
		t.Fatal("Failed: create method")
	}
	_, err = os.Open("../../database.json")
	if err != nil {
		t.Fatalf("Failed: %s", err.Error())
	}
}

func TestCreateChirp(t *testing.T) {

}
