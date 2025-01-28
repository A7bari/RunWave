package app

import (
	"fmt"
	"net/http"

	"github.com/A7bari/RunWave/internal/services"
	"github.com/A7bari/RunWave/internal/taskqueue"
	"github.com/A7bari/RunWave/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterRoutes(router *gin.Engine) {

	router.POST("/submit", func(c *gin.Context) {
		var req types.CodeExecutionReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// check if the language is supported
		if !IsSupportedLanguage(req.Language) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported language"})
			return
		}

		task := taskqueue.NewTask(uuid.New().String(), req.Code, req.Language)

		err := HandleTask(task, c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"task_id": task.TaskID})
	})

	router.GET("/result/:task_id", func(c *gin.Context) {
		taskID := c.Param("task_id")

		taskService, ok := c.Get("taskService")
		if !ok {
			panic("Store not found in context")
		}

		task, err := taskService.(*services.TaskService).GetTask(taskID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"task_id": task.TaskID,
			"status":  task.Status,
			"output":  task.Result.Output,
		})
	})

}

func HandleTask(task *taskqueue.Task, c *gin.Context) error {

	// save the task in the database
	taskService, ok := c.Get("taskService")
	if !ok {
		return fmt.Errorf("task service not found in context")
	}

	_, err := taskService.(*services.TaskService).CreateTask(*task)
	if err != nil {
		return err
	}

	// add the task to the task queue
	taskQueue, exist := c.Get("taskPublisher")
	if !exist {
		panic("task queue not found in context")
	}

	tmsg, err := task.Encode()
	if err != nil {
		return err
	}

	err = taskQueue.(taskqueue.QueuePublisher).PublishTask(taskqueue.QueueMsg{
		Body: tmsg,
	}, task.Language)

	if err != nil {
		fmt.Print("Error publishing task: %v", err)
		return err
	}

	return nil
}
