// Package weather supplies functionality to communicate with OpenWeather API.
package weather

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

type (
	client struct {
		apiBase string
		apiKey  string
	}

	Coordinates struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}

	Weather struct {
		ID          int    `json:"id"`          // 802
		Main        string `json:"main"`        // "Clouds"
		Description string `json:"description"` // "scattered clouds"
		Icon        string `json:"icon"`        // "03d"
	}

	Main struct {
		Temp      float64 `json:"temp"`       // 78.8
		FeelsLike float64 `json:"feels_like"` // 78.8
		TempMin   float64 `json:"temp_min"`   // 77
		TempMax   float64 `json:"temp_max"`   // 81
		Pressure  int     `json:"pressure"`   // 1005
		Humidity  float64 `json:"humidity"`   // 75
	}

	Wind struct {
		Speed float64 `json:"speed"` // 20.71
		Deg   int     `json:"deg"`   // 170
		Gust  float64 `json:"gust"`  // 28.7
	}

	Clouds struct {
		All int `json:"all"`
	}

	Sys struct {
		Type    int    `json:"type"`
		ID      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	}

	Current struct {
		Coord      Coordinates `json:"coord"`
		Weather    []Weather   `json:"weather"`
		Base       string      `json:"base"`
		Main       Main        `json:"main"`
		Visibility int         `json:"visibility"`
		Wind       Wind        `json:"wind"`
		Clouds     Clouds      `json:"clouds"`
		Dt         int         `json:"dt"`
		Sys        Sys         `json:"sys"`
		Timezone   int         `json:"timezone"`
		ID         int         `json:"id"`
		Name       string      `json:"name"`
		Cod        int         `json:"cod"`
	}
)

func New() *client {
	apiKey := os.Getenv("OPEN_WEATHER_API_KEY")
	return &client{
		apiBase: "https://api.openweathermap.org/data/2.5",
		apiKey:  apiKey,
	}
}

// Fetch fetches the current weather from the OpenWeather API.
func (c *client) Fetch(ctx context.Context) (io.ReadCloser, error) {

	endpoint := c.apiBase + "/weather"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return io.NopCloser(bytes.NewReader([]byte{})), fmt.Errorf("NewRequestWithContext %w", err)
	}

	q := req.URL.Query()
	q.Add("zip", "70116")
	q.Add("appid", c.apiKey)
	q.Add("units", "imperial")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp.Body, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp.Body, fmt.Errorf("body ReadAll: %w", err)
		}

		return resp.Body, fmt.Errorf("could not get current weather\nStatusCode: %d\nBody: %s", resp.StatusCode, string(body))
	}

	if err != nil {
		return resp.Body, err
	}

	return resp.Body, nil
}
