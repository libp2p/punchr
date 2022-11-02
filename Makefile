VERSION=0.5.0

default: all

test:
	go test ./...

clean:
	rm -r dist || true
	mkdir dist

build-linux: clean build-linux-honeypot build-linux-client build-linux-server

build-linux-honeypot:
	GOOS=linux GOARCH=amd64 go build -o dist/punchrhoneypot cmd/honeypot/*

build-linux-client:
	GOOS=linux GOARCH=amd64 go build -o dist/punchrclient cmd/client/*

build-raspi-client:
	GOOS=linux GOARCH=arm GOARM=7 go build -o dist/punchrclient cmd/client/*

build-linux-server:
	GOOS=linux GOARCH=amd64 go build -o dist/punchrserver cmd/server/*

build: clean build-honeypot build-client build-server

build-honeypot:
	go build -o dist/punchrhoneypot cmd/honeypot/*

build-client:
	go build -o dist/punchrclient cmd/client/*

build-server:
	go build -o dist/punchrserver cmd/server/*

format:
	gofumpt -w -l .

tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.14.1
	go install github.com/volatiletech/sqlboiler/v4@v4.6.0
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.6.0
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install fyne.io/fyne/v2/cmd/fyne@latest
	go install github.com/fyne-io/fyne-cross@latest

proto:
	protoc --proto_path=. --go_out=pkg/pb --go_opt=paths=source_relative --go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative punchr.proto

models:
	sqlboiler psql

database:
	docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_USER=punchr -e POSTGRES_DB=punchr postgres:14

database-reset: migrate-down migrate-up models

migrate-up:
	migrate -database 'postgres://punchr:password@localhost:5432/punchr?sslmode=disable' -path pkg/db/migrations up

migrate-down:
	migrate -database 'postgres://punchr:password@localhost:5432/punchr?sslmode=disable' -path pkg/db/migrations down

package: package-darwin package-linux

package-darwin:
	fyne-cross darwin \
		-app-id=ai.protocol.punchr \
		-app-version=${VERSION} \
		-arch=arm64,amd64 \
		-icon=`pwd`/gui/client/glove-active.png \
		-name=Punchr

package-linux:
	fyne-cross linux \
		-app-id=ai.protocol.punchr \
		-app-version=${VERSION} \
		-arch=* \
		-icon=`pwd`/gui/client/glove-active.png \
		-name=Punchr \
		-release

sign: package-darwin
	codesign \
		--force \
		--options runtime \
		--deep \
		--sign "${SIGNING_CERTIFICATE}" \
		-i "ai.protocol.punchr" \
		./fyne-cross/dist/darwin-amd64/Punchr.app

	codesign \
		--force \
		--options runtime \
		--deep \
		--sign "${SIGNING_CERTIFICATE}" \
		-i "ai.protocol.punchr" \
		./fyne-cross/dist/darwin-arm64/Punchr.app

dmg: sign
	create-dmg \
		--volname Punchr \
		--volicon ./gui/client/glove-active.png \
		--window-pos 200 120 \
		--window-size 800 400 \
		--icon-size 100 \
		--icon ./dist/Punchr.app 200 190 \
		--app-drop-link 600 185 \
		--codesign "${SIGNING_CERTIFICATE}" \
		--notarize "${NOTARIZATION_PROFILE}" \
		./fyne-cross/dist/darwin-amd64/Punchr.dmg \
		./fyne-cross/dist/darwin-amd64/Punchr.app
		
	create-dmg \
		--volname Punchr \
		--volicon ./gui/client/glove-active.png \
		--window-pos 200 120 \
		--window-size 800 400 \
		--icon-size 100 \
		--icon ./dist/Punchr.app 200 190 \
		--app-drop-link 600 185 \
		--codesign "${SIGNING_CERTIFICATE}" \
		--notarize "${NOTARIZATION_PROFILE}" \
		./fyne-cross/dist/darwin-arm64/Punchr.dmg \
		./fyne-cross/dist/darwin-arm64/Punchr.app
