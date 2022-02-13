all: build run

.PHONY: build

build:
	go build -o ./rsslay main.go

run:
	./rsslay