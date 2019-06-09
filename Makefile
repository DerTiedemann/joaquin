PKGS=$(shell go list ./... | grep -v /vendor)
FILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")
PKG_NAME=github.com/DerTiedemann/joaquin


.PHONY: build
build:  ## build joaquin
	go build \
		-mod vendor \
		-o bin/hkube \
		-ldflags "-s -w \
			-X $(PKG_NAME)/gitVersion=$$(git describe --tags 2>/dev/null || echo pre-release) \
			-X $(PKG_NAME)/gitCommit=$$(git rev-parse HEAD) \
			-X $(PKG_NAME)/buildDate=$$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
		main.go

.PHONY: install
install: build ## install joaquin
	cp bin/joaquin ${GOPATH}/bin/joaquin