GOLANG_VERSION ?= 1.21
IMAGE_NAME ?= iuriikogan-snyk/snyk-kube-watcher

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o snyk-kube-watcher ./cmd/main.go

.PHONY: test
test:
	go test ./... -coverprofile=coverage.out

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: docker-build
docker-build:
	docker build -t iuriikogan/snyk-kube-watcher:latest .
docker-push: docker-build
	docker push iuriikogan/snyk-kube-watcher:latest
