.PHONY: build run test clean docker-build docker-run swagger

docker-run-dependencies:
	docker-compose -f ../docker/docker-compose.yml -p message-sender up -d --build postgres redis

# Generate Swagger docs
swagger:
	@swag init -g transport/http/docs/meta.go -o transport/http/docs

gofumpt:
	$(HOME)/go/bin/gofumpt -l -w .
	@command -v swag >/dev/null 2>&1 && swag fmt || echo "swag command not found, skipping swagger formatting"

lint:
	golangci-lint run --config ../.golangci.yml
