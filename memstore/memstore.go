// Package memstore provides a very simple in-memory implementation of the store interface.
package memstore

import (
	"context"
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

type store struct {
	mu   sync.RWMutex
	data map[string]string
}

func New() *store {
	return &store{
		data: make(map[string]string),
	}
}

func (s *store) set(key, value string) {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
}

func (s *store) get(key string) (string, bool) {
	s.mu.RLock()
	value, ok := s.data[key]
	s.mu.RUnlock()
	return value, ok
}

func (s *store) GetCurWeather(ctx context.Context, zipCode string) (string, error) {
	v, ok := s.get(zipCode)
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (s *store) UpdateWeather(ctx context.Context, data string) error {
	// Like the Firestore implementation, I'm hard coding the document ID to the French Quarter Zip Code. We can get this from the request if necessary, but keeping it simple for now
	docID := "70117"
	s.set(docID, data)
	return nil
}
