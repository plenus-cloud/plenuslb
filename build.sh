#!/bin/bash

set -e

VERSION=latest
PUSH=false
CONTROLLER_ONLY=false
OPERATOR_ONLY=false

#source ./.envs.sh
OPERATOR_REPO=plenus/plenuslb-operator
CONTROLLER_REPO=plenus/plenuslb-controller

usage() {
    echo "usage: ./build.sh [[[-v version ] [-p push]] | [-h]]"
}

while [ "$1" != "" ]; do
    case $1 in
    -v | --version)
        shift
        VERSION=$1
        ;;
    -p | --push)
        PUSH=true
        ;;
    -oo | --operator-only)
        OPERATOR_ONLY=true
        ;;
    -co | --controller-only)
        CONTROLLER_ONLY=true
        ;;
    -h | --help)
        usage
        exit
        ;;
    *)
        usage
        exit 1
        ;;
    esac
    shift
done

echo "Building version ${VERSION}"
if [ ${CONTROLLER_ONLY} == true ]; then
    echo "Excluding operator due user option"
fi
if [ ${OPERATOR_ONLY} == true ]; then
    echo "Excluding controller due user option"
fi
echo "Push image: ${PUSH}"

OPERATOR_TAG=${OPERATOR_REPO}:${VERSION}
CONTROLLER_TAG=${CONTROLLER_REPO}:${VERSION}

if [ ${CONTROLLER_ONLY} == false ]; then
    docker build -t ${OPERATOR_TAG} -f operator.Dockerfile .
fi
if [ ${OPERATOR_ONLY} == false ]; then
    docker build -t ${CONTROLLER_TAG} .
fi

if [ ${PUSH} == true ]; then
    echo "Pushing images..."
    if [ ${CONTROLLER_ONLY} == false ]; then
        docker push ${OPERATOR_TAG}
    fi
    if [ ${OPERATOR_ONLY} == false ]; then
        docker push ${CONTROLLER_TAG}
    fi
fi
