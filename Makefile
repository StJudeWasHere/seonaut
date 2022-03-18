.PHONY: run test linux docker

run:
	go run cmd/server/main.go -c config.local 

test:
	go test github.com/stjudewashere/seonaut/test

linux:
	GOOS=linux GOARCH=amd64 go build -o build/seonaut cmd/server/main.go

docker:
	docker-compose up -d --build
