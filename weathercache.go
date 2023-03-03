package weathercache

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/FQFest/weathercache/firestore"
	"github.com/FQFest/weathercache/weather"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

type (
	App struct {
		log           log.Logger
		store         store
		weatherClient fetcher
	}

	fetcher interface {
		Fetch(ctx context.Context) (io.ReadCloser, error)
	}

	store interface {
		UpdateWeather(ctx context.Context, data string) error
	}
)

func init() {
	// Register an HTTP function with the Functions Framework
	// This handler name maps to the entry point name in the Google Cloud Function platform.
	// https://cloud.google.com/functions/docs/writing/write-http-functions
	functions.HTTP("EntryPoint", func(w http.ResponseWriter, r *http.Request) {
		New().ServeHTTP(w, r)
	})
}

// New creates a new App instance.
func New() *App {
	wClient := weather.New()
	store, err := firestore.New(context.Background(), os.Getenv("GCP_PROJECT_ID"))
	if err != nil {
		log.Fatalf("firestore new: %s", err.Error())
	}

	return &App{
		log:           *log.Default(),
		store:         store,
		weatherClient: wClient,
	}
}

// ServeHTTP is the entry point to the HTTP-triggered Cloud Function.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Trigger weather update
		a.handleUpdateWeather()(w, r)
		return

	default:
		http.Error(w, fmt.Sprintf("method %q not allowed", r.Method), http.StatusMethodNotAllowed)
	}
}

func (a *App) handleUpdateWeather() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		curWeatherRdr, err := a.weatherClient.Fetch(r.Context())
		if err != nil {
			a.log.Printf("error: %s", err.Error())
			http.Error(w, "could not fetch weather data", http.StatusInternalServerError)
			return
		}
		defer curWeatherRdr.Close()

		// TODO: Write current weather to Firestore
		// https://firebase.google.com/docs/firestore/manage-data/add-data#go
		data, err := io.ReadAll(curWeatherRdr)
		if err != nil {
			a.log.Printf("weather readAll: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = a.store.UpdateWeather(
			r.Context(),
			// TODO: Should we just write JSON directly, or unmarshal/marshal and store the weather struct?
			// That seems superflous since we're just going to reserialize the data back to the client.
			// Unlessss..we use the realtime clients. TBC..
			string(data),
		)
		if err != nil {
			a.log.Printf("updateWeather: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// TODO: Chunk for debugging delete later
		var curWeather weather.Current
		err = json.Unmarshal(data, &curWeather)
		if err != nil {
			a.log.Panicf("decode: %s", err)
			return
		}
		fmt.Printf("Current Weather:\n%+v", curWeather)
		// TODO ^^^ Chunk for debugging delete later
	}
}
