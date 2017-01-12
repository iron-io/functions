#!/bin/bash
set -x

fn build

PAYLOAD='{"input": "yoooo"}'

# test it
echo $PAYLOAD | fn run -e TEST=1