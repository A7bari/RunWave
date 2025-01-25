package db

import (
	"database/sql"
	"testing"

	"github.com/A7bari/RunWave/internal/taskqueue"
	"github.com/A7bari/RunWave/internal/types"
	"github.com/stretchr/testify/assert"

	_ "github.com/lib/pq" // PostgreSQL driver
)

const testConnStr = "user=youruser password=yourpassword dbname=testdatabase host=localhost port=5433 sslmode=disable"

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", testConnStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("Test database unreachable: %v", err)
	}

	if err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	return db
}

func TestPostgresStore_SaveResult(t *testing.T) {
	db := setupTestDB(t)
	store := &PostgresStore{DB: db}
	defer store.Close()

	result := types.TaskOutput{
		TaskID:  "1234",
		Version: 1,
		Output:  "Test Output",
		IsError: false,
	}

	err := store.SaveResult(result)
	assert.NoError(t, err, "Saving task output should not return an error")
}

func TestPostgresStore_GetResult(t *testing.T) {
	db := setupTestDB(t)
	store := &PostgresStore{DB: db}
	defer store.Close()

	result := types.TaskOutput{
		TaskID:  "1234",
		Version: 1,
		Output:  "Test Output",
		IsError: false,
	}
	_ = store.SaveResult(result)

	fetchedResult, err := store.GetResult("1234")
	assert.NoError(t, err, "Fetching task output should not return an error")
	assert.Equal(t, result.TaskID, fetchedResult.TaskID, "TaskID should match")
	assert.Equal(t, result.Output, fetchedResult.Output, "Output should match")
}

func TestPostgresStore_CreateTask(t *testing.T) {
	db := setupTestDB(t)
	store := &PostgresStore{DB: db}
	defer store.Close()

	task := taskqueue.NewTaskBuilder().
		SetLang("Go").
		SetCode("fmt.Println(\"Hello, World!\")").
		SetStatus("pending").
		SetMaxRetry(3).
		SetRetryCnt(0).
		Build()

	taskID, err := store.CreateTask(task)
	assert.NoError(t, err, "Creating task should not return an error")
	assert.NotEmpty(t, taskID, "TaskID should not be empty")
}

func TestPostgresStore_UpdateTask(t *testing.T) {
	db := setupTestDB(t)
	store := &PostgresStore{DB: db}
	defer store.Close()

	task := taskqueue.NewTaskBuilder().
		SetLang("Go").
		SetCode("fmt.Println(\"Hello, World!\")").
		SetStatus("pending").
		SetMaxRetry(3).
		SetRetryCnt(0).
		Build()

	taskID, _ := store.CreateTask(task)
	task.(taskID).SetStatus("running")

	err := store.UpdateTask(task)
	assert.NoError(t, err, "Updating task should not return an error")

	fetchedTask, err := store.GetTask(taskID)
	assert.NoError(t, err, "Fetching task should not return an error")
	assert.Equal(t, "running", fetchedTask.GetStatus(), "Status should be updated")
}

func TestPostgresStore_DeleteResult(t *testing.T) {
	db := setupTestDB(t)
	store := &PostgresStore{DB: db}
	defer store.Close()

	result := types.TaskOutput{
		TaskID:  "1234",
		Version: 1,
		Output:  "Test Output",
		IsError: false,
	}
	_ = store.SaveResult(result)

	err := store.DeleteResult(1234, 1)
	assert.NoError(t, err, "Deleting task result should not return an error")

	fetchedResult, err := store.GetResult("1234")
	assert.NoError(t, err, "Fetching task output after deletion should not return an error")
	assert.Empty(t, fetchedResult.TaskID, "Fetched result should be empty after deletion")
}
