
GO := "docker run -it --init --rm -v $(pwd):/app -w /app -p 1983:1983 -v go-cache:/go/pkg/mod -v go-build-cache:/root/.cache/go-build golang:1.25.11-bookworm go"
GOLINT := "docker run -it --init --rm -v $(pwd):/app -w /app -p 1983:1983 -v go-cache:/go/pkg/mod -v go-build-cache:/root/.cache/go-build golangci/golangci-lint"
SQLC := "docker run --rm -v $PWD:/src -w /src sqlc/sqlc"
MIGRATE := "docker run --rm -v $(pwd)/migrations:/migrations --network host migrate/migrate"

# sqlc generate
sqlcgen:
  {{ SQLC }} generate

# sqlc contained command
sqlc:
  {{ SQLC }}

# run the app
run:
  @source $(pwd)/.env && \
  {{GO}} run main.go --debug \
  --gh-token $GITHUB_TOKEN \
  --db-file $DB_FILE \
  --labels "${LABELS}" \
  --matrix-server $MATRIX_HOMESERVER \
  --matrix-user $MATRIX_USERNAME \
  --matrix-password $MATRIX_PASSWORD \
  --matrix-room $MATRIX_ROOM_ID 

# run go tests
test:
  {{GO}} test ./...

vet:
  {{GO}} vet ./...

# run golangci-lint
lint:
  {{GOLINT}} golangci-lint run

# go mod tidy
tidy:
  {{GO}} mod tidy

# upgrade go dependencies
upgrade:
  {{GO}} get -u ./...

# test vet lint
check: test vet lint

# create migration
new-migration name:
  {{ MIGRATE }} -path=/migrations -database "sqlite://db/help-the-stars.db" create -ext sql -dir /migrations -seq {{ name }}
