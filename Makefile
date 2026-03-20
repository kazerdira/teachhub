.PHONY: run setup db-up db-down

# Start PostgreSQL
db-up:
	docker compose up -d

# Stop PostgreSQL
db-down:
	docker compose down

# Install deps and run
run:
	go mod tidy
	go run .

# Full setup: start DB, wait, run app
setup: db-up
	@echo "Waiting for PostgreSQL..."
	@sleep 3
	$(MAKE) run
