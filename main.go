package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var mux http.ServeMux
var corsMux http.Handler
var server http.Server
var config apiConfig

const (
	port string = ":8080"
)

type apiConfig struct {
	fileServerHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) resetCounter() {
	cfg.fileServerHits = 0
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func init() {
	config = apiConfig{}
	mux = *http.NewServeMux()
	corsMux = middlewareCors(&mux)
	server = http.Server{}
	server.Addr = port
	server.Handler = corsMux
}

func main() {
	mux.Handle("/app/*", config.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// mux.HandleFunc("GET /api/metrics", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Write([]byte(fmt.Sprintf("Hits: %v\n", config.fileServerHits)))
	// })

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`
<html>
<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>
</html>`, config.fileServerHits)))
	})

	mux.HandleFunc("/api/reset", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		config.resetCounter()
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Validating chirp")
		type requestBody struct {
			Body string `json:"body"`
		}
		type responsePayload struct {
			Error string `json:"error,omitempty"`
			Valid bool   `json:"valid,omitempty"`
		}
		decoder := json.NewDecoder(r.Body)
		req := requestBody{}
		err := decoder.Decode(&req)
		if err != nil {
			log.Printf("Error decoding request: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("%s", req.Body)
		res := responsePayload{}
		if len(req.Body) > 140 {
			w.WriteHeader(400)
			res.Error = "Chirp is too long"
		} else {
			w.WriteHeader(http.StatusOK)
			res.Valid = true
		}
		data, err := json.Marshal(res)
		if err != nil {
			log.Printf("Error encoding request: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(data)
	})

	fmt.Printf("Server listening at host http://localhost%v\n", port)
	server.ListenAndServe()
}
