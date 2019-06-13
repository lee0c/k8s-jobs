package main

import (
	"context"
	"io/ioutil"

	servicebus "github.com/Azure/azure-service-bus-go"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

const (

)

type JobManager struct {
	Namespace			string
	ConnectionString	string
	QueueName			string
	PodSpecFilename		string
	QueueManager		*servicebus.QueueManager
	InformerFactory		*informers.SharedInformerFactory
	JobClient			*batchv1.JobClient
}

func NewJobManager(clientset *kubernetes.Clientset, namespace, connectionString, queueName, podSpecFilename string) {
	// get SBQ QueueManager
	sbNamespace, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(s.metadata.connection))
	if err != nil {
		return -1, err
	}
	queueManager := ns.NewQueueManager()

	// set up informer
	informerFactory := createInformerFactory(clientset, namespace)

	// get Job Client
	jobClient := clientset.BatchV1().Jobs(namespace)

	// return Manager
	return &JobManager{
		Namespace: namespace,
		ConnectionString: connectionString,
		QueueName: queueName,
		PodSpecFilename: podSpecFilename,
		QueueManager: queueManager,
		InformerFactory: informerFactory,
		JobClient: jobClient,
	}
}

func (m *JobManager) createInformerFactory(clientset *kubernetes.Clientset, namespace string) {
	sharedInformerFactory := informers.NewSharedInformerFactory(clientset, time.Second * 30)
	jobInformer := sharedInformerFactory.Batch().V1().Jobs().Informer()

	jobInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{

		UpdateFunc: func(oldObj, newObj interface{}) {
			new := newObj.(*batch_v1.Job)
			if new.Status.CompletionTime != nil {
				// TODO: job labels
				propagationPolicy := metav1.DeletePropagationBackground
				err := clientset.BatchV1().Jobs(namespace).Delete(new.Name, &metav1.DeleteOptions{
					PropagationPolicy: &propagationPolicy,
				})
				if err != nil {
					log.Errorf("Failed to delete job %s", new.Name)
				}
				log.Infof("Cleaned up job %s", new.Name)
			}
		},
	})
}

func (m *JobManager) Run(chan stopCh) {
	go m.informerFactory.Start(stopCh)

	jobLenCh := make(chan int32)
	queueLenCh := make(chan int32)

	for {
		go m.getActiveJobCount(jobLenCh)
		go m.getQueueLength(queueLenCh)
		
		jobLen := <- jobLenCh
		queueLen := <- queueLenCh

		err = m.createJobs(queueLen - jobLen)
		if err != nil {
			log.Errorf("Error with job creation: %v", err)
		}
	}
}

func (m *JobManager) getQueueLength() int32 {

	return -1
}

func (m *JobManager) getActiveJobCount() int32 {

	// get active job count
	return -1
}

func (m *JobManager) createJobs(n int32) error {
	if n == 0 {
		log.Infof("No job creation needed")
		return nil
	}




	log.Infof("Created %d jobs", n)
	return nil
}

func (m *JobManager) createJobSpec() *batchv1.JobSpec {

}

func (m *JobManager) createPodSpec() *v1.PodSpec {
	podSpec := readFileToYaml(m.PodSpecFilename)
}

func readYamlFileToPodSpec(filename string) {
}