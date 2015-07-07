#!/bin/bash
set -e

DIR=$(cd $(dirname ${0})/.. && pwd)
cd ${DIR}

ORG_PATH="github.com/tcnksm"
REPO_PATH="${ORG_PATH}/license"

docker run \
       -v $PWD:/gopath/src/${REPO_PATH} \
       -w /gopath/src/${REPO_PATH} \
       tcnksm/gox:1.4.2 /bin/bash -c "make package"
