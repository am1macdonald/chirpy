package payloads

import (
	"encoding/json"
	"log"
	"net/http"
)

type ChripPostBody struct {
	Body string `json:"body"`
}

type ResponsePayload struct {
	Body        string `json:"body,omitempty"`
	CleanedBody string `json:"cleaned_body,omitempty"`
}

type UsersPostBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	Email string `json:"email"`
	ID    int    `json:"id"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type UpdateRequest struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func DecodeRequest[T any](r *http.Request, dest *T) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&dest)
	if err != nil {
		log.Printf("Error decoding request: %s", err)
		return err
	}
	return nil
}

func EncodeResponse(res ResponsePayload) ([]byte, error) {
	return json.Marshal(res)
}
