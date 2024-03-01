package weathercache_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	wc "github.com/FQFest/weathercache"
	"github.com/FQFest/weathercache/memstore"
	"github.com/FQFest/weathercache/weather"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	t.Run("has weather data on startup", func(t *testing.T) {
		app := wc.New(wc.WithStore(memstore.New()))
		srv := wc.NewServer(app)

		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		srv.ServeHTTP(res, req)
		require.Equalf(t, http.StatusOK, res.Code, "unexpected HTTP error: %s", res.Body.String())
		require.Equal(t, "application/json", res.Header().Get("Content-Type"))

		var got weather.Current
		err := json.NewDecoder(res.Body).Decode(&got)
		require.NoError(t, err)
		require.NotEmpty(t, got)
	})
}
