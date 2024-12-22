package db

import (
	"fmt"
	"os"
	"sync"

	"github.com/A7bari/RunWave/internal/types"
)

// InMemStore is an in-memory implementation of the Store interface
// this is used for testing purposes
// you can replace this with a database implementation
type InMemStore struct {
	Results map[string]types.TaskOutput
}

var (
	store     *InMemStore
	storeOnce sync.Once
)

func GetInMemStore() *InMemStore {
	storeOnce.Do(func() {
		store = &InMemStore{Results: make(map[string]types.TaskOutput)}
	})
	return store
}

// implement Store interface
func (s *InMemStore) SaveResult(value types.TaskOutput) error {
	s.Results[value.TaskID] = value
	go saveToFile(value)
	return nil
}

// implement Store interface
func (s *InMemStore) GetResult(id string) (types.TaskOutput, error) {
	result, ok := s.Results[id]
	if !ok {
		return types.TaskOutput{}, nil
	}
	return result, nil
}

// implement Store interface
func (s *InMemStore) DeleteResult(id string) error {
	delete(s.Results, id)
	return nil
}

// implement Store interface
func (s *InMemStore) Close() error {
	return nil
}

func saveToFile(value types.TaskOutput) error {
	filePath, err := os.Getwd()
	if err != nil {
		return err
	}
	filePath += "/results.txt" // Specify the file path here

	// Open the file in append mode, create it if it doesn't exist
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Convert the value to a string
	output := fmt.Sprintf("TaskID: %s Status: %s Output: %s ", value.TaskID, value.Status, value.Output)

	// Write the value to the file
	_, err = file.WriteString(output)
	if err != nil {
		return err
	}

	return nil
}
