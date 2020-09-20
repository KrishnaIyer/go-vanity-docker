.PHONY: build

GO_VANITY_DOCKER_VERSION=v0.0.1
GO_VANITY_DOCKER_GIT_COMMIT=$(shell git rev-parse --short HEAD)
GO_VANITY_DOCKER_DATE=$(shell date)
GO_VANITY_DOCKER_PACKAGE="github.com/krishnaiyer/go-vanity-docker"

test:
	go test ./... -cover

build.local:
	go build \
	-ldflags="-X '${GO_VANITY_DOCKER_PACKAGE}/cmd.version=${GO_VANITY_DOCKER_VERSION}' \
	-X '${GO_VANITY_DOCKER_PACKAGE}/cmd.gitCommit=${GO_VANITY_DOCKER_GIT_COMMIT}' \
	-X '${GO_VANITY_DOCKER_PACKAGE}/cmd.buildDate=${GO_VANITY_DOCKER_DATE}'" main.go

build.dist:
	goreleaser --snapshot --skip-publish --rm-dist

clean:
	rm -rf dist
