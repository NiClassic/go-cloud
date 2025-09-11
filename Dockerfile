FROM golang:1.25 AS builder

WORKDIR /app

COPY ./backend/go.mod ./backend/go.sum ./

RUN go mod download

COPY ./backend/ ./

RUN apt-get update && apt-get install -y sqlite3

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main.go

FROM alpine:latest

RUN addgroup -S appuser && adduser -S appuser -G appuser

WORKDIR /app

RUN mkdir -p /data && chown appuser:appuser /data

# Copy binary
COPY --from=builder /app/server .

# Copy html templates
COPY --from=builder /app/templates ./templates
# Copy static css
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD [ "./server" ]