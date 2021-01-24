#!/usr/bin/env bash
result=$(find . -name *.go -exec golint {} \;)
if [ ! -z "${result}" ]; then
  echo ${result}
  echo "Lint failed" && exit 1
else
  echo "Lint successful" && exit 0
fi
