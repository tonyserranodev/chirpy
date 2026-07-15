package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/tonyserranodev/chirpy/internal/auth"
	"github.com/tonyserranodev/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "bad request")
		return
	}

	validatedBody := validateChirp(w, params.Body)

	chirpParams := database.CreateChirpParams{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      validatedBody,
		UserID:    userID,
	}

	dbChirp, err := cfg.queries.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		log.Printf("error creating chirp: %v\n", err)
		respondWithError(w, 400, "error creating chirp")
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(w, 201, chirp)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	sortVal := r.URL.Query().Get("sort")
	if sortVal != "asc" && sortVal != "desc" {
		respondWithError(w, 400, "bad request")
		return
	}

	var dbChirps []database.Chirp
	if authorID != "" {
		parsedID, err := uuid.Parse(authorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "internal server error parsing author id")
			return
		}

		dbChirps, err = cfg.queries.GetChirpsByAuthorID(r.Context(), parsedID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "chirps by author id not found")
			return
		}

		chirps := copyChirps(dbChirps)

		sortChirps(sortVal, chirps)

		respondWithJSON(w, 200, chirps)
		return
	}

	dbChirps, err := cfg.queries.GetChirps(r.Context())
	if err != nil {
		log.Printf("error getting chirps: %v\n", err)
		respondWithError(w, 400, "error getting chirps")
		return
	}

	chirps := copyChirps(dbChirps)
	sortChirps(sortVal, chirps)

	respondWithJSON(w, 200, chirps)
}

func copyChirps(dbChirps []database.Chirp) []Chirp {
	chirps := make([]Chirp, 0, len(dbChirps))
	for _, dbChirp := range dbChirps {
		chirp := Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		}

		chirps = append(chirps, chirp)
	}

	return chirps
}

func sortChirps(sortVal string, chirps []Chirp) {
	if sortVal == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
		return
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
	})
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("error parsing id: %v\n", err)
		respondWithError(w, 400, "invalid id")
		return
	}

	dbChirp, err := cfg.queries.GetChirpByID(r.Context(), parsedID)
	if err != nil {
		log.Printf("error getting chirp: %v\n", err)
		respondWithError(w, 404, "page not found")
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(w, 200, chirp)
}

func validateChirp(w http.ResponseWriter, text string) string {
	// validate chirp body length
	if len(text) > 140 {
		respondWithError(w, 400, "chirp too long")
		return ""
	}

	// clean profanity
	words := strings.Split(text, " ")
	for i, word := range words {
		lowercaseWord := strings.ToLower(word)
		if lowercaseWord == "kerfuffle" || lowercaseWord == "sharbert" || lowercaseWord == "fornax" {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

func (cfg *apiConfig) handlerDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized request")
		return
	}

	id := r.PathValue("chirpID")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("error parsing id: %v\n", err)
		respondWithError(w, 400, "invalid id")
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized request")
		return
	}

	chirpToDelete, err := cfg.queries.GetChirpByID(r.Context(), parsedID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "chirp not found")
		return
	}

	if chirpToDelete.UserID != userID {
		respondWithError(w, http.StatusForbidden, "unauthorized request")
		return
	}

	err = cfg.queries.DeleteChirpByID(r.Context(), chirpToDelete.ID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "chirp not found")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
