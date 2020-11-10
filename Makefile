APP=idme
IMAGE=idme
DOCKER_ROOT=docker.astuart.co/andrew
NAMESPACE=default

FQTAG=$(DOCKER_ROOT)/$(IMAGE)

SHA=$(shell docker inspect --format "{{ index .RepoDigests 0 }}" $(1))

test:
	go test ./...

go:
	CGO_ENABLED=0 GOOS=linux go build -o app

docker: go test
	docker build -t $(FQTAG) . 
	docker push $(FQTAG)

deploy: docker
	kubectl apply -f k8s.yaml
	kubectl --namespace $(NAMESPACE) set image deployment/$(APP) $(APP)=$(call SHA,$(FQTAG))
