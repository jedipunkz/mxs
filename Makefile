.PHONY: build test install clean

build:
	go build -o mxs .

test:
	go vet ./...
	go test ./...

install:
	go install .

clean:
	rm -f mxs
