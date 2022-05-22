#!/usr/bin/env bash

set -ueEo pipefail

LOCAL_BUILD_IMAGE=zk_build
BASE_OS=ubuntu
BASE_OS_VERSION=focal
GO_VERSION=1.17.4

function build_docker_image {
  docker build \
    -f- \
    -t "$LOCAL_BUILD_IMAGE" \
    . <<EOF
FROM ${BASE_OS}:${BASE_OS_VERSION}
RUN DEBIAN_FRONTEND=noninteractive apt -y update && \
    DEBIAN_FRONTEND=noninteractive apt -y upgrade && \
    DEBIAN_FRONTEND=noninteractive apt install -y --no-install-recommends \
      ca-certificates \
      curl \
      build-essential \
      git \
      libicu-dev 

# Install newer golang
WORKDIR /tmp
RUN curl -fsSL https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz -o go.tar.gz && \
    tar -zx -C / -f go.tar.gz && \
    rm -rf go.tar.gz
ENV GOPATH /go
ENV PATH $PATH:/go/bin:$GOPATH/bin
# If you enable this, then gcc is needed to debug your app
ENV CGO_ENABLED 0

WORKDIR /usr/src/myapp
VOLUME /usr/src/myapp 
CMD ["make"]
EOF
}

if [[ -z "$(docker image ls -q $LOCAL_BUILD_IMAGE)" ]]; then
  build_docker_image
fi

docker run -it -v "$PWD":/usr/src/myapp -w /usr/src/myapp "$LOCAL_BUILD_IMAGE" 

