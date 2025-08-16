package in_memory_repository

import "fmt"

// Stores uuid in memory and retrieves it if already present
type InMemoryStore struct {
	data map[string]bool
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]bool),
	}
}

// Add stores a new uuid in memory. Returns an error if the uuid is already present
func (s *InMemoryStore) Add(uuid string) error {
	if s.Exists(uuid) {
		return fmt.Errorf("UUID %s already exists", uuid)
	}
	s.data[uuid] = true
	return nil
}

func (s *InMemoryStore) Exists(uuid string) bool {
	_, exists := s.data[uuid]
	return exists
}
