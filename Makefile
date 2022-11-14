VERSION_CLIENT=`cat cmd/client/version`
VERSION_CLIENT_GUI=`cat gui/client/version`
VERSION_HONEYPOT=`cat cmd/honeypot/version`
VERSION_SERVER=`cat cmd/server/version`

test:
	go test ./...

clean:
	rm -r dist || true
	rm -r fyne-cross || true
	mkdir dist

build: clean build-honeypot build-client build-server

build-honeypot:
	go build -ldflags="-X 'main.Version=$(VERSION_HONEYPOT)'" -o dist/punchrhoneypot cmd/honeypot/*.go

build-client:
	go build -ldflags="-X 'main.Version=$(VERSION_CLIENT)'" -o dist/punchrclient cmd/client/*.go

build-server:
	go build -ldflags="-X 'main.Version=$(VERSION_SERVER)'" -o dist/punchrserver cmd/server/*.go

docker-client:
	docker build -t dennistra/punchr-client:latest -t dennistra/punchr-client:$(VERSION_CLIENT) -f docker/Dockerfile_client .

docker-server:
	docker build -t dennistra/punchr-server:latest -t dennistra/punchr-server:$(VERSION_SERVER) -f docker/Dockerfile_server .

docker-honeypot:
	docker build -t dennistra/punchr-honeypot:latest -t dennistra/punchr-honeypot:$(VERSION_HONEYPOT) -f docker/Dockerfile_honeypot .

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

migrate-up:
	migrate -database 'postgres://punchr:password@localhost:5432/punchr?sslmode=disable' -path pkg/db/migrations up

migrate-down:
	migrate -database 'postgres://punchr:password@localhost:5432/punchr?sslmode=disable' -path pkg/db/migrations down

database-reset: migrate-down migrate-up models

gui: clean package-linux package-freebsd dmg

package-darwin:
	fyne-cross darwin \
		-app-version=$(VERSION_CLIENT_GUI) \
		-arch=arm64,amd64 \
		-name=Punchr

	# fyne-cross doesn't really pick up the right architectures. So we'll compile it again
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'github.com/dennis-tra/punchr/gui/client.Version=$(VERSION_CLIENT_GUI)' -X 'github.com/dennis-tra/punchr/pkg/client.Version=$(VERSION_CLIENT)'" -o fyne-cross/dist/darwin-amd64/Punchr.app/Contents/MacOS/punchr fyne.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-X 'github.com/dennis-tra/punchr/gui/client.Version=$(VERSION_CLIENT_GUI)' -X 'github.com/dennis-tra/punchr/pkg/client.Version=$(VERSION_CLIENT)'" -o fyne-cross/dist/darwin-arm64/Punchr.app/Contents/MacOS/punchr fyne.go

package-linux:
	fyne-cross linux \
		-app-version=$(VERSION_CLIENT_GUI) \
		-arch=amd64,386 \
		-ldflags="-X 'github.com/dennis-tra/punchr/gui/client.Version=$(VERSION_CLIENT_GUI)' -X 'github.com/dennis-tra/punchr/pkg/client.Version=$(VERSION_CLIENT)'" \
		-name=Punchr \
		-release

	mv fyne-cross/dist/linux-amd64/Punchr.tar.xz dist/punchr-gui-linux-amd64.tar.xz
	mv fyne-cross/dist/linux-386/Punchr.tar.xz dist/punchr-gui-linux-386.tar.xz

package-freebsd:
	fyne-cross freebsd \
		-app-version=$(VERSION_CLIENT_GUI) \
		-arch=amd64,arm64 \
		-ldflags="-X 'github.com/dennis-tra/punchr/gui/client.Version=$(VERSION_CLIENT_GUI)' -X 'github.com/dennis-tra/punchr/pkg/client.Version=$(VERSION_CLIENT)'" \
		-name=Punchr \
		-release

	mv fyne-cross/dist/freebsd-amd64/Punchr.tar.xz dist/punchr-gui-freebsd-amd64.tar.xz
	mv fyne-cross/dist/freebsd-arm64/Punchr.tar.xz dist/punchr-gui-freebsd-arm64.tar.xz

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
		./dist/punchr-gui-darwin-amd64.dmg \
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
		./dist/punchr-gui-darwin-arm64.dmg \
		./fyne-cross/dist/darwin-arm64/Punchr.app
