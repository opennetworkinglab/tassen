#!/usr/bin/env sh
TASSEN_DIR=$(pwd)/../../

echo ${TASSEN_DIR}

docker run -v ${TASSEN_DIR}:/Tassen -w /Tassen/util/go-gen-const \
  --entrypoint ./go-gen-p4-const.py opennetworking/p4mn:stable \
  -o /Tassen/mapr/fabric/p4info.go \
  fabric /Tassen/mapr/p4c-out/fabric/p4info.txt