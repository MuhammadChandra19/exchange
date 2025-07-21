# Variables
MIGRATION_DIR=internal/infrastructure/questdb/migrations

# Default target
.DEFAULT_GOAL := help

.PHONY: migrate
migrate: ## Run all pending migrations
	@echo "Running migrations..."
	@go run cmd/migrate/main.go

.PHONY: migration
migration: ## Create a new migration file. Usage: make migration name=create_users_table
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter is required"; \
		echo "Usage: make migration name=create_users_table"; \
		exit 1; \
	fi
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	filename="$${timestamp}_$(name).sql"; \
	filepath="$(MIGRATION_DIR)/$${filename}"; \
	echo "Creating migration file: $${filename}"; \
	mkdir -p $(MIGRATION_DIR); \
	echo "-- Migration: $(name)" > $${filepath}; \
	echo "-- Created at: $$(date)" >> $${filepath}; \
	echo "" >> $${filepath}; \
	echo "-- Write your migration SQL here" >> $${filepath}; \
	echo "Migration file created: $${filepath}"

.PHONY: help
help: ## Show available commands
	@echo "Available commands:"
	@echo "  migrate              Run all pending migrations"
	@echo "  migration name=NAME  Create a new migration file"
	@echo ""
	@echo "Examples:"
	@echo "  make migrate"
	@echo "  make migration name=add_indexes" 