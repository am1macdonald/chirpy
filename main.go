package main

import (
	"encoding/json"
	"fmt"
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

func jsonResponse(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(data)
}

func errorResponse(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	w.Write([]byte(msg))
	return
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

	// mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
	// 	req, err := payloads.DecodeRequest(r)
	// 	if err != nil {
	// 		errorResponse(w, 500, "failed to decode the request")
	// 		return
	// 	}
	// 	res := payloads.ResponsePayload{}
	// 	s, err := validate.Validate(req.Body)
	// 	if err != nil {
	// 		res.Body = s
	// 		jsonResponse(w, 400, res)
	// 		return
	// 	} else {
	// 		res.CleanedBody = s
	// 	}
	// 	jsonResponse(w, 200, res)
	// })
	//

	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {

	})

	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {
	})

	fmt.Printf("Server listening at host http://localhost%v\n", port)
	server.ListenAndServe()
}
