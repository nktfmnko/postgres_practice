include ../../.env
export

ROOT_DIR := $(CURDIR)
GOOSE_VERSION := v3.29.0
BIN_DIR := $(ROOT_DIR)\bin
GOOSE := $(BIN_DIR)\goose.exe

.PHONY: goose-install migrate-up

goose-install:
	@echo Installing goose...
	@if not exist "$(BIN_DIR)" mkdir "$(BIN_DIR)"
	set "GOBIN=$(BIN_DIR)" && go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)
	@if not exist "$(GOOSE)" exit 1

migrate-up: goose-install
	@"$(GOOSE)" -dir migrations postgres "$(PG_DSN)" up

migrate-down: goose-install
	@"$(GOOSE)" -dir migrations postgres "$(PG_DSN)" down

migrate-status: goose-install
	@"$(GOOSE)" -dir migrations postgres "$(PG_DSN)" status

migrate-create: goose-install
	@"$(GOOSE)" -dir migrations create "$(NAME)" sql