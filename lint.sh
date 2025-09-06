# -v $(go env GOCACHE):/.cache/go-build -e GOCACHE=/.cache/go-build \
# -v $(go env GOMODCACHE):/.cache/mod -e GOMODCACHE=/.cache/mod \
# --user $(id -u):$(id -g) \
docker run --rm -t -v $(pwd):/app -w /app \
-v ./.cache/go-build:/.cache/go-build -e GOCACHE=/.cache/go-build \
-v ./.cache/mod:/.cache/mod -e GOMODCACHE=/.cache/mod \
-v ./.cache/golangci-lint:/.cache/golangci-lint -e GOLANGCI_LINT_CACHE=/.cache/golangci-lint \
golangci/golangci-lint:v2.4.0 golangci-lint run
