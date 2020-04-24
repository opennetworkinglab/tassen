#!/usr/bin/env bash

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
MAPR_PY=$(readlink -f "${DIR}"/../mapr/mapr.py)
MAPR_LOG=$(readlink -f "${DIR}"/../log/mapr.log)

python -u "${MAPR_PY}" --server-port 28001 --target-addr localhost:28000 > "${MAPR_LOG}" 2>&1
