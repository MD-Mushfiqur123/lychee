#!/bin/sh

set -eu

export VERSION=${VERSION:-0.0.0}
export GOFLAGS="'-ldflags=-w -s \"-X=github.com/MD-Mushfiqur123/lychee/version.Version=$VERSION\" \"-X=github.com/MD-Mushfiqur123/lychee/server.mode=release\"'"

docker build \
    --push \
    --platform=linux/arm64,linux/amd64 \
    --build-arg=GOFLAGS \
    -f Dockerfile \
    -t lychee/lychee -t lychee/lychee:$VERSION \
    .
