FROM golang:1.24-alpine AS builder

WORKDIR /app
RUN apk add --no-cache \
gcc \
musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o help-the-stars main.go

FROM alpine:3.22.1
WORKDIR /app
COPY --from=builder /app/help-the-stars .
CMD ["./help-the-stars"]