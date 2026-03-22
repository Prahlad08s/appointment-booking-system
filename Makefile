.PHONY: run test test-unit test-race build clean tidy

run:
	go run main.go

build:
	go build -o bin/appointment-booking-system main.go

test:
	go test ./tests/... -v

test-unit:
	go test ./commons/... -v

test-race:
	go test ./tests/... -v -race

test-all:
	go test ./... -v

tidy:
	go mod tidy

clean:
	rm -rf bin/
