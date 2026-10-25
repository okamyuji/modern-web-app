.PHONY: dev build clean

dev:
	@echo "Starting development server..."
	@pnpm run dev &
	@air

build:
	@echo "Building application..."
	@templ generate
	@pnpm run build
	@go build -o ./bin/server ./cmd/server

clean:
	@rm -rf ./tmp ./bin