#!/usr/bin/env bash

set -euo pipefail

image="krancour/brignext-worker:latest"

cd worker/

yarn install

cd ..

docker build -f worker/Dockerfile . -t $image

docker push $image
