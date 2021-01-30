#!/bin/bash

set -euo pipefail

BWHITE='\033[1;37m'
NC='\033[0m'

cd "$(dirname "$0")"

TESTS="${1-*.test}"

export CLUTTER="../../bin/clutter"

for t in ${TESTS}; do
  printf "${BWHITE}${t}${NC}:\n"
  ./clitest "$t"
done
