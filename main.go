package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
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

	// admin
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)

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
