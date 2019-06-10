ifndef VERBOSE
	MAKEFLAGS += --silent
endif

PKGS=$(shell go list ./... | grep -v /vendor)
FILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")
PKG_NAME=github.com/DerTiedemann/joaquin

.PHONY: build_all
build_all: build build_linux  ## build for all platforms

.PHONY: build
build:  ## build joaquin
	go build \
		-mod vendor \
		-o bin/joaquin \
		-ldflags "-s -w \
			-X main.gitVersion=$$(git describe --tags 2>/dev/null || echo pre-release) \
			-X main.gitCommit=$$(git rev-parse HEAD) \
			-X main.buildDate=$$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
		main.go

.PHONY: build_linux
build_linux:  ## build joaquin for linux
	GOOS=linux \
	GOARCH=amd64 \
	go build \
		-mod vendor \
		-o bin/joaquin_linux \
		-ldflags "-s -w \
			-X main.gitVersion=$$(git describe --tags 2>/dev/null || echo pre-release) \
			-X main.gitCommit=$$(git rev-parse HEAD) \
			-X main.buildDate=$$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
		main.go

.PHONY: install
install: build ## install joaquin
	cp bin/joaquin ${GOPATH}/bin/joaquin

.PHONY: update_deps
update_deps: ## Update Dependencies
	go mod verify
	go mod tidy
	rm -rf vendor
	go mod vendor