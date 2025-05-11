FROM golang:1.23.3-alpine AS builder

RUN apk --no-cache add ca-certificates gcc g++ libc-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ./bin/agent ./cmd/agent/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bin/agent /bin/agent
COPY --from=builder /app/.env /.env

CMD ["/bin/agent"]