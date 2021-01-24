#!/usr/bin/env bash
set -e

TAG_1=$(git tag --sort=-version:refname | head -1)
TAG_2=$(git tag --sort=-version:refname | head -2 | tail -1)

echo "Changes in ${TAG_1}  (previous tag = ${TAG_2}):"
git log ${TAG_2}..${TAG_1} --oneline
