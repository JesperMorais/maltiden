package main

import (
	"log"
	"net/http"
	"os"

	"maltiden/internal/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := api.NewRouter()

	log.Printf("Måltiden startar på :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
