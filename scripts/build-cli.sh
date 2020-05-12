#!/usr/bin/env bash

set -euo pipefail

# TODO: Uncomment this
# oses="linux darwin windows"
oses="darwin"
archs="amd64"

for os in $oses; do
  for arch in $archs; do 
    echo "building $os-$arch"
    GOOS=$os GOARCH=$arch CGO_ENABLED=0 \
      go build \
      -ldflags "-w -X github.com/krancour/brignext/v2/internal/common/version.version=$VERSION -X github.com/krancour/brignext/v2/internal/common/version.commit=$COMMIT" \
      -o ./bin/brignext-$os-$arch \
      ./internal/cli
  done
  if [ $os = 'windows' ]; then
    mv ./bin/brignext-$os-$arch ./bin/brignext-$os-$arch.exe
  fi
done
