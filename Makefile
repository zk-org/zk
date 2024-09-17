# Build zk in the current folder.
build:
	$(call go,build)

# Build and install `zk` using go's default bin directory.
install:
	$(call go,install)

# Run unit tests.
test:
	$(call go,test,./...)

# Run end-to-end tests.
tesh: build
	@PATH=".:$(shell pwd):$(PATH)" tesh tests tests/fixtures

# Run end-to-end tests and prints difference as raw bytes.
teshb: build
	@PATH=".:$(shell pwd):$(PATH)" tesh -b tests tests/fixtures

# Update end-to-end tests.
tesh-update: build
	PATH=".:$(shell pwd):$(PATH)" tesh -u tests tests/fixtures

alpine:
	$(call alpine,build)

# Produce a release bundle for all platforms.
dist: dist-macos dist-linux
	rm -f zk

# Produce a release bundle for macOS.
dist-macos:
	rm -f zk && make && zip -r "zk-${VERSION}-macos-`uname -m`.zip" zk

# Produce a release bundle for Linux.
dist-linux: dist-linux-amd64 dist-linux-arm64 dist-linux-i386 dist-alpine-amd64 dist-alpine-arm64 dist-alpine-i386
dist-linux-amd64:
	rm -f zk \
		&& docker run --rm -v "${PWD}":/usr/src/zk -w /usr/src/zk ghcr.io/zk-org/zk-xcompile:linux-amd64 /bin/bash -c 'make' \
		&& tar -zcvf "zk-${VERSION}-linux-amd64.tar.gz" zk
dist-linux-arm64:
	rm -f zk \
		&& docker run --rm -v "${PWD}":/usr/src/zk -w /usr/src/zk ghcr.io/zk-org/zk-xcompile:linux-arm64 /bin/bash -c 'make' \
		&& tar -zcvf "zk-${VERSION}-linux-arm64.tar.gz" zk
dist-linux-i386:
	rm -f zk \
		&& docker run --rm -v "${PWD}":/usr/src/zk -w /usr/src/zk ghcr.io/zk-org/zk-xcompile:linux-i386 /bin/bash -c 'make' \
		&& tar -zcvf "zk-${VERSION}-linux-i386.tar.gz" zk
dist-alpine-amd64:
	rm -f zk \
		&& docker run --rm -v "${PWD}":/usr/src/zk -w /usr/src/zk ghcr.io/zk-org/zk-xcompile:alpine-amd64 /bin/bash -c 'make alpine' \
		&& tar -zcvf "zk-${VERSION}-alpine-amd64.tar.gz" zk
dist-alpine-arm64:
	rm -f zk \
		&& docker run --rm -v "${PWD}":/usr/src/zk -w /usr/src/zk ghcr.io/zk-org/zk-xcompile:alpine-arm64 /bin/bash -c 'make alpine' \
		&& tar -zcvf "zk-${VERSION}-alpine-arm64.tar.gz" zk
dist-alpine-i386:
	rm -f zk \
		&& docker run --rm -v "${PWD}":/usr/src/zk -w /usr/src/zk ghcr.io/zk-org/zk-xcompile:alpine-i386 /bin/bash -c 'make alpine' \
		&& tar -zcvf "zk-${VERSION}-alpine-i386.tar.gz" zk

# Clean build and docs products.
clean:
	rm -rf zk*
	rm -rf docs-build

### Sphinx Docs ###
# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).
zkdocs: Makefile
	mkdir -p docs-build
	sphinx-build -a docs docs-build 

VERSION := `git describe --tags --match v[0-9]* 2> /dev/null`
BUILD := `git rev-parse --short HEAD`

ENV_PREFIX := CGO_ENABLED=1
# Add necessary env variables for Apple Silicon.
ifeq ($(shell uname -sm),Darwin arm64)
	ENV_PREFIX := $(ENV) GOARCH=arm64
endif

# Wrapper around the go binary, to set all the default parameters.
define go
	$(ENV_PREFIX) go $(1) -tags "fts5" -ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)" $(2)
endef

# Alpine (musl) requires statically linked libs. This should be compatible for
# Void linux and other musl based distros aswell.
define alpine
	$(ENV_PREFIX) go $(1) -tags "fts5" -ldflags "-extldflags=-static -X=main.Version=$(VERSION) -X=main.Build=$(BUILD)" $(2)
endef
