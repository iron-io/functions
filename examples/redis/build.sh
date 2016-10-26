#!/bin/bash
set -ex

# build image
docker build --build-arg FUNCPKG=$(pwd | sed "s|$GOPATH/src/||") -t iron/func-redis .
