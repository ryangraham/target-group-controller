test:
	go1.23.0 test -v ./...

build:
	go1.23.0 build -o bin/target-group-controller cmd/controller/main.go

run:
	go1.23.0 run cmd/controller/main.go

image:
	KO_GO_PATH=go1.23.0 KO_DOCKER_REPO=docker.io/ryangraham/target-group-controller ko publish --bare ./cmd/controller

image-local:
	KO_GO_PATH=go1.23.0 KO_DOCKER_REPO=docker.io/ryangraham/target-group-controller ko publish --push=false --bare ./cmd/controller

crds:
	kubectl apply -f pkg/api/crds/targetgroupbindings.yaml

example:
	kubectl apply -f examples/v1/private.yaml

helm-install:
	helm install target-group-controller ./charts/target-group-controller

helm-upgrade:
	helm upgrade target-group-controller ./charts/target-group-controller

helm-uninstall:
	helm uninstall target-group-controller

helm-lint:
	helm lint --strict charts/target-group-controller
