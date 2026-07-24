IMG ?= ghcr.io/your-username/olm-update-sentinel:latest

.PHONY: all build test docker-build docker-push deploy

all: test build

test:
	go test ./... -coverprofile cover.out

build:
	go build -o bin/manager cmd/main.go

docker-build:
	docker build -t $(IMG) .

docker-push:
	docker push $(IMG)

deploy:
	kustomize build config/default | kubectl apply -f -