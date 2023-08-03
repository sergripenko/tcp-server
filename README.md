# TCP-server and client with protection from DDOS attacks

## 1. Getting started
### 1.1 Requirements
+ [Go 1.20+](https://go.dev/dl/) to run tests and server, client without Docker
+ [Docker](https://docs.docker.com/engine/install/) to run with Docker

### 1.2 Install dependencies (run without Docker):
```
make deps
```

### 1.3 Start server without Docker:
```
make run-server
```

### 1.4 Start client without Docker:
```
make run-client
```

### 1.5 Start server and client with docker-compose:
```
make start
```

### 1.6 Run tests:
```
make test
```

### 1.7 Run linters:
```
make golangci
```

## 2. Task description
Design and implement “Word of Wisdom” tcp server.
+ TCP server should be protected from DDOS attacks with the Prof of Work (https://en.wikipedia.org/wiki/Proof_of_work), the challenge-response protocol should be used.
+ The choice of the POW algorithm should be explained.
+ After Prof Of Work verification, server should send one of the quotes from “word of wisdom” book or any other collection of the quotes.
+ Docker file should be provided both for the server and for the client that solves the POW challenge