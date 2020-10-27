#/bin/bash

set -e

# Directory to write generated code to (.go files)
GO_OUT_DIR="./generated"

echo "[INFO] Cleaning generation directory"
rm -fr ./generated

echo "[INFO] Creating directory $GO_OUT_DIR"
mkdir -p $GO_OUT_DIR

echo "[INFO] Generating go files from .proto"
docker run --rm -v $(pwd):$(pwd) -w $(pwd) grpc/go:1.0 protoc -I ./ ./plenuslb.proto --go_out=plugins=grpc:${GO_OUT_DIR}
