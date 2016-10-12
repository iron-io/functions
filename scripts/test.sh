#!/bin/bash

export LOG_LEVEL=debug
go test -v $(glide nv | grep -v examples | grep -v tool)