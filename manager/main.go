package main

import (
	"os"

	"k8s.io/client-go/kubernetes"
)

const (
	envQueueName = "SERVICEBUS_QUEUE_NAME"
	envConnectionString = "SERVICEBUS_CONNECTION_STRING"
	envNamespace = "POD_NAMESPACE"
	envPodSpecFile = "PODSPEC_FILENAME"
)

func main() {
	clientset := getClientset()

	manager, err := NewJobManager(clientset)
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