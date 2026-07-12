# Builds the three backend binaries (api, worker, migrate) from the single
# root Go module into one image; CMD defaults to the API server, but the
# same image serves the worker and migrate jobs by overriding the command
# (see docker-compose.yml / k8s manifests).
FROM golang:1.26-alpine AS builder
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -o bin/api ./cmd/api
RUN go build -o bin/worker ./cmd/worker
RUN go build -o bin/migrate ./cmd/migrate

FROM alpine
ARG APP_VERSION
LABEL org.opencontainers.image.version=${APP_VERSION}
RUN apk add --no-cache ca-certificates tzdata && update-ca-certificates
RUN adduser -S -D -u 1000 appuser

WORKDIR /app
COPY --from=builder /app/bin/api ./bin/
COPY --from=builder /app/bin/worker ./bin/
COPY --from=builder /app/bin/migrate ./bin/
COPY --from=builder /app/db/migrations ./db/migrations

USER appuser
EXPOSE 8181

CMD ["./bin/api"]
