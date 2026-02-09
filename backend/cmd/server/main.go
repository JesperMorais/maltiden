package main

import (
	"log"
	"net/http"
	"os"

	"maltiden/internal/api"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/utils"
)

func main() {
	// Validate JWT secret before any other setup
	jwtSecret := os.Getenv("JWT_SECRET")
	if err := utils.InitJWTSecret(jwtSecret); err != nil {
		log.Fatal("JWT_SECRET must be at least 32 characters")
	}

	// Get config from env variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./data/maltiden.db"
	}

	// Open db (runs migrations)
	db, err := sqlite.Open(dbPath)
	if err != nil {
		log.Fatal("Database error:", err)
	}
	defer db.Close()

	// Create router (injects db)
	router := api.NewRouter(db)

	log.Printf("Måltiden startar på :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
