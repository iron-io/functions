#!/bin/bash
set -x

./build.sh

PAYLOAD='{"username": "getiron"}'

# test it
echo $PAYLOAD | docker run --rm -i -v func:/func -e TEST=1 iron/func-twitter 