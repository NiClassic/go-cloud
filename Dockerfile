FROM golang:1.25

WORKDIR /app

COPY ./backend/go.mod ./backend/go.sum ./

RUN go mod download

COPY ./backend/ ./

RUN apt-get update && apt-get install -y sqlite3

RUN CGO_ENABLED=1 GOOS=linux go build -o server ./main.go

EXPOSE 8080

CMD [ "./server" ]