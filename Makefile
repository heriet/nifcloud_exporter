all: ensure build

ensure:
	dep ensure

build:
	go build