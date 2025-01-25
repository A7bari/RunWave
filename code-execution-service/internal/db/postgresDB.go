package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/A7bari/RunWave/internal/store"
	"github.com/A7bari/RunWave/internal/taskqueue"
	"github.com/A7bari/RunWave/internal/types"
	_ "github.com/lib/pq" // PostgreSQL driver
)

type PostgresStore struct {
	DB *sql.DB
}

var (
	PSstore     *PostgresStore
	PSstoreOnce sync.Once
)

// runtime check PostgresStore implements Store interface
var _ store.Store = (*PostgresStore)(nil)

// GetPostgresStore initializes and returns a singleton PostgresStore instance.
func GetPostgresStore() *PostgresStore {
	PSstoreOnce.Do(func() {
		// Connect to the PostgreSQL container using the service name as hostname
		connStr := "user=youruser password=yourpassword dbname=yourdatabase host=localhost port=5433 sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatal("Error connecting to the database: ", err)
		}
		PSstore = &PostgresStore{DB: db}

		// Initialize the database tables if they don't exist
		if err := InitializeDB(db); err != nil {
			log.Fatal("Error initializing the database: ", err)
		}
	})
	return PSstore
}

// SaveResult stores the TaskOutput in the PostgreSQL database
func (s *PostgresStore) SaveResult(value types.TaskOutput) error {
	query := `INSERT INTO task_results (task_id, version, output, is_error) 
              VALUES ($1, $2, $3, $4, $5)`
	_, err := s.DB.Exec(query, value.TaskID, value.Version, value.Output, value.IsError)
	if err != nil {
		return fmt.Errorf("failed to insert task result: %v", err)
	}
	return nil
}

// GetResult retrieves the TaskOutput by TaskID and Version from the PostgreSQL database
func (s *PostgresStore) GetResult(taskID string) (types.TaskOutput, error) {
	var result types.TaskOutput
	query := `
			SELECT task_id, version, output, is_error, error_message
			FROM task_results
			WHERE task_id = $1 AND version = (
					SELECT MAX(version)
					FROM task_results
					WHERE task_id = $1
			)
	`
	row := s.DB.QueryRow(query, taskID)

	err := row.Scan(&result.TaskID, &result.Version, &result.Output, &result.IsError)
	if err != nil {
		if err == sql.ErrNoRows {
			return types.TaskOutput{}, nil
		}
		return types.TaskOutput{}, fmt.Errorf("failed to get task result: %v", err)
	}
	return result, nil
}

// DeleteResult deletes a TaskOutput from the PostgreSQL database
func (s *PostgresStore) DeleteResult(taskID int, version int) error {
	query := `DELETE FROM task_results WHERE task_id = $1 AND version = $2`
	_, err := s.DB.Exec(query, taskID, version)
	if err != nil {
		return fmt.Errorf("failed to delete task result: %v", err)
	}
	return nil
}

// CreateTask inserts a new task into the PostgreSQL database
func (s *PostgresStore) CreateTask(task taskqueue.Task) (string, error) {
	query := `INSERT INTO tasks (language, code, status, max_retry, retry_cnt) 
              VALUES ($1, $2, $3, $4, $5) RETURNING task_id`
	var taskID string
	err := s.DB.QueryRow(query, task.GetLanguage(), task.GetCode(), task.GetStatus(), task.GetMaxRetry(), task.GetRetryCnt()).Scan(&taskID)
	if err != nil {
		return "", fmt.Errorf("failed to create task: %v", err)
	}
	return taskID, nil
}

// UpdateTask updates an existing task in the PostgreSQL database
func (s *PostgresStore) UpdateTask(task taskqueue.Task) error {
	query := `UPDATE tasks SET status = $3, max_retry = $4, retry_cnt = $5, 
              updated_at = CURRENT_TIMESTAMP WHERE task_id = $6`
	_, err := s.DB.Exec(query, task.GetStatus(), task.GetMaxRetry(), task.GetRetryCnt(), task.GetTaskID())
	if err != nil {
		return fmt.Errorf("failed to update task: %v", err)
	}
	return nil
}

// GetTask retrieves a Task by TaskID from the PostgreSQL database
func (s *PostgresStore) GetTask(taskID string) (taskqueue.Task, error) {
	var (
		taskId   string
		language string
		code     string
		status   string
		maxRetry int
		retryCnt int
	)

	query := `SELECT task_id, language, code, status, max_retry, retry_cnt 
              FROM tasks WHERE task_id = $1`
	row := s.DB.QueryRow(query, taskID)

	err := row.Scan(&taskId, &language, &code, &status, &maxRetry, &retryCnt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task: %v", err)
	}

	task := taskqueue.NewTaskBuilder().
		SetID(taskId).
		SetLang(language).
		SetCode(code).
		SetStatus(status).
		SetMaxRetry(maxRetry).
		SetRetryCnt(retryCnt).
		Build()

	return task, nil
}

// Close closes the PostgreSQL database connection
func (s *PostgresStore) Close() error {
	return s.DB.Close()
}

// Initialize the PostgreSQL table if it doesn't exist
func InitializeDB(db *sql.DB) error {

	query := `
	CREATE TABLE tasks (
    task_id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    language VARCHAR(100) NOT NULL,
    code TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    max_retry INT DEFAULT 3, 
    retry_cnt INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
    CONSTRAINT status_check CHECK (status IN ('pending', 'running', 'completed', 'failed')) 
	);
	
	-- Create TaskResult Table with task_id as the primary key
	CREATE TABLE task_results (
		task_result_id  UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    task_id UUID NOT NULL, 
    version INT NOT NULL, 
    output TEXT,
    is_error BOOLEAN DEFAULT FALSE, 
    error_message TEXT, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
    FOREIGN KEY (task_id) REFERENCES tasks(task_id) ON DELETE CASCADE, 
    CONSTRAINT unique_task_version UNIQUE(task_id, version) 
	);`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return nil
}
