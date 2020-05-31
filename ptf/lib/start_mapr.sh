#!/usr/bin/env bash

set -ex

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
MAPR_LOG=$(readlink -f "${DIR}"/../log/mapr.log)
TARGET_P4C_OUT=${TARGET_P4C_OUT:-/undefined}

/workdir/mapr/mapr \
  -port 28001 \
  -target_addr 127.0.0.1:28000 \
  -translator "${TRANSLATOR}" \
  -logical_p4info /workdir/p4src/build/p4info.bin \
  -target_p4_config "${TARGET_P4C_OUT}/p4info.bin,${TARGET_P4C_OUT}/bmv2.json" \
  >"${MAPR_LOG}" 2>&1
