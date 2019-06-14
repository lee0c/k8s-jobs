package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	servicebus "github.com/Azure/azure-service-bus-go"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	batch_typed "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/tools/cache"
)

// JobManager comment
type JobManager struct {
	Namespace			string
	ConnectionString	string
	QueueName			string
	PodSpecFilename		string
	JobLabel			string
	QueueManager		*servicebus.QueueManager
	InformerFactory		informers.SharedInformerFactory
	JobClient			batch_typed.JobInterface
}

// NewJobManager comment
func NewJobManager(clientset *kubernetes.Clientset, namespace, connectionString, queueName, podSpecFilename, jobLabel string) (*JobManager, error) {
	// get SBQ QueueManager
	sbNamespace, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(connectionString))
	if err != nil {
		return nil, err
	}

	queueManager := sbNamespace.NewQueueManager()

	// set up informer
	informerFactory := createInformerFactory(clientset, namespace, jobLabel)

	// get Job Client
	jobClient := clientset.BatchV1().Jobs(namespace)

	// return Manager
	return &JobManager{
		Namespace: namespace,
		ConnectionString: connectionString,
		QueueName: queueName,
		PodSpecFilename: podSpecFilename,
		JobLabel: jobLabel,
		QueueManager: queueManager,
		InformerFactory: informerFactory,
		JobClient: jobClient,
	}, nil
}

func createInformerFactory(clientset *kubernetes.Clientset, namespace, jobLabel string) informers.SharedInformerFactory {
	sharedInformerFactory := informers.NewSharedInformerFactory(clientset, time.Second * 30)
	jobInformer := sharedInformerFactory.Batch().V1().Jobs().Informer()

	jobInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{

		UpdateFunc: func(oldObj, newObj interface{}) {
			new := newObj.(*batchv1.Job)
			if new.Status.CompletionTime != nil && new.ObjectMeta.Labels["app"] == jobLabel {
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
	return sharedInformerFactory
}

// Run comment
func (m *JobManager) Run(stopCh <-chan struct{}) {
	go m.InformerFactory.Start(stopCh)

	jobLenCh := make(chan int)
	queueLenCh := make(chan int)

	for {
		go m.getActiveJobCount(jobLenCh)
		go m.getQueueLength(queueLenCh)
		
		jobLen := <- jobLenCh
		queueLen := <- queueLenCh

		if jobLen != -1 && queueLen != -1 {
			err := m.createJobs(queueLen - jobLen)
			if err != nil {
				log.Errorf("Error with job creation: %v", err)
			}	
		} else {
			log.Errorf("Failed to get job count or queue length")
		}
	}
}

func (m *JobManager) getQueueLength(resultCh chan<- int) {
	queueEntity, err := m.QueueManager.Get(context.TODO(), m.QueueName)
	if err != nil {
		resultCh<- -1
		return
	}

	queueLen := int(*queueEntity.CountDetails.ActiveMessageCount)
	resultCh<- queueLen
	log.Infof("Returned queue length %d", queueLen)
	return
}

func (m *JobManager) getActiveJobCount(resultCh chan<- int) {
	
	labelSelector := fmt.Sprintf("app=%s", m.JobLabel)
	jobList, err := m.JobClient.List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		resultCh<- -1
		return
	}

	jobLen := len(jobList.Items)
	resultCh<- jobLen
	log.Infof("Returned job count %d", jobLen)
	return
}

func (m *JobManager) createJobs(n int) error {
	if n == 0 {
		log.Infof("No job creation needed")
		return nil
	}

	job, err := m.createJob()
	if err != nil {
		return err
	}

	errCount := 0

	for i := 0; i < n; i++ {
		_, err := m.JobClient.Create(job)
		if err != nil {
			errCount++
		}
	}

	log.Infof("Created %d jobs with %d failures", n - errCount, errCount)
	return nil
}

func (m *JobManager) createJob() (*batchv1.Job, error) {
	podSpec, err := m.createPodSpec()
	if err != nil {
		return nil, err
	}

	nameBase := fmt.Sprintf("%s-job-", podSpec.Containers[0].Name)

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: nameBase,
			Namespace: m.Namespace,
			Labels: map[string]string{
				"app": m.JobLabel,
			},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: *podSpec,
			},
		},
	}

	return &job, nil
}

func (m *JobManager) createPodSpec() (*v1.PodSpec, error) {
	podSpec, err := ReadYamlFileToPodSpec(m.PodSpecFilename)
	if err != nil {
		return nil, err
	}
	podSpec.RestartPolicy = v1.RestartPolicyOnFailure

	return podSpec, nil
}

// ReadYamlFileToPodSpec comment
func ReadYamlFileToPodSpec(filename string) (*v1.PodSpec, error) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Errorf("Failed to read podSpec file: %v", err)
		return nil, err
	}

	podSpec := v1.PodSpec{}
	err = yaml.Unmarshal(dat, &podSpec)
	if err != nil {
		log.Errorf("Failed to unmarshal podSpec yaml: %v", err)
		return nil, err
	}

	return &podSpec, nil
}