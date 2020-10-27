#!/bin/bash

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

cd $SCRIPT_ROOT/pkg/proto/v1alpha1 && ls -lha && ./.protogen.sh