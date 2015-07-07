#!/bin/bash
# Generate all avairable licenses
set -e

DIR=$(cd $(dirname ${0})/.. && pwd)
cd ${DIR}

OUTDIR="test-licenses"

if [ -d ${OUTDIR} ]; then
    echo "${OUTDIR} is already exist"
    exit 1
fi

mkdir ${OUTDIR}

make build
for key in $(./bin/license -list-keys); do
    ./bin/license -output=${OUTDIR}/${key} -no-cache ${key}
    sleep 10s
done

ls ${OUTDIR}

