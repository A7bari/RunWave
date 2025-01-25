package config

import (
	"flag"
	"strings"
	"sync"
)

// ServiceConfig holds the configuration for the code-execution-service

type ServiceConfig struct {
	Namespace    string
	Commands     map[string]string
	Standbylabel string
	InUseLabel   string
	Languages    []string
	QueueAdrr    string
	QueueName    string
}

var (
	config     *ServiceConfig
	configOnce sync.Once
)

// GetConfig returns the service configuration
func GetConfig() *ServiceConfig {
	configOnce.Do(func() {
		config = loadConfig()
	})
	return config
}

func loadConfig() *ServiceConfig {
	// Load the configuration from command line flags or environment variables
	namespace := flag.String("namespace", "code-exec-system", "The namespace to use for the code execution pods")
	standbyLabel := flag.String("standby-label", "app=standby-pod", "The label to use for standby pods")
	inUseLabel := flag.String("in-use-label", "in-use-pod", "The label to use for in-use pods")
	langs := flag.String("languages", "python,javascript", "The languages to support")
	qAdrr := flag.String("queue-addr", "amqp://guest:guest@localhost:5672/", "The RabbitMQ address")
	qName := flag.String("queue-name", "task-queue", "The RabbitMQ queue name")

	flag.Parse()

	return &ServiceConfig{
		Namespace:    *namespace,
		Commands:     map[string]string{"python": "-c", "javascript": "node -e"},
		Standbylabel: *standbyLabel,
		InUseLabel:   *inUseLabel,
		Languages:    strings.Split(*langs, ","),
		QueueAdrr:    *qAdrr,
		QueueName:    *qName,
	}
}
