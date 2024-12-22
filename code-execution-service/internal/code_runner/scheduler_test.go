package coderunner

import (
	"log"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/A7bari/RunWave/internal/db"
	"github.com/A7bari/RunWave/internal/taskqueue"
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
	mockTaskQueue := db.GetInMemTaskQueue(10)

	for i := 1; i <= 4; i++ {
		mockTaskQueue.AddTask(taskqueue.NewTaskBuilder().
			SetID("task" + strconv.Itoa(i)).
			SetLang("python").
			SetCode("print('Hello, World!')").
			Build())
	}

	podManager := connectToK8s()

	finished := []string{}
	w := sync.WaitGroup{}
	w.Add(4)

	scheduler := NewScheduler(SchedulerOpts{
		TaskQueue:  mockTaskQueue,
		PodManager: podManager,
		OnStartExec: func(task taskqueue.Task) {
			log.Printf("Task %s started", task.GetTaskID())
		},
		OnEndExec: func(task taskqueue.Task) {
			log.Printf("Task %s completed", task.GetTaskID())
			out, _ := task.GetResult()
			log.Printf("Output: %s", out)

			finished = append(finished, task.GetTaskID())
			w.Done()
		},

		OnFailExec: func(task taskqueue.Task) {
			log.Printf("Task %s failed", task.GetTaskID())

			w.Done()
		},
	})

	// Run StartWorkers in a goroutine
	go scheduler.StartWorkers()

	// Wait for the workers to process the tasks
	w.Wait()
	assert.ElementsMatch(t, []string{"task1", "task2", "task3", "task4"}, finished)
}
