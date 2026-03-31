build:
	go build -o bin/hevy .

install:
	go install .

test:
	go test ./...

release:
	goreleaser release --clean
