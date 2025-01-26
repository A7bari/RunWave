package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"sync"

	"github.com/A7bari/RunWave/internal/store"
	"github.com/A7bari/RunWave/internal/taskqueue"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type PostgresStore struct {
	DB *gorm.DB
}

var (
	psStore     *PostgresStore
	psStoreOnce sync.Once
)

var _ store.Store = (*PostgresStore)(nil)

type Task struct {
	TaskID    string       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Language  string       `gorm:"type:varchar(100);not null"`
	Code      string       `gorm:"type:text;not null"`
	Status    string       `gorm:"type:varchar(50);default:pending"`
	MaxRetry  int          `gorm:"default:3"`
	RetryCnt  int          `gorm:"default:0"`
	CreatedAt time.Time    `gorm:"autoCreateTime"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime"`
	Results   []TaskResult `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`
}

type TaskResult struct {
	TaskResultID string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TaskID       string    `gorm:"type:uuid;not null;index"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	Output       string    `gorm:"type:text"`
	IsError      bool      `gorm:"default:false"`
}

func GetPostgresStore() *PostgresStore {
	psStoreOnce.Do(func() {
		dsn := "user=youruser password=yourpassword dbname=yourdatabase host=localhost port=5433 sslmode=disable"
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
		})
		if err != nil {
			log.Fatalf("Error connecting to the database: %v", err)
		}

		if err := db.AutoMigrate(&Task{}, &TaskResult{}); err != nil {
			log.Fatalf("Error auto-migrating database schema: %v", err)
		}

		psStore = &PostgresStore{DB: db}
	})
	return psStore
}

func (s *PostgresStore) SaveResult(ctx context.Context, taskID string, taskResult taskqueue.TaskResult) error {
	result := TaskResult{
		TaskID:  taskID,
		Output:  taskResult.Output,
		IsError: taskResult.IsError,
	}

	if err := s.DB.WithContext(ctx).Create(&result).Error; err != nil {
		return fmt.Errorf("failed to save task result: %w", err)
	}
	return nil
}

func (s *PostgresStore) CreateTask(ctx context.Context, task taskqueue.Task) (string, error) {
	newTask := Task{
		Language: task.Language,
		Code:     task.Code,
		Status:   task.Status,
		MaxRetry: task.MaxRetry,
		RetryCnt: task.RetryCnt,
	}

	if err := s.DB.WithContext(ctx).Create(&newTask).Error; err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}
	return newTask.TaskID, nil
}

func (s *PostgresStore) UpdateTask(ctx context.Context, task taskqueue.Task) error {
	updates := map[string]interface{}{
		"status":     task.Status,
		"max_retry":  task.MaxRetry,
		"retry_cnt":  task.RetryCnt,
		"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}

	if err := s.DB.WithContext(ctx).
		Model(&Task{}).
		Where("task_id = ?", task.TaskID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetTaskById(ctx context.Context, taskID string) (taskqueue.Task, error) {
	var task Task
	if err := s.DB.WithContext(ctx).Where("task_id = ?", taskID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return taskqueue.Task{}, nil
		}
		return taskqueue.Task{}, fmt.Errorf("failed to retrieve task: %w", err)
	}

	s.DB.WithContext(ctx).
		Model(&TaskResult{}).
		Where("task_id = ?", taskID).
		Order("created_at DESC").
		First(&task.Results)

	tbuilder := taskqueue.NewTaskBuilder().
		SetID(task.TaskID).
		SetLang(task.Language).
		SetCode(task.Code).
		SetStatus(task.Status).
		SetMaxRetry(task.MaxRetry).
		SetRetryCnt(task.RetryCnt)

	if len(task.Results) > 0 {
		tbuilder.SetResult(taskqueue.TaskResult{
			Output:  task.Results[0].Output,
			IsError: task.Results[0].IsError,
		})
	}

	return tbuilder.Build(), nil
}

func (s *PostgresStore) ListTasks(ctx context.Context, filter store.TaskFilter) ([]taskqueue.Task, error) {
	query := s.DB.WithContext(ctx).Model(&Task{})

	if filter.Lang != nil {
		query = query.Where("language = ?", *filter.Lang)
	}
	if filter.Completed != nil {
		status := "completed"
		if !*filter.Completed {
			status = "pending"
		}
		query = query.Where("status = ?", status)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("created_at <= ?", *filter.CreatedTo)
	}

	var tasks []Task
	if err := query.Limit(filter.Limit).Offset(filter.Offset).Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	taskList := make([]taskqueue.Task, len(tasks))
	for i, task := range tasks {
		taskList[i] = taskqueue.NewTaskBuilder().
			SetID(task.TaskID).
			SetLang(task.Language).
			SetCode(task.Code).
			SetStatus(task.Status).
			SetMaxRetry(task.MaxRetry).
			SetRetryCnt(task.RetryCnt).
			Build()
	}
	return taskList, nil
}

func (s *PostgresStore) UpdateTaskFields(ctx context.Context, taskID string, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	fields["updated_at"] = gorm.Expr("CURRENT_TIMESTAMP")

	if err := s.DB.WithContext(ctx).
		Model(&Task{}).
		Where("task_id = ?", taskID).
		Updates(fields).Error; err != nil {
		return fmt.Errorf("failed to update task fields: %w", err)
	}
	return nil
}

func (s *PostgresStore) Close() error {
	sqlDB, err := s.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
