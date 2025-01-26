package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/A7bari/RunWave/internal/store"
	"github.com/A7bari/RunWave/internal/taskqueue"
)

// Setup test DB
func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "user=youruser password=yourpassword dbname=yourdatabase host=localhost port=5433 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&Task{}, &TaskResult{})
	require.NoError(t, err)

	// Clear tables before each test
	err = db.Exec("TRUNCATE task, task_result RESTART IDENTITY CASCADE").Error
	require.NoError(t, err)

	return db
}

func TestCreateTask(t *testing.T) {
	database := setupTestDB(t)
	store := &PostgresStore{DB: database}

	task := taskqueue.Task{
		Language: "Go",
		Code:     "fmt.Println(\"Hello, World!\")",
		Status:   "pending",
		MaxRetry: 3,
		RetryCnt: 0,
	}
	taskID, err := store.CreateTask(context.Background(), task)
	assert.NoError(t, err)
	assert.NotEmpty(t, taskID)

	// Verify task in DB
	var dbTask Task
	err = database.Where("task_id = ?", taskID).First(&dbTask).Error
	assert.NoError(t, err)
	assert.Equal(t, task.Language, dbTask.Language)
}

func TestSaveAndGetResult(t *testing.T) {
	database := setupTestDB(t)
	store := &PostgresStore{DB: database}

	// Create a task
	task := taskqueue.Task{
		Language: "Go",
		Code:     "fmt.Println(\"Hello, Test!\")",
		Status:   "pending",
		MaxRetry: 3,
		RetryCnt: 0,
	}
	taskID, err := store.CreateTask(context.Background(), task)
	require.NoError(t, err)

	task.Result = taskqueue.TaskResult{
		Output:  "Hello, Test!",
		IsError: false,
	}
	err = store.SaveResult(context.Background(), taskID, task.Result)
	assert.NoError(t, err)

	// Retrieve task and results
	dbTask, err := store.GetTaskById(nil, taskID) // Adjust to your `GetTaskById` method signature
	assert.NoError(t, err)
	assert.Equal(t, task.Result.Output, dbTask.Result.Output)
	assert.Equal(t, task.Result.IsError, dbTask.Result.IsError)
}

func TestUpdateTask(t *testing.T) {
	database := setupTestDB(t)
	store := &PostgresStore{DB: database}

	// Create a task
	task := taskqueue.Task{
		Language: "Go",
		Code:     "fmt.Println(\"Update Test\")",
		Status:   "pending",
		MaxRetry: 3,
		RetryCnt: 0,
	}
	taskID, err := store.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Update the task
	task.TaskID = taskID
	task.Status = "completed"
	task.RetryCnt = 1
	err = store.UpdateTask(context.Background(), task)
	assert.NoError(t, err)

	// Verify updated task
	updatedTask, err := store.GetTaskById(nil, taskID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", updatedTask.Status)
	assert.Equal(t, 1, updatedTask.RetryCnt)
}

func TestGetTask(t *testing.T) {
	database := setupTestDB(t)
	store := &PostgresStore{DB: database}

	// Create a task
	task := taskqueue.Task{
		Language: "Python",
		Code:     "print(\"Hello, Task Test\")",
		Status:   "pending",
		MaxRetry: 5,
		RetryCnt: 2,
	}
	taskID, err := store.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Retrieve the task
	retrievedTask, err := store.GetTaskById(context.Background(), taskID)
	assert.NoError(t, err)
	assert.Equal(t, task.Language, retrievedTask.Language)
	assert.Equal(t, task.Code, retrievedTask.Code)
	assert.Equal(t, task.MaxRetry, retrievedTask.MaxRetry)
	assert.Equal(t, task.RetryCnt, retrievedTask.RetryCnt)
}

func TestUpdateTaskFields(t *testing.T) {
	database := setupTestDB(t)
	store := &PostgresStore{DB: database}

	task := taskqueue.Task{
		Language: "Python",
		Code:     "print(\"Hello, Task Test\")",
		Status:   "pending",
		MaxRetry: 5,
		RetryCnt: 2,
	}
	taskID, err := store.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Update the task status
	store.UpdateTaskFields(context.Background(), taskID, map[string]interface{}{
		"status": "completed",
	})

	// Verify updated task
	updatedTask, err := store.GetTaskById(nil, taskID)
	assert.NoError(t, err)

	assert.Equal(t, "completed", updatedTask.Status)
}

func TestListTasks(t *testing.T) {
	database := setupTestDB(t)
	db := &PostgresStore{DB: database}

	// Create tasks
	task := taskqueue.Task{
		Language: "Python",
		Code:     "print(\"Hello, Task Test\")",
		Status:   "completed",
		MaxRetry: 5,
		RetryCnt: 2,
	}
	_, err := db.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Create another task
	task = taskqueue.Task{
		Language: "Python",
		Code:     "print(\"Hello, Task Test\")",
		Status:   "completed",
		MaxRetry: 5,
		RetryCnt: 2,
	}
	_, err = db.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// List tasks
	lang := "Python"
	completed := true
	tasks, err := db.ListTasks(context.Background(), store.TaskFilter{
		Lang:      &lang,
		Completed: &completed,
	})

	assert.NoError(t, err)
	for _, ts := range tasks {
		assert.Equal(t, "Python", ts.Language)
		assert.Equal(t, "completed", ts.Status)
	}
}
