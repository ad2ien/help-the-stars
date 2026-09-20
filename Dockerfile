FROM golang:1.26-alpine AS builder

WORKDIR /app
RUN apk add --no-cache \
gcc \
musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o help-the-stars main.go

FROM alpine:3.24
WORKDIR /app
RUN addgroup -S app && adduser -S -G app app \
 && mkdir -p /app/db \
 && chown -R app:app /app

COPY --from=builder /app/help-the-stars .
CMD ["./help-the-stars"]