package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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
		wc.WithMockData(mockData()),
	)

	rawSec := os.Getenv("POLL_INTERVAL_SECS")
	sec, err := strconv.Atoi(rawSec)
	if err != nil && rawSec != "" {
		log.Fatalf("POLL_INTERVAL_SECS must be an integer.\nGot: %s", rawSec)
	}

	if sec < 1 {
		// Default to 2 minutes.
		app.StartPoll(time.Minute * 2)
	} else {
		app.StartPoll(time.Duration(sec) * time.Second)
	}

	if useMemStore {
		if err := app.PreFetch(); err != nil {
			log.Fatalf("app.PreFetch: %s", err.Error())
		}
	}

	port := "9876"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           wc.NewServer(app),
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}
	srvErrs := make(chan error, 1)
	go func() {
		log.Printf("server starting on port: %s...", port)
		srvErrs <- srv.ListenAndServe()
	}()

	// Wait for errors from the server or for a signal to shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	shutdown := gracefulShutdown(app, srv)

	select {
	case err := <-srvErrs:
		shutdown(err)
	case sig := <-quit:
		shutdown(sig)
	}
	log.Println("server shutdown")
}

func gracefulShutdown(app *wc.App, srv *http.Server) func(reason interface{}) {
	return func(reason interface{}) {
		log.Printf("shutting down server: %v", reason)
		shutdownTimeout := 3 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		app.StopPoll()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("srv.Shutdown: %s", err.Error())
		}
	}
}

// mockData returns a mock weather data for testing if USE_MOCK_DATA is true.
func mockData() []byte {
	if os.Getenv("USE_MOCK_DATA") != "true" {
		return nil
	}

	fmt.Println("using mock data")
	weather := weather.Current{
		Main: weather.Main{
			Temp:      504.0,
			FeelsLike: 78.8,
			TempMin:   77,
			TempMax:   550,
		},
		Base: "stations",
	}

	data, err := json.Marshal(weather)
	if err != nil {
		panic(err)
	}
	return data
}
