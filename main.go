package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

var badWords = [...]string{"kerfuffle", "sharbert", "fornax"}

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("app/"))
	apiCfg := apiConfig{}
	handler := http.StripPrefix("/app/", fs)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", readinessEndpoint)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsEndpoint)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetricsEndpoint)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

	server := http.Server{Handler: mux, Addr: ":8080"}

	fmt.Println("Listening locally at: http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		println(err.Error)
	}
}

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
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Printf("%s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func resopondWithError(w http.ResponseWriter, code int, message string) {
	type errorVals struct {
		Error string `json:"error"`
	}

	respBody := errorVals{
		Error: message,
	}

	data, err := json.Marshal(respBody)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %s\n", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %s\n", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string
	}

	type response struct {
		Body string `json:"body"`
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Error decoding parameters: %s\n", err)
		resopondWithError(w, 500, "something went wrong")
		return
	}

	if params.Body == "" {
		resopondWithError(w, 400, "no body")
		return
	}

	if len(params.Body) > 140 {
		resopondWithError(w, 400, "Chirp is too long")
		return
	}

	censored := censorProfaneWords(params.Body)

	respondWithJSON(w, 200, response{Body: censored})
}

func censorProfaneWords(body string) string {
	split := strings.Split(body, " ")

	for i, word := range split {
		for _, badWord := range badWords {
			if strings.EqualFold(word, badWord) {
				split[i] = "****"
			}
		}
	}

	return strings.Join(split, " ")
}
