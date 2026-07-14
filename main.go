package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/tonyserranodev/chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	queries        *database.Queries
	platform       string
	jwtSecret      string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error connecting to db, %v", err)
	}

	dbQueries := database.New(db)

	cfg := apiConfig{
		queries:   dbQueries,
		platform:  os.Getenv("PLATFORM"),
		jwtSecret: jwtSecret,
	}

	const port = "8080"

	mux := http.NewServeMux()

	fs := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	// frontend
	mux.Handle("/app/", cfg.middleWareMetricsInc(fs))
	mux.Handle("assets/logo.png", fs)

	// api
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/refresh", cfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handlerRevoke)

	mux.HandleFunc("POST /api/users", cfg.handlerCreateUser)
	mux.HandleFunc("PUT /api/users", cfg.handlerUpdateUser)
	mux.HandleFunc("POST /api/login", cfg.handlerLogin)

	mux.HandleFunc("POST /api/chirps", cfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps/", cfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerGetChirpByID)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.handlerDeleteChirpByID)

	// admin
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
