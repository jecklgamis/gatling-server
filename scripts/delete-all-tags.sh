#!/usr/bin/env bash
for tag in $(git tag -l); do
  git tag -d ${tag}
  echo "Deleted ${tag}"
done
git tag -a v0.0.0 -m "The origin of species"
