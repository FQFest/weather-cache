package main

import (
	"context"
	"log"
	"net/http"
	"os"

	wc "github.com/FQFest/weathercache"
	"github.com/FQFest/weathercache/firestore"
	"github.com/FQFest/weathercache/memstore"
	"github.com/FQFest/weathercache/weather"
)

func main() {
	store := memstore.New()
	storeOption := wc.WithStore(store)

	useMemStore := os.Getenv("USE_MEM_STORE") == "true"
	if !useMemStore {
		store, err := firestore.New(context.Background(), os.Getenv("GCP_PROJECT_ID"))
		if err != nil {
			log.Fatalf("firestore.New: %s", err.Error())
		}
		storeOption = wc.WithStore(store)
	}

	wClient := weather.New()
	app := wc.New(
		storeOption,
		wc.WithWeatherClient(wClient),
	)

	if useMemStore {
		if err := app.PreFetch(); err != nil {
			log.Fatalf("app.PreFetch: %s", err.Error())
		}
	}

	port := "9876"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("server starting on port: %s...", port)
	if err := http.ListenAndServe(":"+port, wc.NewServer(app)); err != nil {
		log.Fatalf("http server: %s", err.Error())
	}
}
