package main

import (
	"fmt"
	"path/filepath"

	"github.com/A7bari/RunWave/internal/app"
	coderunner "github.com/A7bari/RunWave/internal/code_runner"
	"github.com/A7bari/RunWave/internal/config"
	"github.com/A7bari/RunWave/internal/db"
	"github.com/A7bari/RunWave/internal/services"
	"github.com/A7bari/RunWave/internal/store"
	"github.com/A7bari/RunWave/internal/taskqueue"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	// Version is the version of the service
	Version = "v0.0.1"

	// Distributed task queue
	QueuePublisher taskqueue.QueuePublisher
	QueueConsumer  taskqueue.QueueConsumer

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

	// taskService is the task service
	taskService services.ITaskService

	// podManager is the pod manager
	podManager coderunner.PodManager
)

func main() {

	// setup k8s
	restConfig, clientset, err := setupK8s()
	if err != nil {
		panic(fmt.Sprintf("Error setting up k8s: %v", err))
	}

	serviceConf := config.GetConfig()

	// create the main store
	mStore = db.GetPostgresStore()
	taskService = services.NewTaskService(mStore)

	rbmq, err := db.GetRabbitMQConn()
	if err != nil {
		panic(fmt.Sprintf("Error connecting to RabbitMQ: %v", err))
	}

	QueuePublisher = rbmq
	QueueConsumer = rbmq

	// create the task qu
	// create the pod manager
	podManager = coderunner.NewPodManager(clientset, restConfig, serviceConf.Namespace)

	// create a scheduler
	s := coderunner.NewScheduler(
		coderunner.SchedulerOpts{
			Consumer:    QueueConsumer,
			TaskService: taskService,
			PodManager:  podManager,
			OnEndExec: func(lang, taskId string) {
				fmt.Printf("Scheduler: Task %s in language: %s\n", taskId, lang)
			},

			OnFailExec: func(lang, taskId string) {
				fmt.Printf("Scheduler [failed task]: Task %s \n", taskId)
			},
		},
		config.GetConfig().Languages...,
	)

	go s.StartWorkers()

	// start the server
	server = app.NewServer(app.ServerOpts{
		TaskService:   taskService,
		TaskPublisher: QueuePublisher,
	})
	server.Start()
}

func setupK8s() (*rest.Config, *kubernetes.Clientset, error) {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	// Build the config from the kubeconfig file
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("Error building kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating Kubernetes client: %w", err)
	}

	return restConfig, clientset, nil
}
