.RECIPEPREFIX += 

build:
    go build -o patchman-client-go patchman-client/patchman-client.go

install:
    go install patchman-client/patchman-client.go

format:
    find -name *.go -exec go fmt {} \;

get_modules:
    go clean -modcache
    go get -u all
