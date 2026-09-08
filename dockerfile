FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /roomly-api ./cmd/api

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser

WORKDIR /app

COPY --from=builder /roomly-api /app/roomly-api

USER appuser

EXPOSE 8080

CMD ["/app/roomly-api"]