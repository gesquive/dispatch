#
#  Makefile
#
#  The kickoff point for all project management commands.
#

GOCC := go

# Program version
VERSION := $(shell git describe --always --tags)

# Binary name for bintray
BIN_NAME=dispatch

# Project owner for bintray
OWNER=gesquive

# Project name for bintray
PROJECT_NAME=dispatch

# Project url used for builds
# examples: github.com, bitbucket.org
REPO_HOST_URL=github.com

# Grab the current commit
GIT_COMMIT=$(shell git rev-parse HEAD)

# Check if there are uncommited changes
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)

# Use a local vendor directory for any dependencies; comment this out to
# use the global GOPATH instead
# GOPATH=$(PWD)

INSTALL_PATH=$(GOPATH)/src/${REPO_HOST_URL}/${OWNER}/${PROJECT_NAME}
LOCAL_BIN=bin
GOTEMP:=$(shell mktemp -d)

export PATH := ${LOCAL_BIN}:${PATH}

default: test build

.PHONY: help
help:
	@echo 'Management commands for $(PROJECT_NAME):'
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
	 awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Compile the project
	@echo "building ${OWNER} ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	${GOCC} build -ldflags "-X main.version=${VERSION} -X main.dirty=${GIT_DIRTY}" -o ${BIN_NAME}

.PHONY: install
install: build ## Install the binary
	install -d ${DESTDIR}/usr/local/bin/
	install -m 755 ./${BIN_NAME} ${DESTDIR}/usr/local/bin/${BIN_NAME}

.PHONY: deps
deps: glide ## Download project dependencies
	glide install

.PHONY: test
test: ## Run golang tests
	${GOCC} test ./...

.PHONY: bench
bench: glide ## Run golang benchmarks
	${GOCC} test -benchmem -bench=. ./...

.PHONY: clean
clean: ## Clean the directory tree
	${GOCC} clean
	rm -f ./${BIN_NAME}.test
	rm -f ./${BIN_NAME}
	rm -rf ./${LOCAL_BIN}
	rm -rf ./dist

.PHONY: build-dist
build-dist: gox
	gox -verbose \
	-ldflags "-X main.version=${VERSION} -X main.dirty=${GIT_DIRTY}" \
	-os="linux darwin windows" \
	-arch="amd64 386" \
	-output="dist/{{.OS}}-{{.Arch}}/{{.Dir}}" .

.PHONY: package-dist
package-dist: gop
	gop --delete \
	--os="linux darwin windows" \
	--arch="amd64 386" \
	--archive="tar.gz" \
	--files="LICENSE README.md pkg" \
	--input="dist/{{.OS}}-{{.Arch}}/{{.Dir}}" \
	--output="dist/{{.Dir}}-${VERSION}-{{.OS}}-{{.Arch}}.{{.Archive}}" .

.PHONY: dist
dist: build-dist package-dist ## Cross compile and package the full distribution

.PHONY: fmt
fmt: ## Reformat the source tree with gofmt
	find . -name '*.go' -not -path './.vendor/*' -exec gofmt -w=true {} ';'

.PHONY: link
link: $(INSTALL_PATH) ## Symlink this project into the GOPATH
$(INSTALL_PATH):
	@mkdir -p `dirname $(INSTALL_PATH)`
	@ln -s $(PWD) $(INSTALL_PATH) >/dev/null 2>&1

${LOCAL_BIN}:
	@mkdir -p ${LOCAL_BIN}

.PHONY: glide
glide: bin/glide
bin/glide: ${LOCAL_BIN}
	@echo "Installing glide"
	@export GOPATH=${GOTEMP} && ${GOCC} get -u github.com/Masterminds/glide
	@cp ${GOTEMP}/bin/glide ${LOCAL_BIN}
	@glide --version
	@rm -rf ${GOTEMP}

.PHONY: gox
gox: bin/gox
bin/gox: ${LOCAL_BIN}
	@echo "Installing gox"
	@GOPATH=${GOTEMP} ${GOCC} get -u github.com/mitchellh/gox
	@cp ${GOTEMP}/bin/gox ${LOCAL_BIN}/gox
	@rm -rf ${GOTEMP}

.PHONY: gop
gop: bin/gop
bin/gop: ${LOCAL_BIN}
	@echo "Installing gop"
	@export GOPATH=${GOTEMP} && ${GOCC} get -u github.com/gesquive/gop
	@cp ${GOTEMP}/bin/gop ${LOCAL_BIN}
	@gop --version
	@rm -rf ${GOTEMP}

