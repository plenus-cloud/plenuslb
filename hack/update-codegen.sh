#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

CURRENT_DIR=$(pwd)
REPO_DIR=$CURRENT_DIR

PROJECT_MODULE="plenus.io/plenuslb"
IMAGE_NAME="kubernetes-codegen:latest"

RESOURCE="loadbalancing:v1alpha1"

export GOPATH=${GOPATH-"$HOME/go"}

CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
bash "${CODEGEN_PKG}/generate-groups.sh" all \
  "$PROJECT_MODULE/pkg/client" \
  "$PROJECT_MODULE/pkg/apis" \
  "$RESOURCE" \
  --output-base "$(dirname "${BASH_SOURCE[0]}")/../../../" \
  --go-header-file="${SCRIPT_ROOT}/hack/boilerplate.go.txt"

# bash "${CODEGEN_PKG}/generate-internal-groups.sh" "deepcopy,defaulter,conversion,openapi" \
#   "$PROJECT_MODULE/pkg/client" \
#   "$PROJECT_MODULE/pkg/apis" \
#   "$PROJECT_MODULE/pkg/apis" \
#   "$IPALLOCATION_CUSTOM_RESOURCE $EPHEMERALPPOOL_CUSTOM_RESOURCE $PERSISTENTIPPOOL_CUSTOM_RESOURCE" \
#   --output-base "$(dirname "${BASH_SOURCE[0]}")/../../../" \
#   --go-header-file="${SCRIPT_ROOT}/hack/boilerplate.go.txt"

# To use your own boilerplate text append:
#   --go-header-file "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt"