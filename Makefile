.PHONY: dev build clean

dev:
	@mkdir -p tmp	
	@cp ./.env ./tmp
	@air

build:
	@go build -o build/agora ./src/main.go

build-linux: ## Build for linux with current date in filename
	$(eval DATE := $(shell date +%Y-%m-%d))
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-linux-musl-gcc  CXX=x86_64-linux-musl-g++ go build --ldflags '-linkmode external -extldflags "-static"' -o build/agora_$(DATE) ./src/main.go

clean:
	rm -rf tmp

publish: build-linux ## Build with date and upload
	$(eval DATE := $(shell date +%Y-%m-%d))
	scp ./build/agora_$(DATE) agora:~/agora