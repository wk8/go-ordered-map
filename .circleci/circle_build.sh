#!/usr/bin/env bash

set -ex

# there are too many golangci plugins that don't work for 1.19 just yet, so just skip linting for it
if [[ "$GO_VER" == 1.18.* ]]; then
  make
else
  make test fuzz
fi
