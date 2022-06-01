.PHONY: run test vet linux clean docker

run:
	go run -race cmd/server/main.go -c config.local 

test:
	go test ./internal/...

vet:
	go vet ./internal/...

linux:
	GOOS=linux GOARCH=amd64 go build -o bin/seonaut cmd/server/main.go

clean:
	find ./bin ! -name .gitignore -delete

docker:
	docker-compose up -d --build
