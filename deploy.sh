#!/usr/bin/env bash

image="krancour/brignext-apiserver:latest"

docker build -f docker/apiserver/Dockerfile . -t $image

docker push $image

runningPod=$(kubectl get pods -n brigade | grep apiserver | awk '{print $1}')

kubectl delete pod $runningPod -n brigade
