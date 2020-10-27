#!/bin/bash

set -e

eval $(minikube docker-env)

. ./minikube.envs.sh

./hack/update-codegen.sh 

docker build --rm -t $OPERATOR_IMAGE_REPO:latest -f operator.Dockerfile .