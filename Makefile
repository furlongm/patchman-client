.RECIPEPREFIX := $() $()

format:
    find -name *.go -exec go fmt {} \;

build:
    go build -o patchman-client-go patchman-client/patchman-client.go

all: format build

install:
    go install patchman-client/patchman-client.go

get_modules:
    go clean -modcache
    go get -u all
