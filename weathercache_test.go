package weathercache_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	wc "github.com/FQFest/weathercache"
	"github.com/FQFest/weathercache/memstore"
	"github.com/FQFest/weathercache/weather"
	"github.com/stretchr/testify/require"
)

type mockFetcher struct {
	curWeather weather.Current
}

func (m *mockFetcher) Fetch(ctx context.Context) (io.ReadCloser, error) {
	cw, err := json.Marshal(m.curWeather)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(cw)), nil
}

func TestServer(t *testing.T) {
	t.Run("has weather data on startup", func(t *testing.T) {
		client := &mockFetcher{
			curWeather: weather.Current{Main: weather.Main{Temp: 72.5}},
		}

		app := wc.New(wc.WithStore(memstore.New()), wc.WithWeatherClient(client))
		srv := wc.NewServer(app)
		err := app.PreFetch()
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		srv.ServeHTTP(res, req)
		require.Equalf(t, http.StatusOK, res.Code, "unexpected HTTP error: %s", res.Body.String())
		require.Equal(t, "application/json", res.Header().Get("Content-Type"))

		var got weather.Current
		err = json.NewDecoder(res.Body).Decode(&got)
		require.NoError(t, err)
		require.NotEmpty(t, got)
		require.Equal(t, 72.5, got.Main.Temp)
	})

	t.Run("polls weather API for new data", func(t *testing.T) {
		client := &mockFetcher{
			curWeather: weather.Current{Main: weather.Main{Temp: 72.5}},
		}

		app := wc.New(wc.WithStore(memstore.New()), wc.WithWeatherClient(client))
		app.PreFetch()
		app.StartPoll(time.Millisecond * 100)

		srv := wc.NewServer(app)

		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		srv.ServeHTTP(res, req)
		require.Equalf(t, http.StatusOK, res.Code, "unexpected HTTP error: %s", res.Body.String())
		require.Equal(t, "application/json", res.Header().Get("Content-Type"))

		var got weather.Current
		err := json.NewDecoder(res.Body).Decode(&got)
		require.NoError(t, err)
		require.Equal(t, 72.5, got.Main.Temp)

		client.curWeather.Main.Temp = 86.5
		time.Sleep(time.Millisecond * 200)

		res = httptest.NewRecorder()
		srv.ServeHTTP(res, req)
		require.Equalf(t, http.StatusOK, res.Code, "unexpected HTTP error: %s", res.Body.String())
		require.Equal(t, "application/json", res.Header().Get("Content-Type"))

		err = json.NewDecoder(res.Body).Decode(&got)
		require.NoError(t, err)
		require.Equal(t, 86.5, got.Main.Temp)
	})
}
