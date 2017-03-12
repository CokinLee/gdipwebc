all: gdipwebc

gdipwebc:
	go build

install:
	go install

format:
	find * -type f -name '*.go' -print0 | xargs -0n1 go fmt
