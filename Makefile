
include .env

.PHONY: init

GO_VANITY_DOCKER_GIT_COMMIT=$(shell git rev-parse --short HEAD)
GO_VANITY_DOCKER_DATE=$(shell date)
GO_VANITY_DOCKER_DIST_FOLDER=dist

init:
	mkdir -p ${GO_VANITY_DOCKER_DIST_FOLDER}

test:
	go test ./... -cover

build.local:
	go build \
	-ldflags="-X '${GO_VANITY_DOCKER_PACKAGE}/cmd.version=${GO_VANITY_DOCKER_VERSION}' \
	-X '${GO_VANITY_DOCKER_PACKAGE}/cmd.gitCommit=${GO_VANITY_DOCKER_GIT_COMMIT}' \
	-X '${GO_VANITY_DOCKER_PACKAGE}/cmd.buildDate=${GO_VANITY_DOCKER_DATE}'" main.go

build.docker:
	GOOS=linux GOARCH=amd64 \
	go build \
	-ldflags="-X '${GO_VANITY_DOCKER_PACKAGE}/cmd.version=${GO_VANITY_DOCKER_VERSION}' \
	-X '${GO_VANITY_DOCKER_PACKAGE}/cmd.gitCommit=${GO_VANITY_DOCKER_GIT_COMMIT}' \
	-X '${GO_VANITY_DOCKER_PACKAGE}/cmd.buildDate=${GO_VANITY_DOCKER_DATE}'" \
	-o "${GO_VANITY_DOCKER_DIST_FOLDER}/go-vanity" \
	main.go
	docker build -t ${GO_VANITY_DOCKER_IMAGE}:${GO_VANITY_DOCKER_VERSION} .

clean:
	rm -rf dist
