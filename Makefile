all: build run

build:
	go build -o reform ./cmd/reform.go

run:
	./reform