FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/main ./cmd/main.go

FROM alpine:3.18
WORKDIR /app

COPY --from=builder /app/bin/main .
COPY .env . 

EXPOSE 8080

CMD ["./main"]