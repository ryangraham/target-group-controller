test:
	go test -v ./...

build:
	go build -o bin/target-group-controller cmd/controller/main.go

run:
	go run cmd/controller/main.go

image:
	KO_DOCKER_REPO=docker.io/ryangraham/target-group-controller ko publish --bare --push=false ./cmd/controller
