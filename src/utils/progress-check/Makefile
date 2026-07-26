BINARY_NAME=progress-check
GOOS=linux
GOARCH=amd64
LDFLAGS="-s -w"

.PHONY: all build garble upx release clean setup check-garble check-upx

all: release

check-garble:
	@command -v garble >/dev/null 2>&1 || { echo >&2 "Error: garble is not installed. Install it"; exit 1; }

check-upx:
	@command -v upx >/dev/null 2>&1 || { echo >&2 "Error: upx is not installed. Install it."; exit 1; }

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags=$(LDFLAGS) -o $(BINARY_NAME) main.go

garble: check-garble
	GOOS=$(GOOS) GOARCH=$(GOARCH) garble -literals build -ldflags=$(LDFLAGS) -o $(BINARY_NAME) main.go

upx: check-upx
	upx --brute $(BINARY_NAME)

release: garble upx

clean:
	rm -f $(BINARY_NAME)
