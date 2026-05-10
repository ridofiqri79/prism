FROM golang:1.26.2 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

ARG SQLC_VERSION=v1.31.1
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION}

COPY . .

RUN sqlc generate

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o bin/prism ./cmd/api

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates tzdata && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/bin/prism .
EXPOSE 8080
CMD ["./prism"]
