include .env

LOCAL_MIGRATION_DIR=$(MIGRATION_DIR) 

POSTGRES_DSN="host=localhost port=5432 dbname=$(POSTGRES_DB) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) sslmode=disable"

migration-up:
	goose -dir ${LOCAL_MIGRATION_DIR} postgres ${POSTGRES_DSN} up -v