package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github/zanewillgruber/http-server/internal/database"
	"net/http"

	"github.com/google/uuid"
)

func readinessEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
		`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) resetMetricsEndpoint(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "nice try buddy")
		return
	}

	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)

	err := cfg.dbQueries.ResetUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, err.Error())
	}
}

func (cfg *apiConfig) addUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string
	}

	params := parameters{}
	err := recieveJSON(w, r, &params)
	if err != nil {
		respondWithError(w, 500, "Could not parse json: "+err.Error())
		return
	}

	if params.Email == "" {
		respondWithError(w, 400, "missing email field")
		return
	}

	dbUser, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 500, "Could not create user: "+err.Error())
		return
	}

	user := User{dbUser.ID, dbUser.CreatedAt, dbUser.UpdatedAt, dbUser.Email}

	respondWithJSON(w, 201, user)

}

func (cfg *apiConfig) addChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body    string
		User_Id string
	}

	params := parameters{}
	err := recieveJSON(w, r, &params)
	if err != nil {
		return
	}

	err = uuid.Validate(params.User_Id)
	if err != nil {
		respondWithError(w, 400, "not a valid user id")
		return
	}

	userId := uuid.MustParse(params.User_Id)

	err = validateChirp(w, params.Body)
	if err != nil {
		return
	}

	censored := censorProfaneWords(params.Body)

	dbChirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: censored, UserID: userId})
	if err != nil {
		respondWithError(w, 500, "Could not create chirp: "+err.Error())
		return
	}

	chirp := Chirp{dbChirp.ID, dbChirp.CreatedAt, dbChirp.UpdatedAt, dbChirp.Body, dbChirp.UserID}

	respondWithJSON(w, 201, chirp)
}

func validateChirp(w http.ResponseWriter, body string) error {

	if body == "" {
		respondWithError(w, 400, "no body")
		return fmt.Errorf("chirp body can't be empty")
	}

	if len(body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return fmt.Errorf("chirp is too long")
	}

	return nil
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "could not retrive chirps")
		return
	}

	chirps := mapSlice(dbChirps, func(dbChirp database.Chirp) Chirp {
		return Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			User_Id:   dbChirp.UserID,
		}
	})

	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) getChirpsByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("chirpID")

	err := uuid.Validate(idParam)
	if err != nil {
		respondWithError(w, 400, "not a valid uuid")
		return
	}

	id := uuid.MustParse(idParam)

	dbChirp, err := cfg.dbQueries.GetChirpByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(404)
			return
		}

		respondWithError(w, 500, "failed to get chirp: "+err.Error())
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		User_Id:   dbChirp.UserID,
	}

	respondWithJSON(w, 200, chirp)
}
