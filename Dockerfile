FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY .env .
RUN go build -o myapp ./cmd

EXPOSE 8080

CMD ["./myapp"]