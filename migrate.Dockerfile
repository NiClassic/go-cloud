FROM golang:1.25 AS migrate_builder

WORKDIR /app

RUN apt-get update && apt-get install -y \
    libsqlite3-dev \
    gcc \
    && rm -rf /var/lib/apt/lists/*

RUN CGO_ENABLED=1 go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM debian:stable-slim

RUN apt-get update && apt-get install -y \
    libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=migrate_builder /go/bin/migrate /usr/local/bin/migrate

ENTRYPOINT ["/usr/local/bin/migrate"]