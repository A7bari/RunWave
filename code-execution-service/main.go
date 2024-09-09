package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/A7bari/RunWave/internal"
	"github.com/A7bari/RunWave/internal/app"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	// Build the config from the kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	serviceConf := &internal.ServiceConfig{
		Namespace: "code-exec-system",
		Commands: map[string]string{
			"python":     "-c",
			"javascript": "node -e",
		},
		GetPodRetries: 4,
		Standbylabel:  "app=standby-pod",
		InUseLabel:    "in-use-pod",
		Timeout:       6 * time.Second,
	}

	// setup pod manager
	podManager := internal.NewPodManager(config, serviceConf)

	// start the server
	server := app.NewServer(podManager)
	server.Start()
}
