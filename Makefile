.PHONY: manager

manager:
	docker build -t lee0c/job-manager ./manager
	docker push lee0c/job-manager