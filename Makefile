.PHONY: run worker test clean

run:
	go run cmd/api/main.go

worker:
	go run cmd/worker/main.go

test:
	go test -v ./tests/...

clean:
	rm -rf dist