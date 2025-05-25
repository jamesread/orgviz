BUILDDATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo HEAD)

default:
	go build -o orgviz -ldflags "-s -w \
		-X 'github.com/jamesread/orgviz/internal/buildinfo.Commit=$(GIT_COMMIT)' \
		-X 'github.com/jamesread/orgviz/internal/buildinfo.BuildDate=$(BUILDDATE)' \
		-X 'github.com/jamesread/orgviz/internal/buildinfo.Version=$(GIT_TAG)' \
		"

codestyle:
	go fmt ./...
	go vet ./...
	gocritic check ./...
	gocyclo -over 3 .

.PHONY: default codestyle
