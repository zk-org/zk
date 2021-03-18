#!/bin/bash

VERSION=`git describe --tags --match v[0-9]* 2> /dev/null`
BUILD=`git rev-parse --short HEAD`

CGO_ENABLED=1 go $1 -tags "fts5 icu" -ldflags "-X=main.Version=$VERSION -X=main.Build=$BUILD" ${@:2}
