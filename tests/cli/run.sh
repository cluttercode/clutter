#!/bin/bash

set -euo pipefail

BWHITE='\033[1;37m'
NC='\033[0m'

cd "$(dirname "$0")"

TESTS="${1-*.clitest}"

export CLUTTER="../../bin/clutter"

if [[ ! -x ${CLUTTER} ]]; then
  echo "error: clutter need to be built first. run: make clutter".
  exit 1
fi

for t in ${TESTS}; do
  printf "${BWHITE}${t}${NC}:\n"
  ./clitest --diff-options "-u -w" "$t"
done
