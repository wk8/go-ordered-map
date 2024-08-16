#!/usr/bin/env bash

set -ex

# might as well run a little longer
export FUZZ_TIME=20s

make
