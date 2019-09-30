package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Color struct {
	Color   string
	Version string
}

func color(w http.ResponseWriter, r *http.Request) {
	color := os.Getenv("APP_COLOR")
	version := os.Getenv("APP_VERSION")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Color{
		Color:   color,
		Version: version,
	})
}

func main() {
	http.HandleFunc("/getcolor", color)
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "9691"
	}
	log.Printf("Listening on port: %s", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
