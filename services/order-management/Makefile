# Variables
MIGRATION_DIR=internal/infrastructure/postgresql/migrations

# Default target
.DEFAULT_GOAL := help

.PHONY: migrate
migrate: ## Run all pending migrations (alias for migrate-up)  
	@echo "Running all pending migrations..."
	@go run cmd/migrate/main.go -direction=up -steps=0

.PHONY: migrate-up
migrate-up: ## Run pending migrations up. Usage: make migrate-up [steps=N]
	@echo "Running migrations up..."
	@if [ -z "$(steps)" ]; then \
		go run cmd/migrate/main.go -direction=up -steps=0; \
	else \
		go run cmd/migrate/main.go -direction=up -steps=$(steps); \
	fi

.PHONY: migrate-down  
migrate-down: ## Run migrations down. Usage: make migrate-down steps=N
	@if [ -z "$(steps)" ]; then \
		echo "Error: steps parameter is required for down migrations"; \
		echo "Usage: make migrate-down steps=N"; \
		exit 1; \
	fi
	@echo "Running $(steps) migrations down..."
	@go run cmd/migrate/main.go -direction=down -steps=$(steps)

.PHONY: migration
migration: ## Create a new migration file. Usage: make migration name=create_users_table
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter is required"; \
		echo "Usage: make migration name=create_users_table"; \
		exit 1; \
	fi
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	basename="$${timestamp}_$(name)"; \
	upfile="$(MIGRATION_DIR)/$${basename}.up.sql"; \
	downfile="$(MIGRATION_DIR)/$${basename}.down.sql"; \
	echo "Creating migration files: $${basename}.up.sql and $${basename}.down.sql"; \
	mkdir -p $(MIGRATION_DIR); \
	echo "-- Migration: $(name)" > $${upfile}; \
	echo "-- Created at: $$(date)" >> $${upfile}; \
	echo "" >> $${upfile}; \
	echo "-- Write your UP migration SQL here" >> $${upfile}; \
	echo "-- Migration: $(name)" > $${downfile}; \
	echo "-- Created at: $$(date)" >> $${downfile}; \
	echo "" >> $${downfile}; \
	echo "-- Write your DOWN migration SQL here" >> $${downfile}; \
	echo "Migration files created:"; \
	echo "  UP:   $${upfile}"; \
	echo "  DOWN: $${downfile}"

.PHONY: help
help: ## Show available commands
	@echo "Available commands:"
	@echo "  migrate              Run all pending migrations (alias for migrate-up)"
	@echo "  migrate-up           Run pending migrations up (optional: steps=N)"
	@echo "  migrate-down         Run migrations down (required: steps=N)"
	@echo "  migration name=NAME  Create new migration files (.up.sql and .down.sql)"
	@echo ""
	@echo "Examples:"
	@echo "  make migrate                    # Run all pending migrations"
	@echo "  make migrate-up                 # Run all pending migrations"  
	@echo "  make migrate-up steps=3         # Run next 3 pending migrations"
	@echo "  make migrate-down steps=2       # Rollback last 2 migrations"
	@echo "  make migration name=add_indexes # Create new migration files (.up.sql/.down.sql)"
	