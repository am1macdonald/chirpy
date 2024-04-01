package payloads

import (
	"encoding/json"
	"log"
	"net/http"
)

type RequestBody struct {
	Body string `json:"body"`
}
type ResponsePayload struct {
	Body        string `json:"body,omitempty"`
	CleanedBody string `json:"cleaned_body,omitempty"`
}

func DecodeRequest(r *http.Request) (*RequestBody, error) {
	decoder := json.NewDecoder(r.Body)
	req := RequestBody{}
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding request: %s", err)
		return nil, err
	}
	return &req, nil
}

func EncodeResponse(res ResponsePayload) ([]byte, error) {
	return json.Marshal(res)
}
