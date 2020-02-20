#!/usr/bin/env bash

set -euo pipefail

image="krancour/brignext-controller:latest"

docker build -f controller/Dockerfile . -t $image

docker push $image

runningPod=$(kubectl get pods -n brigade | grep controller | awk '{print $1}')

kubectl delete pod $runningPod -n brigade
