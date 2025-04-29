#!/bin/bash

# Установка protoc и плагинов
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@latest

# Генерация Go файлов
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    --go-pulsar_out=. --go-pulsar_opt=paths=source_relative \
    zenrock/validation/query.proto 