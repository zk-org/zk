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

# Clean build and docs products.
clean:
	rm -rf zk*
	rm -rf docs-build

# Docs
zkdocs:
	mkdir -p docs-build
	sphinx-build -a docs docs-build 

VERSION ?= $(shell \
	if grep -vq '^\$$Format' VERSION.txt 2>/dev/null; then \
		cat VERSION.txt; \
	else \
		git describe --tags --always --dirty --match v[0-9]* 2>/dev/null; \
	fi \
)

ENV_PREFIX := CGO_ENABLED=1
# Add necessary env variables for Apple Silicon.
ifeq ($(shell uname -sm),Darwin arm64)
	ENV_PREFIX := $(ENV) GOARCH=arm64
endif

# Wrapper around the go binary, to set all the default parameters.
define go
	$(ENV_PREFIX) go $(1) -buildvcs=false -tags "fts5" -ldflags "-X=main.Version=$(VERSION)" $(2)
endef

# Alpine (musl) requires statically linked libs. This should be compatible for
# Void linux and other musl based distros aswell.
define alpine
	$(ENV_PREFIX) go $(1) -buildvcs=false -tags "fts5" -ldflags "-extldflags=-static -X=main.Version=$(VERSION)" $(2)
endef
