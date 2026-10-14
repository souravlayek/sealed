# Makefile for sealed-server

.PHONY: build start sqlc migrateup migratedown install-tools gen-docs


# Build the project
build:
	go build -o bin/server ./cmd/server

# Start the project
start:
	 air

# Install required development tools
install-tools:
	@echo "Installing required development tools..."

	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)


	@which sqlc > /dev/null || (echo "Installing sqlc..." && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest)
	@which migrate > /dev/null || (echo "Installing migrate..." && echo "Please install migrate manually using: curl -L https://github.com/golang-migrate/migrate/releases/download/\$$version/migrate.\$$os-\$$arch.tar.gz | tar xvz")

	@echo "All required tools are installed!"



# Database connection string
db := postgresql://postgres:aQ6H0kYZd+sBBgx22lLvZwS6AwzRRVqn5M05ArBX+6I=@100.91.128.74:5432/sealed?sslmode=disable

# generate ORM
sqlc:
	sqlc generate

# Run database migrations up
migrateup:
	migrate -path ./internal/database/migration -database "$(db)" -verbose up

# Run database migrations down
migratedown:
	migrate -path ./internal/database/migration -database "$(db)" -verbose down

# Generate Swagger documentation
# -d points at directories that contain .go files with swagger annotations, so
# swag can resolve their package names (the project root has no .go files).
gen-docs:
	swag init -g main.go -d ./cmd/server,./internal/api/handlers,./pkg/sqlc --parseDependency
	swag fmt
