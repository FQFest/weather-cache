package weathercache

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/FQFest/weathercache/firestore"
	"github.com/FQFest/weathercache/memstore"
	"github.com/rs/cors"
)

type (
	App struct {
		log           *log.Logger
		store         store
		weatherClient fetcher
	}

	fetcher interface {
		Fetch(ctx context.Context) (io.ReadCloser, error)
	}

	store interface {
		UpdateWeather(ctx context.Context, data string) error
		GetCurWeather(ctx context.Context, zipCode string) (string, error)
	}
)

type Option func(*App)

func WithStore(s store) Option {
	return func(a *App) {
		a.store = s
	}
}

func WithWeatherClient(w fetcher) Option {
	return func(a *App) {
		a.weatherClient = w
	}
}

// New creates a new App instance.
func New(opts ...Option) *App {
	app := &App{}
	for _, opt := range opts {
		opt(app)
	}

	if app.log == nil {
		app.log = log.Default()
	}

	return app
}

func NewServer(app *App) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.ServeHTTP)
	// TODO: Only Allow necessary origins
	// https://ionicframework.com/docs/troubleshooting/cors#capacitor
	return cors.AllowAll().Handler(mux)
}

// PreFetch triggers the initial fetch of weather data.
// If mockData is not nil, it will be used to update the weather data, bypassing the actual fetch.
func (a App) PreFetch(mockData io.Reader) error {
	var data []byte

	if mockData != nil {
		var err error
		data, err = io.ReadAll(mockData)
		if err != nil {
			return fmt.Errorf("mockData readAll: %w", err)
		}
	} else {
		// Trigger weather update
		curWeatherRdr, err := a.weatherClient.Fetch(context.Background())
		if err != nil {
			return fmt.Errorf("weatherClient.Fetch: %w", err)
		}
		defer curWeatherRdr.Close()

		data, err = io.ReadAll(curWeatherRdr)
		if err != nil {
			return fmt.Errorf("weather readAll: %w", err)
		}
	}

	if err := a.store.UpdateWeather(context.Background(), string(data)); err != nil {
		return fmt.Errorf("store.UpdateWeather: %w", err)
	}
	return nil
}

// ServeHTTP is the entry point to the HTTP-triggered Cloud Function.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Trigger weather update
		a.handleUpdateWeather()(w, r)
		return
	case http.MethodGet:
		a.handleGetWeather()(w, r)
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

		data, err := io.ReadAll(curWeatherRdr)
		if err != nil {
			a.log.Printf("weather readAll: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := a.store.UpdateWeather(r.Context(), string(data)); err != nil {
			a.log.Printf("updateWeather: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		a.log.Println("Weather updated.")
	}
}

func (a *App) handleGetWeather() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Hard code to French Quarter zip
		curJson, err := a.store.GetCurWeather(r.Context(), "70117")
		if err != nil {
			if err == firestore.ErrNotFound || err == memstore.ErrNotFound {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			a.log.Printf("getCurWeather: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		dataR := strings.NewReader(curJson)
		_, err = dataR.WriteTo(w)
		if err != nil {
			a.log.Printf("response writeTo: %s", err.Error())
		}
	}
}
