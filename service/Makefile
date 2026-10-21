BUILDDATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo HEAD)

.PHONY: all deps build run test lint fmt tidy clean

all: build

deps:
	go mod tidy

build:
	go build -o orgviz -ldflags "-s -w \
		-X 'github.com/jamesread/orgviz/internal/buildinfo.Commit=$(GIT_COMMIT)' \
		-X 'github.com/jamesread/orgviz/internal/buildinfo.BuildDate=$(BUILDDATE)' \
		-X 'github.com/jamesread/orgviz/internal/buildinfo.Version=$(GIT_TAG)' \
		"

run:
	./orgviz -configdir ../var/config-skel

test:
	go test ./...

lint:
	go vet ./...
	@unformatted=$$(gofmt -l $$(find . -name '*.go' -not -path './gen/*')); \
	if [ -n "$$unformatted" ]; then echo "gofmt needed (run make fmt):"; echo "$$unformatted"; exit 1; fi
	gocyclo -over 4 -ignore '(^|/)gen(/|$$)' .

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './gen/*')

tidy:
	go mod tidy

clean:
	rm -f orgviz
