package firestore

import (
	"context"
	"fmt"
	"os"

	gcpfirestore "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

type (
	Store struct {
		client *gcpfirestore.Client
	}

	jsonDoc map[string]interface{}
)

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
		sa := option.WithCredentialsFile("../serviceAccount.json")
		app, err = firebase.NewApp(ctx, nil, sa)
		if err != nil {
			return &Store{}, err
		}
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return &Store{}, err
	}

	return &Store{client: client}, nil
}

func (s *Store) UpdateWeather(ctx context.Context, dataJSON string) error {
	// Hard coding the document ID to the French Quarter Zip Code. We can get this from the request if necessary, but keeping it simple for now
	docID := "70117"
	// Firestore Documents are maps
	// https://firebase.google.com/docs/firestore/manage-data/add-data#data_types
	doc := jsonDoc{"json": dataJSON}

	_, err := s.client.Collection("weather").Doc(docID).Set(ctx, doc)
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
