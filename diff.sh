#!/bin/bash

set +eu -o pipefail

: ${NEW:=${1}}
: ${OLD:=${2}}

: ${CHAR_CNT:=${3}}

if [[ -n ${CHAR_CNT} ]]; then
  cat ${NEW} | cut -c-${CHAR_CNT} > ${NEW}.part
  cat ${OLD} | cut -c-${CHAR_CNT} > ${OLD}.part

  NEW="${NEW}.part"
  OLD="${OLD}.part"
fi

diff -u ${NEW} ${OLD}
