package main

import (
	"fmt"
  "os"
  "flag"
	"log"
	"net/http"
  "github.com/kekekekyle/database"
  "github.com/joho/godotenv"
)

type apiConfig struct {
  fileserverHits int
  database *database.DB
  jwtSecret string
  polkaApiKey string
}

func (cfg *apiConfig) middlewareMetricsInc (next http.Handler) http.Handler {
  nextHandler := func (w http.ResponseWriter, req *http.Request) {
    cfg.fileserverHits++
    next.ServeHTTP(w, req)
  }
  return http.HandlerFunc(nextHandler)
}

func (cfg *apiConfig) resetMetricsHandler (w http.ResponseWriter, r *http.Request) {
  cfg.fileserverHits = 0
  w.Write([]byte("Metrics reset."))
}

func (cfg *apiConfig) getMetricsHandler (w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/html")
  w.Write([]byte(fmt.Sprintf(`<html>
<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>
</html>`, cfg.fileserverHits)))
}

func handleHealth (w http.ResponseWriter, r *http.Request) {
  r.Header.Set("Content-Type", "text/plain; charset=utf-8")
  w.WriteHeader(200)
  w.Write([]byte("OK"))
}

// deleteDB deletes the database file
func deleteDB(path string) error {
  if err := os.Remove(path); err != nil {
    return err
  }
  return nil
}

func main() {
  const filepathRoot = "app"
  const databasePath = "database.json"
	const port = "42069"

  godotenv.Load()
  jwtSecret := os.Getenv("JWT_SECRET")
  polkaApiKey := os.Getenv("POLKA_API_KEY")

  debug := flag.Bool("debug", false, "Enable debug mode")
  flag.Parse()
  if *debug == true {
    deleteDB(databasePath)
  }

  database, err := database.NewDB(databasePath)
  if err != nil {
    fmt.Println("Unable to create database.")
  }

  apiCfg := &apiConfig {
    fileserverHits: 0,
    database: database,
    jwtSecret: jwtSecret,
    polkaApiKey: polkaApiKey,
  }

	mux := http.NewServeMux()
  handleFiles := http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))
  mux.Handle("/app/", apiCfg.middlewareMetricsInc(handleFiles))
  mux.HandleFunc("GET /api/healthz", handleHealth)
  mux.HandleFunc("GET /admin/metrics", apiCfg.getMetricsHandler)
  mux.HandleFunc("/api/reset", apiCfg.resetMetricsHandler)
  mux.Handle("POST /api/chirps", apiCfg.authenticate(&HandleCreateChirps{api: apiCfg}))
  mux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirps)
  mux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.handleGetChirpById)
  mux.Handle("DELETE /api/chirps/{chirpId}", apiCfg.authenticate(&HandleDeleteChirps{api: apiCfg}))
  mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
  mux.HandleFunc("PUT /api/users", apiCfg.handleUpdateUser)
  mux.HandleFunc("POST /api/login", apiCfg.handleLogin)
  mux.HandleFunc("POST /api/refresh", apiCfg.handleRefreshToken)
  mux.HandleFunc("POST /api/revoke", apiCfg.handleRevokeToken)
  mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlePolkaWebhooks)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
