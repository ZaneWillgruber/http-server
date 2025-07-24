package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("app/"))
	apiCfg := apiConfig{}
	handler := http.StripPrefix("/app/", fs)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", readinessEndpoint)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsEndpoint)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetricsEndpoint)

	server := http.Server{Handler: mux, Addr: ":8080"}

	fmt.Println("Listening locally at: http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		println(err.Error)
	}
}

func readinessEndpoint(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsEndpoint(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html")
	writer.WriteHeader(200)
	writer.Write([]byte(fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
		`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) resetMetricsEndpoint(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(200)
	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Printf("%s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
