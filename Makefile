.PHONY: deps
deps:
	go mod vendor

.PHONY: golangci
golangci:
	golangci-lint run -v

.PHONY: run-server
run-server:
	go run cmd/server/main.go

.PHONY: run-client
run-client:
	go run cmd/client/main.go

.PHONY: test
test:
	go test ./... -v

.PHONY: start
start:
	docker-compose up --build --abort-on-container-exit

.PHONY: stop
stop:
	docker-compose down