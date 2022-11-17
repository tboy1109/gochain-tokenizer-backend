.PHONY: dep build docker release install test backup

build:
	go build -o app

run: build
	./app
