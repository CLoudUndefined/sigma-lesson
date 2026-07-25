BINARY_NAME=container-generator
GOOS=linux
GOARCH=amd64
LDFLAGS="-s -w"

.PHONY: all build garble upx release clean setup

all: release

setup:
	go install mvdan.cc/garble@latest

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags=$(LDFLAGS) -o $(BINARY_NAME) main.go

garble:
	GOOS=$(GOOS) GOARCH=$(GOARCH) garble -literals build -ldflags=$(LDFLAGS) -o $(BINARY_NAME) main.go

upx:
	upx --brute $(BINARY_NAME)

release: garble upx

clean:
	rm -f $(BINARY_NAME)
