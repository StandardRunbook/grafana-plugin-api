package main

import (
	"log"
	"net/http"

	"grafana-plugin-api/internal/api"
	"grafana-plugin-api/internal/config"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Server config - host: %s, port: %d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("ClickHouse config - url: %s, user: %s, database: %s",
		cfg.ClickHouse.URL, cfg.ClickHouse.User, cfg.ClickHouse.Database)

	// Create API handler
	handler := api.NewHandler(cfg)

	// Verify tables exist (only warn if it fails, don't exit)
	if err := handler.VerifyTables(); err != nil {
		log.Printf("Warning: Failed to verify ClickHouse tables: %v", err)
		log.Printf("Server will continue and return mock data when ClickHouse is unavailable")
	}

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/query_logs", handler.QueryLogs)

	// Wrap with CORS middleware
	server := &http.Server{
		Addr:    cfg.Server.GetAddress(),
		Handler: corsMiddleware(mux),
	}

	// Start server
	log.Printf("Starting Grafana Plugin API server on http://%s", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
