#!/bin/bash
set -x

fn build

PAYLOAD='{"username": "getiron"}'

# test it
echo $PAYLOAD | fn run