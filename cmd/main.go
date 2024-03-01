package main

import (
	"context"
	"log"
	"net/http"
	"os"

	wc "github.com/FQFest/weathercache"
	"github.com/FQFest/weathercache/firestore"
	"github.com/FQFest/weathercache/weather"
)

func main() {
	wClient := weather.New()
	store, err := firestore.New(context.Background(), os.Getenv("GCP_PROJECT_ID"))
	if err != nil {
		log.Fatalf("firestore.New: %s", err.Error())
	}

	app := wc.New(
		wc.WithStore(store),
		wc.WithWeatherClient(wClient),
	)
	srv := wc.NewServer(app)

	port := "9876"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("server starting on port: %s...", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		log.Fatalf("http server: %s", err.Error())
	}
}
