package weathercache

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/FQFest/weathercache/weather"
)

type (
	App struct {
		weatherClient fetcher
		log           log.Logger
	}

	fetcher interface {
		Fetch(ctx context.Context) (weather.Current, error)
	}
)

// New creates a new App instance.
func New() *App {
	wClient := weather.New()

	return &App{
		weatherClient: wClient,
		log:           *log.Default(),
	}
}

// ServeHTTP is the entry point to the HTTP-triggered Cloud Function.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Trigger weather update
		curWeather, err := a.weatherClient.Fetch(r.Context())
		if err != nil {
			a.log.Printf("error: %s", err.Error())
			http.Error(w, "could not fetch weather data", http.StatusInternalServerError)
			return
		}

		// TODO: Write current weather to Firestore
		// https://firebase.google.com/docs/firestore/manage-data/add-data#go
		fmt.Printf("Current Weather:\n%+v", curWeather)
		return

	default:
		http.Error(w, fmt.Sprintf("method %q not allowed", r.Method), http.StatusMethodNotAllowed)
	}
}
