BINARY_NAME=go-serial-tty-client

.PHONY: all build build-linux run clean

all: build

build:
	CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "-static"' -o $(BINARY_NAME) main.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags '-s -w -extldflags "-static"' -o $(BINARY_NAME)-linux-arm64 main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-s -w -extldflags "-static"' -o $(BINARY_NAME)-linux-amd64 main.go

run: build
	./$(BINARY_NAME) -sim

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux-arm64
	rm -f $(BINARY_NAME)-linux-amd64
