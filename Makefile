test:
	go1.23.0 test -v ./...

build:
	go1.23.0 build -o bin/target-group-controller cmd/controller/main.go

run:
	go1.23.0 run cmd/controller/main.go

image:
	KO_GO_PATH=go1.23.0 KO_DOCKER_REPO=docker.io/ryangraham/target-group-controller ko publish --bare ./cmd/controller
