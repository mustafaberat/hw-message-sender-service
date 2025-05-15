DIRECTORIES := $(shell find . -name "go.mod" -maxdepth 2 -exec dirname {} \;)
DOCKER_ENV_FILES := $(shell find docker/env -name "*.env" -maxdepth 1 2>/dev/null || echo "")
APP_ENV_FILES := $(shell find . -name "*.env" -maxdepth 2)

.PHONY: test lint gofumpt

# Run tests
test:
	go test -v ./...

# Run only dependencies with Docker Compose
docker-run-dependencies:
	docker-compose -f docker/docker-compose.yml -p message-sender up -d --build postgres redis

# Format Go code using gofumpt
gofumpt:
	@echo "Running gofumpt on all directories"
	@command -v gofumpt >/dev/null 2>&1 || go install mvdan.cc/gofumpt@latest
	@for directory in $(DIRECTORIES); do \
		echo "Formatting directory: $$directory"; \
		find "$$directory" -name "*.go" -not -path "*/vendor/*" -exec $(HOME)/go/bin/gofumpt -l -w {} \; ; \
	done

lint:
	@for directory in $(DIRECTORIES); do \
		echo "Linting directory: $$directory ==="; \
		cd "$$directory" && make lint && cd .. || exit 1; \
	done
