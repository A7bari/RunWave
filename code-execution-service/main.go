package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/A7bari/RunWave/internal/app"
	coderunner "github.com/A7bari/RunWave/internal/code_runner"
	"github.com/A7bari/RunWave/internal/config"
	"github.com/A7bari/RunWave/internal/db"
	"github.com/A7bari/RunWave/internal/store"
	"github.com/A7bari/RunWave/internal/taskqueue"
	"github.com/A7bari/RunWave/internal/types"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	// Version is the version of the service
	Version = "v0.0.1"

	// taskQueues is a map of task queues for each language
	taskQueues = make(map[string]taskqueue.TaskQueue)

	// schedulers is a map of schedulers for each language
	schedulers = make(map[string]*coderunner.Scheduler)

	// RestConfig is the Kubernetes REST config
	restConfig *rest.Config

	// clientset is the Kubernetes client
	clientset *kubernetes.Clientset

	// server is the main server
	server *app.Server

	// mStore is the main store
	mStore store.Store

	// podManager is the pod manager
	podManager coderunner.PodManager
)

func main() {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	var err error
	// Build the config from the kubeconfig file
	restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	clientset, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	serviceConf := config.GetConfig()

	// create the main store
	mStore = db.GetInMemStore()

	// create the pod manager
	podManager = coderunner.NewPodManager(clientset, restConfig, serviceConf.Namespace)

	// create task queues and schedulers for each language
	for _, lang := range serviceConf.Languages {
		q := db.GetInMemTaskQueue(100)
		taskQueues[lang] = q

		// create a scheduler
		s := coderunner.NewScheduler(coderunner.SchedulerOpts{
			TaskQueue:  q,
			PodManager: podManager,
			OnEndExec: func(task taskqueue.Task) {
				output, isError := task.GetResult()
				fmt.Printf("Scheduler: Task %s ended with output: %s\n", task.GetTaskID(), output)
				mStore.SaveResult(types.TaskOutput{
					TaskID:  task.GetTaskID(),
					Output:  output,
					Status:  task.GetStatus(),
					IsError: isError,
					Lang:    task.GetLanguage(),
				})
			},

			OnFailExec: func(task taskqueue.Task) {
				fmt.Printf("Scheduler: Task %s failed with err: %s\n", task.GetTaskID(), task.GetError())
			},
		})

		schedulers[lang] = s

		go s.StartWorkers()
	}

	// start the server
	server = app.NewServer(app.ServerOpts{
		Queues: taskQueues,
	})
	server.Start()
}
