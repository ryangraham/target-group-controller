test:
	go test -v ./...

build:
	go build -o bin/target-group-controller cmd/controller/main.go

run:
	go run cmd/controller/main.go
