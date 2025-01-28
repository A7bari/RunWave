package coderunner

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"testing"

	"github.com/A7bari/RunWave/internal/db"
	"github.com/A7bari/RunWave/internal/services"
	"github.com/A7bari/RunWave/internal/taskqueue"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func connectToK8s() PodManager {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	// Build the config from the kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	return NewPodManager(clientset, config, "code-exec-system")
}

// TestSchedulerStartWorkers verifies the worker's task processing logic
func TestSchedulerStartWorkers(t *testing.T) {
	// Mock dependencies
	mockTaskQueue, err := db.GetRabbitMQConn()

	db := db.GetPostgresStore()

	taskService := services.NewTaskService(db)

	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ: %v", err)
	}

	taskIds := []string{}

	for i := 1; i <= 4; i++ {
		taskID := uuid.New().String()
		tmsg := taskqueue.NewTask(taskID, "python", "print('Hello, World!')")

		// create a task
		id, err := taskService.CreateTask(*taskqueue.NewTask(taskID, "python", "print('Hello, World!')"))
		if err != nil {
			log.Fatalf("Error creating task: %v", err)
		}

		if id != taskID {
			log.Fatalf("Task ID mismatch: %s != %s", id, taskID)
		}

		emsg, err := tmsg.Encode()
		if err != nil {
			log.Fatalf("Error encoding task message: %v", err)
		}

		mockTaskQueue.PublishTask(taskqueue.QueueMsg{
			Body: emsg,
		}, "python")

		taskIds = append(taskIds, taskID)
	}

	fmt.Println("Task IDs created : ", taskIds)

	podManager := connectToK8s()

	finished := []string{}
	w := sync.WaitGroup{}
	w.Add(4)

	scheduler := NewScheduler(SchedulerOpts{
		Consumer:    mockTaskQueue,
		PodManager:  podManager,
		TaskService: taskService,
		OnStartExec: func(lang, taskID string) {
			log.Printf("Task %s started", taskID)
		},
		OnEndExec: func(lang, taskID string) {
			log.Printf("Task %s completed", taskID)
			out, _, err := taskService.GetResult(taskID)
			if err != nil {
				log.Printf("Error getting task result: %v", err)
			}

			log.Printf("Output: %s", out)

			finished = append(finished, taskID)
			w.Done()
		},

		OnFailExec: func(lang, taskID string) {
			log.Printf("Task %s failed", taskID)

			w.Done()
		},
	}, "python")

	// Run StartWorkers in a goroutine
	go scheduler.StartWorkers()

	// Wait for the workers to process the tasks
	w.Wait()
	assert.ElementsMatch(t, taskIds, finished)
}
