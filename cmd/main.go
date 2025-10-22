package main

import (
	"log"
	"net/http"

	"grafana-plugin-api/internal/api"
	"grafana-plugin-api/internal/config"

	"github.com/gin-gonic/gin"
)

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

	// Setup router
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Routes
	router.POST("/query_logs", handler.QueryLogs)

	// Start server
	addr := cfg.Server.GetAddress()
	log.Printf("Starting Grafana Plugin API server on http://%s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
