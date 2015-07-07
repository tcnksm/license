#!/bin/bash
set -e

DIR=$(cd $(dirname ${0})/.. && pwd)
cd ${DIR}

rm -rf pkg/
gox \
    -output "pkg/{{.OS}}_{{.Arch}}/{{.Dir}}"
