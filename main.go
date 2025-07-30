package main

import (
	"database/sql"
	"fmt"
	"github/zanewillgruber/http-server/internal/database"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func main() {
	godotenv.Load()

	platform := os.Getenv("PLATFORM")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Can't connect to the database")
		return
	} else {
		fmt.Println("Connected to database")
	}

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("app/"))
	apiCfg := apiConfig{dbQueries: database.New(db), platform: platform}
	handler := http.StripPrefix("/app/", fs)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", readinessEndpoint)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsEndpoint)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetricsEndpoint)

	//chirps
	mux.HandleFunc("POST /api/chirps", apiCfg.addChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpsByID)

	//users
	mux.HandleFunc("POST /api/users", apiCfg.addUser)

	server := http.Server{Handler: mux, Addr: ":8080"}

	fmt.Println("Listening locally at: http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		println(err.Error)
	}
}
