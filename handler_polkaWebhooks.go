package main

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/am1macdonald/chirpy/internal/payloads"
)

type polkaWebhookBody struct {
	Event string `json:"event"`
	Data  struct {
		UserID int `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) HandlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	ts := r.Header.Get("Authorization")
	if ts == "" {
		errorResponse(w, 401, errors.New("api token required"))
		return
	}
	auth := strings.Split(ts, " ")
	if auth[0] != "ApiKey" || auth[1] != config.polkaKey {
		errorResponse(w, 401, errors.New("api token required"))
		return
	}
	req := polkaWebhookBody{}
	err := payloads.DecodeRequest(r, &req)
	if err != nil {
		errorResponse(w, 500, err)
		return
	}
	log.Println(req.Event)
	if req.Event != "user.upgraded" {
		jsonResponse(w, 200, "success")
		return
	}
	user, err := db.GetUser(req.Data.UserID)
	if err != nil {
		errorResponse(w, 404, err)
		return
	}
	user.IsChirpyRed = true
	user, err = db.UpdateUser(user)
	if err != nil {
		errorResponse(w, 500, err)
		return
	}
	jsonResponse(w, 200, struct{}{})
}
