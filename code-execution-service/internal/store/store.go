package store

import "github.com/A7bari/RunWave/internal/types"

type Store interface {
	SaveResult(value types.TaskOutput) error
	GetResult(id string) (types.TaskOutput, error)
	DeleteResult(id string) error
	Close() error
}
