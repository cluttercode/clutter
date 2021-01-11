#!/bin/bash
set -euo pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
BWHITE='\033[1;37m'
NC='\033[0m'


cd "$(dirname "$0")"

clutter() {
  ../../bin/clutter "$@"
}

TESTS="${*-$(ls -d -- */)}"

tmp=$(mktemp -d)
trap 'rm -fR "${tmp}"' 0

run() {
  TEST="$1"

  CMD="clutter $(sed -e "s/TEST/${TEST}/g" < "${TEST}/args")"

  printf "${BWHITE}${tst}${NC}:\n\$ ${CMD}\n"

  if [[ -r "${TEST}/exitcode" ]]; then
    expected_rc="$(cat "${TEST}/exitcode")"
  else
    expected_rc=
  fi

  if $CMD 0<&- > "${tmp}/stdout" 2> $"${tmp}/stderr"; then
    rc=0
  else
    rc=$?
  fi

  printf "> done, exitcode=${rc}\n"

  pass=1

  if [[ -n $expected_rc ]]; then
    if [[ $expected_rc -ne $rc ]]; then
      pass=0
      printf "${RED}exit code ${rc}!=${expected_rc}${NC}\n"
    fi
  fi

  if [[ -r "${TEST}/stdout" ]]; then
    if ! diff "${tmp}/stdout" "${TEST}/stdout"; then
      pass=0
      printf "${RED}stdout differs${NC}\n"
    fi
  elif ! diff "${tmp}/stdout" empty-file; then
    pass=0
    printf "${RED}stdout is not empty${NC}\n"
  fi

  if [[ -r "${TEST}/stderr" ]]; then
    if ! diff "${tmp}/stderr" "${TEST}/stderr"; then
      pass=0
      printf "${RED}stderr differs${NC}\n"
    fi
  elif ! diff "${tmp}/stderr" empty-file; then
    pass=0
    printf "${RED}stderr is not empty${NC}\n"
  fi

  if [[ $pass -eq 0 ]]; then
    printf "${BWHITE}${TEST} ${RED}failed${NC}\n"
    exit 1
  fi

  printf "${BWHITE}${TEST} ${GREEN}passed${NC}\n"
}

for tst in ${TESTS}; do
  tst="${tst%/}"
  run "${tst}"
done
