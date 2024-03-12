# Weather Cache

![Coverage](https://img.shields.io/badge/Coverage-30.4%25-yellow)

An application that caches weather data from the OpenWeatherMap API to ease upstream load.

## Environment Variables

To run this project, you will need to add the following environment variables to your `.env` file

| Name                   | Description                                                                                                                |
| ---------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `OPEN_WEATHER_API_KEY` | [OpenWeather API](https://openweathermap.org/api) Key                                                                      |
| `USE_MEM_STORE`        | If `true`, use an in-memory store. If `false`, the store is backed by [GCP Firestore](https://cloud.google.com/firestore). |

## Run Locally

Clone the project

```bash
  git clone git@github.com:FQFest/weather-cache.git
```

Go to the project directory

```bash
  cd weather-cache
```

Install dependencies

```bash
  go mod tidy
```

Start the server

```bash
  go run ./cmd/...
```

## Deployment

The application is deployed to [App Engine](https://cloud.google.com/appengine). The deployment process is automated using GitHub Actions with [`deploy.yml` workflow](.github/workflows/deploy.yml).

## License

[MIT](https://choosealicense.com/licenses/mit/)
