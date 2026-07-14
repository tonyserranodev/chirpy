package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/tonyserranodev/chirpy/internal/auth"
	"github.com/tonyserranodev/chirpy/internal/database"
)

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "invalid input")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 400, "bad request")
		return
	}

	userParams := database.CreateUserParams{
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.queries.CreateUser(r.Context(), userParams)
	if err != nil {
		respondWithError(w, 400, "error creating user")
		return
	}

	response := UserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, 201, response)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "invalid input")
		return
	}

	user, err := cfg.queries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 401, "incorrect email or password")
		return
	}

	var match bool
	if match, err = auth.CheckPasswordHash(params.Password, user.HashedPassword); err != nil || !match {
		respondWithError(w, 401, "incorrect email or password")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(3600)*time.Second)
	if err != nil {
		respondWithError(w, 400, "error creating token")
		return
	}

	refreshToken := auth.MakeRefreshToken()

	createRefreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Duration(24*60) * time.Hour),
		RevokedAt: sql.NullTime{},
	}

	err = cfg.queries.CreateRefreshToken(r.Context(), createRefreshTokenParams)
	if err != nil {
		respondWithError(w, 400, "error creating refresh token")
		return
	}

	response := UserResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, 200, response)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	dbRefreshToken, err := cfg.queries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, 401, "error getting refresh token")
		return
	}

	if time.Now().After(dbRefreshToken.ExpiresAt) || dbRefreshToken.RevokedAt.Valid {
		respondWithError(w, 401, "refresh token expired or revoked")
		return
	}

	type refreshResponse struct {
		Token string `json:"token"`
	}

	accessToken, err := auth.MakeJWT(dbRefreshToken.UserID, cfg.jwtSecret, time.Duration(3600)*time.Hour)
	if err != nil {
		respondWithError(w, 401, "error creating access token")
		return
	}

	response := refreshResponse{
		Token: accessToken,
	}

	respondWithJSON(w, 200, response)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	dbRefreshToken, err := cfg.queries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, 401, "error getting refresh token")
		return
	}

	refreshTokenParams := database.UpdateRefreshTokenParams{
		Token:     dbRefreshToken.Token,
		UpdatedAt: time.Now(),
		RevokedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}

	err = cfg.queries.UpdateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		respondWithError(w, 401, "error updating refresh token")
		return
	}

	respondWithJSON(w, 204, nil)
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "bad request")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, "unauthorized request")
		return
	}

	credParams := database.UpdateUserEmailAndPasswordParams{
		UpdatedAt:      time.Now(),
		Email:          params.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	}

	user, err := cfg.queries.UpdateUserEmailAndPassword(r.Context(), credParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	res := response{
		ID:        userID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusOK, res)
}
