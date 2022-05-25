.PHONY: run test vet linux docker

run:
	go run cmd/server/main.go -c config.local 

test:
	go test ./test/...

vet:
	go vet ./internal/...

linux:
	GOOS=linux GOARCH=amd64 go build -o bin/seonaut cmd/server/main.go

docker:
	docker-compose up -d --build
