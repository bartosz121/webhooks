FROM golang:alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ./bin/api ./cmd/api

CMD ["/app/bin/api"]

EXPOSE 8080
