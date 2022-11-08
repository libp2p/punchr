#!/bin/bash
oss=(darwin linux windows freebsd netbsd openbsd)
archs=(amd64 arm64 386 arm)

for os in "${oss[@]}"; do
  for arch in "${archs[@]}"; do
    echo "Compiling $os $arch..."
    GOOS=${os} GOARCH=${arch} go build -ldflags="-X 'pkg/client/client.Version=$(cat cmd/client/version)'" -o dist/punchr_cli_${os}_${arch} cmd/client/*.go
  done
done