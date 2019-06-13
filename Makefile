all: consumer manager

consumer:
	docker build -t lee0c/sbq-consumer ./consumer
	docker push lee0c/sbq-consumer

manager:
	docker build -t lee0c/job-manager ./manager
	docker push lee0c/job-manager