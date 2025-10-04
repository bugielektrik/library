# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/library-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/library-worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/library-migrate ./cmd/migrate

# API server image
FROM alpine:3.22 AS api
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/bin/library-api ./
COPY --from=builder /build/migrations ./migrations
EXPOSE 8080
ENTRYPOINT ["./library-api"]

# Worker image
FROM alpine:3.22 AS worker
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/bin/library-worker ./
ENTRYPOINT ["./library-worker"]

# Migration image
FROM alpine:3.22 AS migrate
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/bin/library-migrate ./
COPY --from=builder /build/migrations ./migrations
ENTRYPOINT ["./library-migrate"]
