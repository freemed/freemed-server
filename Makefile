VERSION := `date +%FT%T%z`
GOROOT  := /opt/go
BINARY  := freemed

all: clean deps binary

binary:
	@echo "- Building binary version ${VERSION}"
	( cd cmd/freemed-server ; go build -ldflags "-X main.Version=${VERSION}" -v )

deps:
	@echo "- Refreshing dependencies"
	( cd cmd/freemed-server ; go get -v -d ./... )

sqlc:
	sqlc generate -f internal/db/sqlc.yaml
.PHONY: sqlc

generate: sqlc
.PHONY: generate

clean:
	@echo "- Cleaning old build files"
	( cd cmd/freemed-server ; go clean -v )

# Database migrations (requires golang-migrate CLI or go tool)
migrate-install:
	go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
.PHONY: migrate-install

migrate-up:
	migrate -path internal/db/migrations \
	  -database "mysql://${DB_USER}:${DB_PASS}@tcp(${DB_HOST}:3306)/${DB_NAME}" up
.PHONY: migrate-up

migrate-down:
	migrate -path internal/db/migrations \
	  -database "mysql://${DB_USER}:${DB_PASS}@tcp(${DB_HOST}:3306)/${DB_NAME}" down 1
.PHONY: migrate-down

crosscompile:
	( cd cmd/freemed-server ; \
		GOROOT=${GOROOT} CGO_ENABLED=0 GOOS=linux GOARCH=386 \
		go build -v -ldflags "-X main.Version=${VERSION}" \
			-o ${BINARY}.linux.x86 )
	( cd cmd/freemed-server ; \
		GOROOT=${GOROOT} CGO_ENABLED=0 GOOS=windows GOARCH=386 \
		go build -v -ldflags "-X main.Version=${VERSION}" \
			-o ${BINARY}.x86.exe )
	( cd cmd/freemed-server ; \
		GOROOT=${GOROOT} CGO_ENABLED=0 GOOS=darwin GOARCH=386 \
		go build -v -ldflags "-X main.Version=${VERSION}" \
			-o ${BINARY}.mac.bin )

# === Frontend (SvelteKit) ===

frontend-deps:
	@echo "- Installing frontend dependencies"
	( cd frontend ; npm ci )
.PHONY: frontend-deps

frontend-dev:
	@echo "- Starting SvelteKit dev server (proxies /api and /auth to :3000)"
	( cd frontend ; npm run dev )
.PHONY: frontend-dev

frontend-build:
	@echo "- Building SvelteKit frontend for production"
	( cd frontend ; npm run build )
.PHONY: frontend-build

frontend-check:
	@echo "- Running svelte-check"
	( cd frontend ; npx svelte-check )
.PHONY: frontend-check

frontend-clean:
	@echo "- Cleaning frontend build output"
	rm -rf frontend/build frontend/.svelte-kit
.PHONY: frontend-clean

# === Docker ===

docker-build:
	docker build -t freemed-server .
	docker build -t freemed-frontend frontend/
.PHONY: docker-build
