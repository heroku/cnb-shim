.EXPORT_ALL_VARIABLES:

.PHONY: test \
        build \
        clean \
        package \
        release

SHELL=/bin/bash -o pipefail

GO111MODULE := on

VERSION := "0.0.1"

build:
	@GOOS=linux go build -o "bin/release" ./cmd/release/...

clean:
	-rm -f cnb-shim-$(VERSION).tgz
	-rm -f bin/releaser

package: clean build
	@tar cvzf cnb-shim-$(VERSION).tgz bin/ README.md LICENSE

release:
	@git tag $(VERSION)
	@git push --tags origin master