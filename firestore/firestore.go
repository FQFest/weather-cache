package firestore

import (
	"context"
	"errors"
	"fmt"
	"os"

	gcpfirestore "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	Store struct {
		client   *gcpfirestore.Client
		collName string
		dataKey  string
	}

	jsonDoc map[string]interface{}
)

var ErrNotFound = errors.New("not found")

func New(ctx context.Context, projectID string) (*Store, error) {
	var app *firebase.App
	var err error

	if inGCP() {
		// Use default credentials when running in GCP context
		conf := &firebase.Config{ProjectID: projectID}
		app, err = firebase.NewApp(ctx, conf)
		if err != nil {
			return &Store{}, err
		}
	} else {
		// Use a service account when running locally
		sa := option.WithCredentialsFile("../service-account.json")
		app, err = firebase.NewApp(ctx, nil, sa)
		if err != nil {
			return &Store{}, err
		}
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return &Store{}, err
	}

	return &Store{
		client:   client,
		collName: "weather",
		dataKey:  "json",
	}, nil
}

func (s *Store) UpdateWeather(ctx context.Context, dataJSON string) error {
	// Hard coding the document ID to the French Quarter Zip Code. We can get this from the request if necessary, but keeping it simple for now
	docID := "70117"
	// Firestore Documents are maps
	// https://firebase.google.com/docs/firestore/manage-data/add-data#data_types
	doc := jsonDoc{s.dataKey: dataJSON}

	_, err := s.client.Collection(s.collName).Doc(docID).Set(ctx, doc)
	if err != nil {
		return fmt.Errorf("firestore set: %w", err)
	}
	return nil
}

// inGCP returns true if the Function is running in the GCP environment, otherwise false.
func inGCP() bool {
	// Assume K_REVISION is only set in the GCP context
	// https://cloud.google.com/functions/docs/configuring/env-var
	return len(os.Getenv("K_REVISION")) > 0
}

// GetCurWeather retrieves the current weather data from the store as a JSON string.
func (s *Store) GetCurWeather(ctx context.Context, zipCode string) (string, error) {
	docSnap, err := s.client.Collection(s.collName).Doc(zipCode).Get(ctx)
	if err != nil && status.Code(err) != codes.NotFound {
		return "", err
	}
	if !docSnap.Exists() {
		return "", ErrNotFound
	}
	doc := docSnap.Data()
	curWeather, ok := doc[s.dataKey].(string)
	if !ok {
		return "", fmt.Errorf("invalid type for document value. Expected JSON string")
	}
	return curWeather, nil
}
