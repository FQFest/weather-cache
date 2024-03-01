package memstore_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/FQFest/weathercache/memstore"
	"github.com/FQFest/weathercache/weather"
	"github.com/stretchr/testify/require"
)

func TestWeatherStore(t *testing.T) {
	testCases := []struct {
		name    string
		zipCode string
		data    weather.Current
	}{
		{
			name:    "French Quarter",
			zipCode: "70117",
			data: weather.Current{
				Coord:   weather.Coordinates{Lat: 29.9686, Lon: -90.0646},
				Weather: []weather.Weather{{ID: 802, Main: "Clouds", Description: "scattered clouds", Icon: "03d"}},
				Base:    "stations",
				Main:    weather.Main{Temp: 78.8, FeelsLike: 78.8, TempMin: 77, TempMax: 81, Pressure: 1005, Humidity: 75},
				Wind:    weather.Wind{Speed: 20.71, Deg: 170, Gust: 28.7},
				Clouds:  weather.Clouds{All: 40},
				Sys:     weather.Sys{Type: 1, ID: 3920, Country: "US", Sunrise: 1619827680, Sunset: 1619876460},
				Name:    "New Orleans",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store := memstore.New()
			serialized, err := json.Marshal(tc.data)
			if err != nil {
				require.NoError(t, err)
			}

			err = store.UpdateWeather(ctx, string(serialized))
			require.NoError(t, err)

			data, err := store.GetCurWeather(ctx, tc.zipCode)
			require.NoError(t, err)
			var got weather.Current

			err = json.Unmarshal([]byte(data), &got)
			require.NoError(t, err)
			require.Equal(t, tc.data, got)
		})
	}
}
