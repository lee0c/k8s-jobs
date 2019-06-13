package main

import (

)

const (

)

type JobManager struct {
	namespace			string
	connectionString	string
	queueName			string
	podSpecFilename		string
	queueManager		*servicebus.QueueManager
	informerFactory		*informers.SharedInformerFactory
	jobClient			*batchv1.JobClient
}

func NewJobManager(clientset *kubernetes.Clientset) {
	// get SBQ QueueManager

	// set up informer

	// get Job Client

	// return Manager
}

func (m *JobManager) createInformerFactory(clientset *kubernetes.Clientset, namespace string) {

}

func (m *JobManager) Run() {

}

func (m *JobManager) getQueueLength() {

}

func (m *JobManager) getActiveJobCount() {

}

func (m *JobManager) createJobs(n int32) {

}

func (m *JobManager) createJobSpec() {

}

func (m *JobManager) createPodSpec() {

}

func (m *JobManager) readFileToYaml() {
	
}