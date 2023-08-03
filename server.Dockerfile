FROM golang:1.20 AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/server

FROM scratch

COPY --from=builder /app/main /
COPY --from=builder /app/.env ./.env

EXPOSE 3333

ENTRYPOINT ["/main"]
