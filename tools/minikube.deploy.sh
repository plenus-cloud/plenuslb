#/bin/bash

set -e

. ./minikube.envs.sh

NAMESPACE=default
TAG="latest"
while [ "$1" != "" ]; do
    case $1 in
    -v | --version)
        shift
        TAG=$1
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

kubectl version
# kubectl config use-context gke_nephosgame_europe-west4-c_nephos-dev
kubectl config use-context minikube
helm version
rm -fr _deploy/charts-template-out
mkdir -p _deploy/charts-template-out
helm template ./_deploy/plenuslb \
    --namespace=$NAMESPACE \
    --set image.repository=$CONTROLLER_IMAGE_REPO \
    --set image.tag=$TAG \
    --set image.pullPolicy=IfNotPresent \
    --set forceDeploy=true \
    --output-dir=./_deploy/charts-template-out
helm lint ./_deploy/plenuslb
helm upgrade \
    --install \
    --namespace=$NAMESPACE \
    --set image.repository=$CONTROLLER_IMAGE_REPO \
    --set image.tag=$TAG \
    --set image.pullPolicy=IfNotPresent \
    --set forceDeploy=true \
    --wait \
    --atomic \
    --version=$TAG \
    plenuslb \
    ./_deploy/plenuslb




