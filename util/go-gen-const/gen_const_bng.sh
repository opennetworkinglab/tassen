#!/usr/bin/env sh

set -e

TASSEN_DIR=$(pwd)/../../

PTF_DOCKER_IMG=${PTF_DOCKER_IMG:-undefined}
GOLANG_DOCKER_IMG=${GOLANG_DOCKER_IMG:-undefined}

P4INFO_FILE=/tassen/p4src/build/p4info.txt
OUTPUT_FILE=/tassen/mapr/p4info/constants.go

GO_PACKAGE=p4info

echo "Generate Go constants for bng.p4"
echo "P4Info: ${P4INFO_FILE}"

docker run -v ${TASSEN_DIR}:/tassen -w /tassen/util/go-gen-const \
  --entrypoint ./go-gen-p4-const.py ${PTF_DOCKER_IMG} \
  -o ${OUTPUT_FILE} \
  ${GO_PACKAGE} ${P4INFO_FILE}

docker run -v ${TASSEN_DIR}:/tassen -w /tassen/ \
  ${GOLANG_DOCKER_IMG} gofmt -w ${OUTPUT_FILE}

echo "Output: ${OUTPUT_FILE}"
