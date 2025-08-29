DB_URL = sqlite://data/storage.db
MIGRATIONS_DIR = ./backend/db/migrations
TAILWIND = tailwindcss-linux-x64
TAILWIND_INPUT = backend/static/input.css
TAILWIND_OUTPUT = backend/static/output.css

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