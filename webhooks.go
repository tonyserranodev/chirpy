package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/tonyserranodev/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerUpgradeToChirpyRed(w http.ResponseWriter, r *http.Request) {
	apiKeyFromHeader, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized request")
		return
	}

	if apiKeyFromHeader != cfg.apiKey {
		respondWithError(w, http.StatusUnauthorized, "invalid api key")
		return
	}

	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid input")
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	err = cfg.queries.UpgradeUserToChirpyRedByID(r.Context(), params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "user not found")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
