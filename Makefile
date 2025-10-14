DB_URL = sqlite://data/storage.db
MIGRATIONS_DIR = db/migrations
TAILWIND = tailwindcss-linux-x64
TAILWIND_INPUT =static/input.css
TAILWIND_OUTPUT = static/output.css

.PHONY: all
all: migrate-up tailwind

.PHONY: migrate-up
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "❌ Please provide a migration name: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	migrate create -ext sql -dir $(MIGRATIONS_DIR) $(name)

.PHONY: tailwind
tailwind:
	npm run build

.PHONY: tailwind-watch
tailwind-watch:
	$(TAILWIND) -i $(TAILWIND_INPUT) -o $(TAILWIND_OUTPUT) --watch


.PHONY: test
test:
	go test -v ./...

.PHONY: test-cover
test-cover:
	go test -v -coverprofile=coverage.out ./...

.PHONY: cover-html
cover-html: test-cover
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

.PHONY: test-cover-service
test-cover-service:
	go test -v -coverprofile=coverage-service.out ./internal/service

.PHONY: test-cover-service-html
test-cover-service-html: test-cover-service
	go tool cover -html=coverage-service.out -o coverage-service.html
	@echo "✅ Coverage report generated: coverage-service.html"

.PHONY: test-cover-repo
test-cover-repo:
	go test -v -coverprofile=coverage-repo.out ./internal/repository

.PHONY: test-cover-repo-html
test-cover-repo-html: test-cover-repo
	go tool cover -html=coverage-repo.out -o coverage-repo.html
	@echo "✅ Coverage report generated: coverage-repo.html"

