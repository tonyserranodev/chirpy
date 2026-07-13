package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	_ "github.com/lib/pq"
)

func main() {
	cfg := apiConfig{}

	const port = "8080"

	mux := http.NewServeMux()

	fs := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	// frontend
	mux.Handle("/app/", cfg.middleWareMetricsInc(fs))
	mux.Handle("assets/logo.png", fs)

	// api
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	// admin
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("GET /admin/reset", cfg.handlerReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port %s\n", port)
	log.Fatal(srv.ListenAndServe())
}

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middleWareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	msg := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileServerHits.Load())

	w.Write([]byte(msg))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	cfg.fileServerHits.Store(0)
	msg := fmt.Sprintf("Reset successfully! Hits: %v", cfg.fileServerHits.Load())
	w.Write([]byte(msg))
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "invalid input")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "chirp too long")
		return
	}

	cleanedBody := cleanProfanity(params.Body)

	type successResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}
	success := successResponse{
		CleanedBody: cleanedBody,
	}
	respondWithJSON(w, 200, success)
}

func cleanProfanity(text string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		lowercaseWord := strings.ToLower(word)
		if lowercaseWord == "kerfuffle" || lowercaseWord == "sharbert" || lowercaseWord == "fornax" {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	respBody := errorResponse{
		Error: msg,
	}

	respondWithJSON(w, code, respBody)
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	dat, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}
