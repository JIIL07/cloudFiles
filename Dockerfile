FROM golang:latest

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server cmd/server/example.go

RUN mkdir -p /app/secrets
COPY secrets/.env /app/secrets/.env
ENV ENV_PATH=/app/secrets/.env

RUN mkdir -p /app/config
COPY config/config.yaml /app/config/config.yaml
ENV CONFIG_PATH=/app/config/config.yaml
COPY config/client.yaml /app/config/client.yaml
ENV CLIENT_CONFIG_PATH=/app/config/client.yaml

RUN mkdir -p /app/storage
COPY storage/storage.db /app/storage/storage.db
ENV DATABASE_PATH=/app/storage/storage.db

EXPOSE 8080

CMD ["./server"]