FROM golang:1.23.3-alpine AS builder

RUN apk --no-cache add ca-certificates gcc g++ libc-dev

WORKDIR /app

COPY go.mod go.sum ./
COPY ./web/index.html ./web/

RUN go mod download

COPY . .

RUN go build -o ./bin/orchestrator ./cmd/orchestrator/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bin/orchestrator /bin/orchestrator
COPY --from=builder /app/.env /.env
COPY --from=builder /app/web/index.html /bin/web/
COPY --from=builder /app/migrations/ /bin/migrations/

EXPOSE 8080

CMD ["/bin/orchestrator"]