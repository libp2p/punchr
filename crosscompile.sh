#!/bin/bash
oss=(darwin linux windows freebsd netbsd openbsd)
archs=(amd64 arm64 386)

for os in "${oss[@]}"; do
  for arch in "${archs[@]}"; do
    echo "Compiling $os $arch..."
    GOOS=${os} GOARCH=${arch} go build -ldflags="-X 'github.com/dennis-tra/punchr/pkg/client.Version=$(cat cmd/client/version)'" -o dist/punchr_cli_${os}_${arch} cmd/client/*.go
  done
done

arms=(5 6 7)
for arm in "${arms[@]}"; do
    echo "Compiling linux arm$arm..."
    GOOS=linux GOARCH=arm GOARM=${arm} go build -ldflags="-X 'github.com/dennis-tra/punchr/pkg/client.Version=$(cat cmd/client/version)'" -o dist/punchr_cli_linux_arm${arm} cmd/client/*.go
done