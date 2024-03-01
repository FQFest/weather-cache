package weathercache_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
		app.PreFetch()

		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		srv.ServeHTTP(res, req)
		require.Equalf(t, http.StatusOK, res.Code, "unexpected HTTP error: %s", res.Body.String())
		require.Equal(t, "application/json", res.Header().Get("Content-Type"))

		var got weather.Current
		err := json.NewDecoder(res.Body).Decode(&got)
		require.NoError(t, err)
		require.NotEmpty(t, got)
		require.Equal(t, 72.5, got.Main.Temp)
	})
}
