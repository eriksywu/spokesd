BUILD_DATE = `date +%FT%T%z`
LDFLAGS := "-X 'github.com/eriksywu/spokesd/cmd.buildDate=$(BUILD_DATE)'"

default: build test

.PHONY: build
build:
	go build -ldflags ${LDFLAGS} -o spokesd

.PHONY: test
test:
	go test ./...