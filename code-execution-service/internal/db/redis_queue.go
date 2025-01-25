package db

import (
	"context"
	"sync"

	"github.com/A7bari/RunWave/internal/taskqueue"
	"github.com/redis/go-redis/v9"
)

var (
	redisTaskQueue     *RedisTaskQueue
	redisTaskQueueOnce sync.Once
)

// GetRedisTaskQueue initializes the RedisTaskQueue with a Redis client
// only initializes the RedisTaskQueue once
func GetRedisTaskQueue(redisAddr, queueName string) *RedisTaskQueue {
	redisTaskQueueOnce.Do(func() {
		client := redis.NewClient(&redis.Options{
			Addr: redisAddr, // e.g., "localhost:6379"
		})
		redisTaskQueue = &RedisTaskQueue{
			client:    client,
			queueName: queueName,
		}
	})
	return redisTaskQueue
}

type RedisTaskQueue struct {
	client    *redis.Client
	queueName string
}

// AddTask adds a task to the Redis list
func (r *RedisTaskQueue) AddTask(task taskqueue.Task) error {
	// Serialize the task to a string or JSON
	taskId := task.GetTaskID()

	return r.client.RPush(context.Background(), r.queueName, taskId).Err()
}

// GetTask retrieves and removes a task from the Redis list
func (r *RedisTaskQueue) GetTask() (string, error) {
	ctx := context.Background()
	taskStr, err := r.client.LPop(ctx, r.queueName).Result()
	if err != nil {
		return "", err // Handle empty queue or other errors
	}
	return taskStr, nil
}
