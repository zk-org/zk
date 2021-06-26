# Build zk in the current folder.
build:
	$(call go,build)

# Build and install `zk` using go's default bin directory.
install:
	$(call go,install)

# Run unit tests.
test:
	$(call go,test,./...)

# Produce a release bundle for all platforms.
dist: dist-macos dist-linux
	rm -f zk

# Produce a release bundle for macOS.
dist-macos:
	rm -f zk && make && zip -r "zk-${VERSION}-macos-`uname -m`.zip" zk

# Produce a release bundle for Linux.
dist-linux:
	rm -f zk && docker run --rm -v "${PWD}":/usr/src/zk -w /usr/src/zk mickaelmenu/zk-xcompile:linux-i386 /bin/bash -c 'make' && tar -zcvf "zk-${VERSION}-linux-i386.tar.gz" zk
	rm -f zk && docker run --rm -v "${PWD}":/usr/src/zk -w /usr/src/zk mickaelmenu/zk-xcompile:linux-amd64 /bin/bash -c 'make' && tar -zcvf "zk-${VERSION}-linux-amd64.tar.gz" zk
	rm -f zk && docker run --rm -v "${PWD}":/usr/src/zk -w /usr/src/zk mickaelmenu/zk-xcompile:linux-arm64 /bin/bash -c 'make' && tar -zcvf "zk-${VERSION}-linux-arm64.tar.gz" zk

# Clean build products.
clean:
	rm -rf zk*


VERSION := `git describe --tags --match v[0-9]* 2> /dev/null`
BUILD := `git rev-parse --short HEAD`

ENV_PREFIX := CGO_ENABLED=1
# Add necessary env variables for Apple Silicon.
ifeq ($(shell uname -sm),Darwin arm64)
	ENV_PREFIX := $(ENV) GOARCH=arm64 CGO_CFLAGS="-I/opt/homebrew/opt/icu4c/include" CGO_LDFLAGS="-L/opt/homebrew/opt/icu4c/lib"
endif

# Wrapper around the go binary, to set all the default parameters.
define go
	$(ENV_PREFIX) go $(1) -tags "fts5 icu" -ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)" $(2)
endef

