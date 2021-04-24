VERSION := `git describe --tags --match v[0-9]* 2> /dev/null`

all: macos linux
	rm -f zk

macos:
	rm -f zk && ./go build && zip -r "zk-${VERSION}-macos-`uname -m`.zip" zk

linux:
	rm -f zk && docker run --rm -v "${PWD}":/usr/src/zk -w /usr/src/zk mickaelmenu/zk-xcompile:linux-i386 /bin/bash -c './go build' && tar -zcvf "zk-${VERSION}-linux-i386.tar.gz" zk
	rm -f zk && docker run --rm -v "${PWD}":/usr/src/zk -w /usr/src/zk mickaelmenu/zk-xcompile:linux-amd64 /bin/bash -c './go build' && tar -zcvf "zk-${VERSION}-linux-amd64.tar.gz" zk
	rm -f zk && docker run --rm -v "${PWD}":/usr/src/zk -w /usr/src/zk mickaelmenu/zk-xcompile:linux-arm64 /bin/bash -c './go build' && tar -zcvf "zk-${VERSION}-linux-arm64.tar.gz" zk

clean:
	rm -rf zk*
