package app

import (
	"fmt"
	"net/http"

	"github.com/A7bari/RunWave/internal/store"
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

		task := taskqueue.NewTaskBuilder().
			SetID(uuid.New().String()).
			SetLang(req.Language).
			SetCode(req.Code).
			Build()

		err := EnqueueTask(task, c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"task_id": task.GetTaskID()})
	})

	router.GET("/result/:task_id", func(c *gin.Context) {
		taskID := c.Param("task_id")

		db, ok := c.Get("store")
		if !ok {
			panic("Store not found in context")
		}

		task, err := db.(store.Store).GetResult(taskID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"task_id": task.TaskID, "status": task.Status, "output": task.Output})
	})

	// router.POST("/submit/event", func(c *gin.Context) {
	// 	// Set headers for streaming
	// 	c.Header("Content-Type", "text/event-stream")
	// 	c.Header("Cache-Control", "no-cache")
	// 	c.Header("Connection", "keep-alive")

	// 	var req types.CodeExecutionReq
	// 	if err := c.ShouldBindJSON(&req); err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 		return
	// 	}

	// 	// Flush the response writer
	// 	flusher := c.Writer
	// 	wg := sync.WaitGroup{}
	// 	wg.Add(1)

	// 	task := taskqueue.NewTaskBuilder().
	// 		SetID(uuid.New().String()).
	// 		SetLang(req.Language).
	// 		SetCode(req.Code).
	// 		Build()

	// 	err := EnqueueTask(task, c)

	// 	if err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 		return
	// 	}

	// 	wg.Wait()
	// })

}

func EnqueueTask(task taskqueue.Task, c *gin.Context) error {
	taskQueue, exist := c.Get(task.GetLanguage())
	if !exist {
		return fmt.Errorf("unsupported language")
	}

	taskQueue.(taskqueue.TaskQueue).AddTask(task)
	return nil
}
