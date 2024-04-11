package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/am1macdonald/chirpy/internal/chirps"
	"github.com/am1macdonald/chirpy/internal/database"
	"github.com/am1macdonald/chirpy/internal/payloads"
	"github.com/golang-jwt/jwt/v5"

	"github.com/joho/godotenv"
)

var mux http.ServeMux
var corsMux http.Handler
var server http.Server
var config apiConfig

const (
	port string = ":8080"
)

var db *database.DB

type apiConfig struct {
	fileServerHits int
	jwtSecret      string
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	config = apiConfig{}
	config.jwtSecret = os.Getenv("JWT_SECRET")
	mux = *http.NewServeMux()
	corsMux = middlewareCors(&mux)
	server = http.Server{}
	server.Addr = port
	server.Handler = corsMux
	dbp, err := database.NewDB()
	if err != nil {
		log.Fatalln("Failed to load database")
	}
	db = dbp
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

	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		req := payloads.ChripPostBody{}
		err := payloads.DecodeRequest(r, &req)
		if err != nil {
			errorResponse(w, 500, "failed to decode the request")
			return
		}
		s, err := chirps.Validate(req.Body)
		if err != nil {
			jsonResponse(w, 400, err.Error())
			return
		}
		chirp, err := db.CreateChirp(s)
		if err != nil {
			jsonResponse(w, 500, err.Error())
			return
		}
		jsonResponse(w, 201, chirp)
	})

	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		chirps, err := db.GetChirps()
		if err != nil {
			jsonResponse(w, 500, err.Error())
			return
		}
		jsonResponse(w, 200, chirps)
	})

	mux.HandleFunc("GET /api/chirps/{chirp_id}", func(w http.ResponseWriter, r *http.Request) {
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
		jsonResponse(w, 200, chirp)
	})

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		req := payloads.UsersPostBody{}
		err := payloads.DecodeRequest(r, &req)
		if err != nil {
			jsonResponse(w, 500, err.Error())
			return
		}
		user, err := db.CreateUser(req.Email, req.Password)
		if err != nil {
			jsonResponse(w, 500, err.Error())
			return
		}
		pl := payloads.CreateUserResponse{
			ID:    user.ID,
			Email: user.Email,
		}
		jsonResponse(w, 201, pl)
	})

	mux.HandleFunc("PUT /api/users", func(w http.ResponseWriter, r *http.Request) {
		req := payloads.UpdateRequest{}
		err := payloads.DecodeRequest(r, &req)
		if err != nil {
			jsonResponse(w, 500, err.Error())
			return
		}
		ts := r.Header.Get("Authorization")
		if ts == "" {
			jsonResponse(w, 401, "Authorization header is required")
		}
		ts = strings.Split(ts, " ")[1]
		log.Println(ts)
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(ts, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.jwtSecret), nil
		})
		if err != nil {
			log.Println(err.Error())
			jsonResponse(w, 401, "invalid token")
			return
		}
		userID, err := token.Claims.GetSubject()
		if err != nil {
			jsonResponse(w, 500, "no claims subject")
			return
		}
		idInt, err := strconv.Atoi(userID)
		if err != nil {
			jsonResponse(w, 500, "failed to parse userID")
			return
		}
		user, err := db.GetUser(idInt)
		if err != nil {
			log.Println("No user with id")
			jsonResponse(w, 404, "User with id does not exist")
			return
		}
		user.Email = req.Email
		err = user.UpdatePassword(req.Password)
		if err != nil {
			jsonResponse(w, 500, "failed to update password")
		}
		user, err = db.UpdateUser(idInt, user)
		if err != nil {
			jsonResponse(w, 500, "Could not update user")
		}
		pl := payloads.CreateUserResponse{
			ID:    user.ID,
			Email: user.Email,
		}
		jsonResponse(w, 200, pl)
	})

	mux.HandleFunc("GET /api/users/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("user_id"))
		if err != nil {
			jsonResponse(w, 500, err.Error())
			return
		}
		chirp, err := db.GetUser(id)
		if err != nil {
			jsonResponse(w, 404, err.Error())
			return
		}
		jsonResponse(w, 200, chirp)
	})

	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		req := payloads.LoginRequest{}
		err := payloads.DecodeRequest(r, &req)
		if err != nil {
			jsonResponse(w, 500, err.Error())
			return
		}
		user, err := db.GetUserByEmail(req.Email)
		if err != nil {
			jsonResponse(w, 400, err.Error())
			return
		}
		valid := user.Validate(req.Password)
		if !valid {
			jsonResponse(w, 401, "Invalid password")
			return
		}
		token, err := user.GetToken(config.jwtSecret, req.ExpiresInSeconds)
		if err != nil {
			log.Printf("%v", err)
			jsonResponse(w, 500, "Failed to generate token")
			return
		}
		pl := payloads.LoginResponse{
			ID:    user.ID,
			Email: user.Email,
			Token: token,
		}
		log.Println("Here")
		jsonResponse(w, 200, pl)
	})

	fmt.Printf("Server listening at host http://localhost%v\n", port)
	server.ListenAndServe()
}
