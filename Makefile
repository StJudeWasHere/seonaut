.PHONY: run test vet linux clean docker watch front

run:
	go run -race cmd/server/main.go -c config.local 

test:
	go test -race ./internal/...

vet:
	go vet ./internal/...

linux:
	GOOS=linux GOARCH=amd64 go build -o bin/seonaut cmd/server/main.go

clean:
	find ./bin ! -name .gitignore -delete

docker:
	docker-compose up -d --build

watch:
	esbuild ./web/css/style.css \
		--bundle \
		--outdir=./web/static \
		--public-path=/resources \
		--loader:.woff=file \
		--loader:.woff2=file \
		--watch

front:
	esbuild ./web/css/style.css \
		--bundle \
		--minify \
		--outdir=./web/static \
		--public-path=/resources \
		--loader:.woff=file \
		--loader:.woff2=file

watch:
	esbuild ./web/css/style.css \
		--bundle \
		--minify \
		--outdir=./web/static \
		--public-path=/resources \
		--loader:.woff=file \
		--loader:.woff2=file \
		--watch
