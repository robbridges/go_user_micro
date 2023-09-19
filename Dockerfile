FROM golang:1.20.0

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o main ./cmd/api/

EXPOSE 8080

CMD ["./main"]