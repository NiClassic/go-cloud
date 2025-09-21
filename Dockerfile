FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

FROM debian:bookworm-slim

RUN useradd -u 1000 -m clouduser \
    && mkdir -p /data \
    && chown -R clouduser:clouduser /data

COPY --chown=clouduser:clouduser --from=builder /app/db/migrations/ ./db/migrations/

COPY --chown=clouduser:clouduser --from=builder /app/server .

COPY --chown=clouduser:clouduser --from=builder /app/templates ./templates

COPY --chown=clouduser:clouduser --from=builder /app/static ./static

COPY --chown=clouduser:clouduser --from=builder /app/.env .env

EXPOSE 8080

CMD [ "./server" ]
