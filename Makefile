vendor:
	go mod vendor

run:
	go run -mod=vendor main.go

build:
	go build -mod=vendor

PHONY: run build vendor