# Run the main server
run:
	go run -race cmd/server/main.go -c config.local 
.PHONY: run

# Run tests
test:
	go test -race ./internal/...
.PHONY: test

# Run vet
vet:
	go vet ./internal/...
.PHONY: vet

# Run docker compose
docker:
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d --build
.PHONY: docker

# Watch frontend files to run esbuild on any change
watch:
	esbuild ./web/css/style.css \
		--bundle \
		--outdir=./web/static \
		--public-path=/resources \
		--loader:.woff=file \
		--loader:.woff2=file \
		--loader:.png=file \
		--watch
.PHONY: watch

# Run esbuild to compile the frontend files
front:
	esbuild ./web/css/style.css \
		--bundle \
		--minify \
		--outdir=./web/static \
		--public-path=/resources \
		--loader:.woff=file \
		--loader:.woff2=file \
		--loader:.png=file
.PHONY: front
