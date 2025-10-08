FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
RUN go build -o /bin/transaction_manager ./cmd/transaction_manager

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/transaction_manager /usr/local/bin/transaction_manager

# Directorios para montar CSV / templates
RUN mkdir -p /data /templates
# usuario no root (opcional)
RUN adduser -D appuser
USER appuser

ENTRYPOINT ["/usr/local/bin/transaction_manager"]
# sin CMD: los flags los pasamos desde docker compose run
