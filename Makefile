GITHUB_ORG  = michael-doubez
GITHUB_REPO = filestat_exporter
VERSION    ?= $(shell git describe --tags --match="v[0-9]*")

# Binary build parameters
#   - build in release mode
RELEASE_MODE ?= 0

# Go projet
GO ?= go
GOBIN ?= $(shell $(GO) env GOBIN)
ifeq ($(GOBIN),)
  GOPATH ?= $(shell $(GO) env GOPATH)
  ifeq ($(GOPATH),)
    $(error Expecting GOPATH to be set)
  endif
  GOBIN = $(GOPATH)/bin
endif
GOLINT = $(shell ls $(GOBIN)/staticcheck 2>/dev/null || true)

# Inject version information
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
REVISION ?= $(shell git rev-parse --short HEAD)
BUILDUSER ?= $(USER)
BUILDDATE ?= $(shell date +%FT%T%z)
LDFLAGS = -X github.com/prometheus/common/version.Version=$(VERSION) \
          -X github.com/prometheus/common/version.Branch=$(BRANCH) \
          -X github.com/prometheus/common/version.Revision=$(REVISION) \
          -X github.com/prometheus/common/version.BuildUser=$(BUILDUSER) \
          -X github.com/prometheus/common/version.BuildDate=$(BUILDDATE)

# Distribution
DIST_DIR?=./dist
EXPORTER=filestat_exporter
DIST_ARCHITECTURES=darwin-amd64 linux-amd64 windows-amd64 linux-arm64

# Main source files
SRC_MAIN = cmd/filestat_exporter.go
SRCS = $(SRC_MAIN) $(wildcard internal/*/*.go)

# ------------------------------------------------------------------------
# Main targets
# - all: check code and build it
# - build: build exporter for current platform
# - clean: remove build files
# - check: run all following checks
#   - fmt: run formating check
#   - vet: vetting code
#   - lint: run static check
#   - test: run unit tests
# - dist: build distribution packages
#   - dist-linux-amd64/dist-darwin-amd64/...: distribution for arch
# - go-version: version of golang in gomod
# - run: launch exporter on sample config
# - version: display version number
.PHONY: all build clean check dist fmt vet test lint run dist dist-% docker-build docker-tag docker-push go-version

all:: check build

build: $(EXPORTER)

clean:
	@rm -f $(EXPORTER)

check: fmt vet lint test

fmt:
	@$(GO) fmt ./...

vet:
	@$(GO) vet ./...

test:
	@$(GO) test -cover ./internal/...

lint:
ifeq ($(GOLINT),)
	@echo >&2 "Warning: staticcheck not installed - lint skipped"
	@echo >&2 "         run 'go install honnef.co/go/tools/cmd/staticcheck@latest' to install"
else
	@$(GOLINT) ./...
endif

RUN_OPTIONS=-log.level=debug -metric.crc32 -metric.nb_lines -tree.name debug -tree.root /
RUN_PATTERN?='/etc/*.conf'
run:
	@$(GO) run $(SRC_MAIN) $(RUN_OPTIONS) $(RUN_PATTERN)

version:
	@echo $(VERSION)

GO_VERSION=$(shell sed -n 's/^go \(.*\)$$/\1/p' go.mod)
ifeq ($(GO_VERSION),)
  GO_VERSION="latest"
endif
go-version:
	@echo "$(GO_VERSION)"

DIST_EXPORTER=$(DIST_DIR)/$(EXPORTER)-$(VERSION)
dist: $(foreach ARCH, $(DIST_ARCHITECTURES), $(DIST_EXPORTER).$(ARCH).tar.gz)
dist-%: $(DIST_EXPORTER).%.tar.gz
	@echo "Done generating $(notdir $<)"

.PRECIOUS: $(DIST_EXPORTER).%.tar.gz

# ------------------------------------------------------------------------
# Build and package exporter

# List of files to include in packages
PACKAGE_FILES = LICENSE NOTICE filestat.yaml

# In release mode
#   - build without debug symbol
ifneq ($(RELEASE_MODE),0)
  LDFLAGS += -s
  LDFLAGS += -w
  BUILD_FLAGS += -trimpath
endif

# Simple build for current os/architecture
$(EXPORTER): $(SRCS)
	@$(GO) build -ldflags "$(LDFLAGS)" -o $@ $(BUILD_FLAGS) $<

# Ensure dist path exists
$(DIST_EXPORTER)/:
	@mkdir -p $@

# Package distribution file
$(DIST_EXPORTER).%.tar.gz: $(DIST_EXPORTER).%/$(EXPORTER) $(PACKAGE_FILES)
	@echo "Packaging $(notdir $@)"
	@cp -f $(PACKAGE_FILES)  $(DIST_EXPORTER).$*/
	@cd $(DIST_DIR) ; tar czf $(abspath $@) $(notdir $(DIST_EXPORTER)).$*/
	@rm -rf $(DIST_EXPORTER).$*/

# Generating exporter for archi
$(DIST_EXPORTER).%/$(EXPORTER): $(DIST_EXPORTER)/
$(DIST_EXPORTER).%/$(EXPORTER): GOOS=$(word 1,$(subst -, ,$*))
$(DIST_EXPORTER).%/$(EXPORTER): GOARCH=$(word 2,$(subst -, ,$*))
$(DIST_EXPORTER).%/$(EXPORTER): $(SRCS)
	@echo "Building $(notdir $@) GOOS=$(GOOS) GOARCH=$(GOARCH)"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -ldflags "$(LDFLAGS)" -o $@ $(SRC_MAIN)

# ------------------------------------------------------------------------
# Docker build of image
docker-build:
	docker build --build-arg GO_VERSION=$(GO_VERSION) --build-arg VERSION=$(VERSION) -t filestat_exporter:$(VERSION:v%=%) .

docker-tag:
	docker tag filestat_exporter:$(VERSION:v%=%) mdoubez/filestat_exporter:$(VERSION:v%=%)
	docker tag filestat_exporter:$(VERSION:v%=%) mdoubez/filestat_exporter:latest

docker-push:
	docker push mdoubez/filestat_exporter:$(VERSION:v%=%)
	docker push mdoubez/filestat_exporter:latest
