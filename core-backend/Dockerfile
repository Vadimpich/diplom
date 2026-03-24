FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY cmd ./cmd
COPY config ./config
COPY db ./db
COPY internal ./internal
COPY migrations ./migrations

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod -o /bin/api ./cmd/api

FROM alpine:3.20

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /bin/api /app/api
COPY --from=builder /app/migrations /app/migrations
RUN chown -R app:app /app

USER app

EXPOSE 8080

ENTRYPOINT ["/app/api"]
