.PHONY: dev build clean

dev:
	@mkdir -p tmp	
	@cp ./.env ./tmp
	@air

build:
	@go build -o build/agora ./src/main.go

clean:
	rm -rf tmp