package app

import (
	"fmt"
	"net/http"
	"sync"

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

		task := taskqueue.NewTask(
			uuid.New().String(),
			req.Language,
			req.Code,
			taskqueue.TaskCallbacksOpts{})

		taskQueue, exist := c.Get(req.Language)
		if !exist {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported language"})
			return
		}

		taskQueue.(taskqueue.TaskQueue).AddTask(task)

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

	router.POST("/submit/event", func(c *gin.Context) {
		// Set headers for streaming
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		var req types.CodeExecutionReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Flush the response writer
		flusher := c.Writer
		wg := sync.WaitGroup{}
		wg.Add(1)

		task := taskqueue.NewTask(
			uuid.New().String(),
			req.Language,
			req.Code,
			taskqueue.TaskCallbacksOpts{
				OnCreated: func(task taskqueue.Task) {
					fmt.Fprintf(flusher, "data: NEW TASK task: %s\n\n", task.GetTaskID())
					flusher.Flush()
				},
				OnChanged: func(task taskqueue.Task) {
					fmt.Fprintf(flusher, "data: status: %s task: %s\n\n", task.GetStatus(), task.GetTaskID())
					flusher.Flush()

					if task.GetError() != nil {
						fmt.Fprintf(flusher, "data: ERROR: %v task: %s\n\n", task.GetError(), task.GetTaskID())
						flusher.Flush()
					}
				},

				OnResult: func(task taskqueue.Task) {
					output, isError := task.GetResult()
					fmt.Fprintf(flusher, "data: #### FINISHED ####  task: %s [is error: %v] output: %s  \n\n", task.GetTaskID(), isError, output)
					flusher.Flush()

					wg.Done()
				},
			})

		taskQueue, exist := c.Get(req.Language)
		if !exist {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported language"})
			return
		}

		taskQueue.(taskqueue.TaskQueue).AddTask(task)

		wg.Wait()
	})

}
