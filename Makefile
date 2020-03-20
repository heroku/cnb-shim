.EXPORT_ALL_VARIABLES:

.PHONY: test \
        build \
        clean \
        package \
        release

SHELL=/bin/bash -o pipefail

GO111MODULE := on

VERSION := "v0.2"

build:
	@GOOS=linux go build -o "bin/release" ./cmd/release/...
	@GOOS=linux go build -o "bin/exports" ./cmd/exports/...

test:
	go test ./... -v

clean:
	-rm -f cnb-shim-$(VERSION).tgz
	-rm -f bin/exports
	-rm -f bin/release

package: clean build
	@tar cvzf cnb-shim-$(VERSION).tgz bin/ README.md LICENSE

release:
	@git tag $(VERSION)
	@git push --tags origin master
