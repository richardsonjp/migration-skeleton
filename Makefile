# Variables
SERVICE_NAME := migration-skeleton
BINARY_NAME := bin/$(SERVICE_NAME)
MAIN_PACKAGE := ./cmd/apiserver
SCHEMA ?=
GO := go

# Shell
SHELL := /bin/bash

# Default target
.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: deps
deps: ## Download dependencies
	$(GO) mod download

.PHONY: tidy
tidy: ## Run go mod tidy
	$(GO) mod tidy

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@$(GO) clean

.PHONY: migration-up
migration-up: ## Run database migrations up for all schemas (optional: SCHEMA=my_schema for specific schema)
	@echo "Running migrations up..."
	@if [ -n "$(SCHEMA)" ]; then \
		echo "Target schema: $(SCHEMA)"; \
		$(GO) run $(MAIN_PACKAGE) migrate up --schema=$(SCHEMA) $(N); \
	else \
		echo "Running for all schemas defined in DB_SCHEMAS"; \
		$(GO) run $(MAIN_PACKAGE) migrate up $(N); \
	fi

.PHONY: migration-down
migration-down: ## Run database migrations down for all schemas (optional: SCHEMA=my_schema for specific schema)
	@echo "Running migrations down..."
	@if [ -n "$(SCHEMA)" ]; then \
		echo "Target schema: $(SCHEMA)"; \
		$(GO) run $(MAIN_PACKAGE) migrate down --schema=$(SCHEMA) $(N); \
	else \
		echo "Running for all schemas defined in DB_SCHEMAS"; \
		$(GO) run $(MAIN_PACKAGE) migrate down $(N); \
	fi

.PHONY: migration-force
migration-force: ## Force database migration version for all schemas (usage: make migration-force V=1 SCHEMA=my_schema)
	@echo "Forcing migration version $(V)..."
	@if [ -z "$(V)" ]; then echo "Error: V (version) is required"; exit 1; fi
	@if [ -n "$(SCHEMA)" ]; then \
		echo "Target schema: $(SCHEMA)"; \
		$(GO) run $(MAIN_PACKAGE) migrate force --schema=$(SCHEMA) $(V); \
	else \
		echo "Running for all schemas defined in DB_SCHEMAS"; \
		$(GO) run $(MAIN_PACKAGE) migrate force $(V); \
	fi

.PHONY: migration-version
migration-version: ## Check database migration version for all schemas (optional: SCHEMA=my_schema for specific schema)
	@echo "Checking migration version..."
	@if [ -n "$(SCHEMA)" ]; then \
		echo "Target schema: $(SCHEMA)"; \
		$(GO) run $(MAIN_PACKAGE) migrate version --schema=$(SCHEMA); \
	else \
		echo "Running for all schemas defined in DB_SCHEMAS"; \
		$(GO) run $(MAIN_PACKAGE) migrate version; \
	fi

.PHONY: migration-create
migration-create: ## Create a new migration file (usage: make migration-create NAME=migration_name SCHEMA=public)
	@if [ -z "$(NAME)" ]; then echo "Error: NAME is required"; exit 1; fi
	@if [ -z "$(SCHEMA)" ]; then echo "Error: SCHEMA is required"; exit 1; fi
	@echo "Creating migration files for $(NAME) in schema: $(SCHEMA)..."
	@mkdir -p "cmd/apiserver/app/migrations/$(SCHEMA)"
	@MAX_VERSION=$$(ls cmd/apiserver/app/migrations/$(SCHEMA)/*.up.sql 2>/dev/null | sed 's/.*\///' | sed 's/_.*//' | sort -n | tail -1 2>/dev/null || echo "000000"); \
	NEXT_VERSION=$$(printf "%06d" $$((10#$${MAX_VERSION} + 1))); \
	touch "cmd/apiserver/app/migrations/$(SCHEMA)/$${NEXT_VERSION}_$(NAME).up.sql"; \
	touch "cmd/apiserver/app/migrations/$(SCHEMA)/$${NEXT_VERSION}_$(NAME).down.sql"; \
	echo "Created migration files:"; \
	echo "  - cmd/apiserver/app/migrations/$(SCHEMA)/$${NEXT_VERSION}_$(NAME).up.sql"; \
	echo "  - cmd/apiserver/app/migrations/$(SCHEMA)/$${NEXT_VERSION}_$(NAME).down.sql"
