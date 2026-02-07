.RECIPEPREFIX := $() $()

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/furlongm/patchman-client/go/lib/report.Version=$(VERSION)"

format:
    find -name *.go -exec go fmt {} \;

build:
    CGO_ENABLED=0 go build $(LDFLAGS) -o patchman-client-go patchman-client/patchman-client.go

build-linux-amd64:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/patchman-client-linux-amd64 patchman-client/patchman-client.go

build-linux-arm64:
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/patchman-client-linux-arm64 patchman-client/patchman-client.go

build-all: build-linux-amd64 build-linux-arm64

all: format build

install:
    go install $(LDFLAGS) patchman-client/patchman-client.go

get_modules:
    go clean -modcache
    go get -u all

test:
    go test ./...

clean:
    rm -f patchman-client-go bin/patchman-client-*
