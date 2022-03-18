default: all

test:
	go test ./...

clean:
	rm -r dist || true

build-linux: clean build-linux-honeypot build-linux-punchr build-linux-api

build-linux-honeypot:
	GOOS=linux GOARCH=amd64 go build -o dist/honeypot cmd/honeypot/*

build-linux-punchr:
	GOOS=linux GOARCH=amd64 go build -o dist/punchr cmd/punchr/*

build-linux-api:
	GOOS=linux GOARCH=amd64 go build -o dist/punchrapi cmd/api/*

build: clean build-honeypot build-punchr build-api

build-honeypot:
	go build -o dist/honeypot cmd/honeypot/*

build-punchr:
	go build -o dist/punchr cmd/punchr/*

build-api:
	go build -o dist/punchrapi cmd/api/*

format:
	gofumpt -w -l .

tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.14.1
	go install github.com/volatiletech/sqlboiler/v4@v4.6.0
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.6.0
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto:
	protoc --proto_path=. --go_out=pkg/pb --go_opt=paths=source_relative --go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative api.proto

models:
	sqlboiler psql

database:
	docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_USER=punchr -e POSTGRES_DB=punchr postgres:13

database-reset: migrate-down migrate-up models

migrate-up:
	migrate -database 'postgres://punchr:password@localhost:5432/punchr?sslmode=disable' -path migrations up

migrate-down:
	migrate -database 'postgres://punchr:password@localhost:5432/punchr?sslmode=disable' -path migrations down

