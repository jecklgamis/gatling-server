#!/usr/bin/env bash

function is_running_on_mac() {
  if [[ "$(uname -s)" == *"Darwin"* ]]; then return 0; fi
  return 1
}

function speak() {
  is_running_on_mac && say $1
}