#!/usr/bin/env bash

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
MAPR_LOG=$(readlink -f "${DIR}"/../log/mapr.log)

mapr -port 28001 -target_addr 127.0.0.1:28000 > "${MAPR_LOG}" 2>&1
