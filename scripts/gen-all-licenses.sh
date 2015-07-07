#!/bin/bash
# Generate all avairable licenses
set -e

DIR=$(cd $(dirname ${0})/.. && pwd)
cd ${DIR}

OUTDIR="test-licenses"
rm -fr ${OUTDIR}
mkdir ${OUTDIR}

make build
./bin/license -list-keys \
    | xargs -P 3 -I {} ./bin/license -output=${OUTDIR}/{} -no-cache {} 

ls ${OUTDIR}
