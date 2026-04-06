DB_BASE=postgres://postgres:postgres@localhost:5432/video_platform?sslmode=disable

USER_MIGRATIONS=services/user-service/db/migrations
CONTENT_MIGRATIONS=services/content-service/db/migrations
INTERACTION_MIGRATIONS=services/interaction-service/db/migrations
UPLOAD_MIGRATIONS=services/upload-service/db/migrations

USER_DB_URL=$(DB_BASE)&x-migrations-table=user_service_schema_migrations
CONTENT_DB_URL=$(DB_BASE)&x-migrations-table=content_service_schema_migrations
INTERACTION_DB_URL=$(DB_BASE)&x-migrations-table=interaction_service_schema_migrations
UPLOAD_DB_URL=$(DB_BASE)&x-migrations-table=upload_service_schema_migrations

SERVICES ?= user-service content-service upload-service interaction-service

migrate-user-create:
	migrate create -ext sql -dir $(USER_MIGRATIONS) -seq $(name)

migrate-user-up:
	migrate -path $(USER_MIGRATIONS) -database "$(USER_DB_URL)" up

migrate-user-down:
	migrate -path $(USER_MIGRATIONS) -database "$(USER_DB_URL)" down 1

migrate-user-version:
	migrate -path $(USER_MIGRATIONS) -database "$(USER_DB_URL)" version

migrate-content-create:
	migrate create -ext sql -dir $(CONTENT_MIGRATIONS) -seq $(name)

migrate-content-up:
	migrate -path $(CONTENT_MIGRATIONS) -database "$(CONTENT_DB_URL)" up

migrate-content-down:
	migrate -path $(CONTENT_MIGRATIONS) -database "$(CONTENT_DB_URL)" down 1

migrate-content-version:
	migrate -path $(CONTENT_MIGRATIONS) -database "$(CONTENT_DB_URL)" version

migrate-interaction-create:
	migrate create -ext sql -dir $(INTERACTION_MIGRATIONS) -seq $(name)

migrate-interaction-up:
	migrate -path $(INTERACTION_MIGRATIONS) -database "$(INTERACTION_DB_URL)" up

migrate-interaction-down:
	migrate -path $(INTERACTION_MIGRATIONS) -database "$(INTERACTION_DB_URL)" down 1

migrate-interaction-version:
	migrate -path $(INTERACTION_MIGRATIONS) -database "$(INTERACTION_DB_URL)" version

migrate-upload-create:
	migrate create -ext sql -dir $(UPLOAD_MIGRATIONS) -seq $(name)

migrate-upload-up:
	migrate -path $(UPLOAD_MIGRATIONS) -database "$(UPLOAD_DB_URL)" up

migrate-upload-down:
	migrate -path $(UPLOAD_MIGRATIONS) -database "$(UPLOAD_DB_URL)" down 1

migrate-upload-version:
	migrate -path $(UPLOAD_MIGRATIONS) -database "$(UPLOAD_DB_URL)" version

compose-up:
	docker compose up --build

compose-down:
	docker compose down -v

compose-logs:
	docker compose logs -f

build-user:
	docker compose build user-service

build-content:
	docker compose build content-service

build-interaction:
	docker compose build interaction-service

build-upload:
	docker compose build upload-service

sqlc:
	@for service in $(SERVICES); do \
		echo "Generating sqlc for $$service..."; \
		(cd services/$$service && sqlc generate) || exit 1; \
	done