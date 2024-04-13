package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/am1macdonald/chirpy/internal/database"
)

func (cfg *apiConfig) GetChirpsHandler(w http.ResponseWriter, r *http.Request) {
	author_id := r.URL.Query().Get("author_id")
	sort_order := r.URL.Query().Get("sort")
	chirps, err := db.GetChirps()
	if author_id != "" {
		id, err := strconv.Atoi(author_id)
		if err != nil {
			errorResponse(w, 500, errors.New("bad id"))
			return
		}
		filteredChirps := []database.Chirp{}
		for _, chirp := range chirps {
			if chirp.AuthorID == id {
				filteredChirps = append(filteredChirps, chirp)
			}
		}
		chirps = filteredChirps
	}
	if err != nil {
		jsonResponse(w, 500, err.Error())
		return
	}
	if sort_order == "desc" {
		temp := []database.Chirp{}
		for i := len(chirps) - 1; i >= 0; i-- {
			temp = append(temp, chirps[i])
		}
		chirps = temp
	}
	log.Println(chirps)
	jsonResponse(w, 200, chirps)
}
