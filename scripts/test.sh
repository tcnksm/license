#!/bin/bash
# This script runs test (`gofmt`, `go vet` & `go test`)
# If `gofmt` & `golint` has output (means something wrong),
# it will exit with non-zero status
set -e

DIR=$(cd $(dirname ${0})/.. && pwd)
cd ${DIR}

TARGET=$(find . -name "*.go")
echo -e "Run gofmt"
FMT_RES=$(gofmt -l ${TARGET})
if [ -n "${FMT_RES}" ]; then
    echo -e "gofmt failed: \n${FMT_RES}"
    exit 255
fi

echo -e "Run go vet"
go list -f '{{.Dir}}' ./... | xargs go tool vet
if [ $? -ne 0 ]; then
    echo -e "go vet failed"
    exit 255
fi

echo -e "Run go test"
go test -v ./...
