.PHONY: dev build clean

dev:
	@echo "Starting development server..."
	@npm run dev &
	@air

build:
	@echo "Building application..."
	@templ generate
	@npm run build
	@go build -o ./bin/server ./cmd/server

clean:
	@rm -rf ./tmp ./bin