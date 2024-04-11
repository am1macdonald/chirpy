package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
)

func (cfg *apiConfig) HandleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	ts, err := getTokenString(r)
	if err != nil {
		errorResponse(w, 500, err)
		return
	}

	token, err := parseTokenString(ts)
	if err != nil {
		errorResponse(w, 500, err)
		return
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil || issuer != "chirpy-access" {
		log.Println("Invalid token")
		errorResponse(w, 401, errors.New("invalid token"))
		return
	}

	u, err := token.Claims.GetSubject()
	if err != nil {
		errorResponse(w, 401, err)
		return
	}

	userId, err := strconv.Atoi(u)
	id, err := strconv.Atoi(r.PathValue("chirp_id"))
	if err != nil {
		jsonResponse(w, 500, err.Error())
		return
	}

	chirp, err := db.GetChirp(id)
	if err != nil {
		jsonResponse(w, 404, err.Error())
		return
	}

	if userId != chirp.AuthorID {
		jsonResponse(w, 403, "unauthorized")
		return
	}

	ok := db.DeleteChirp(id)
	if !ok {
		jsonResponse(w, 500, errors.New("delete failed"))
		return
	}
}
