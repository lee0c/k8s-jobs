package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	envQueueName = "SERVICEBUS_QUEUE_NAME"
	envConnectionString = "SERVICEBUS_CONNECTION_STRING"
	envNamespace = "POD_NAMESPACE"
	envPodSpecFile = "PODSPEC_FILENAME"
	envJobLabel = "JOB_APP_LABEL"
)

func main() {
	clientset, err := getClientset()
	if err != nil {
		log.Errorf("Cannot get clientset: %v", err)
	}

	manager, err := NewJobManager(clientset, os.Getenv(envNamespace), os.Getenv(envConnectionString), os.Getenv(envQueueName), os.Getenv(envPodSpecFile), os.Getenv(envJobLabel))
	if err != nil {
		log.Errorf("Failed to create manager: %v", err)
		os.Exit(1)
	}

	stopCh := make(chan struct{})
	manager.Run(stopCh)

	<-stopCh
}

func getClientset() (*kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}