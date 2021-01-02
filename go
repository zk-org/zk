#!/bin/sh

VERSION=`git describe --tags 2> /dev/null`
BUILD=`git rev-parse --short HEAD`

go $1 -tags "fts5 icu" -ldflags "-X=main.Version=$VERSION -X=main.Build=$BUILD" ${@:2}
