GITHUB_ORG  = mdoubez
GITHUB_REPO = filestat_exporter
VERSION    ?= 0.0.1

# Go projet
GO = go

# Inject version information
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
REVISION ?= $(shell git rev-parse --short HEAD)
BUILDUSER ?= $(USER)
BUILDDATE ?= $(shell date +%FT%T%z)
LDFLAGS = -s -X github.com/prometheus/common/version.Version=$(VERSION) \
		     -X github.com/prometheus/common/version.Branch=$(BRANCH) \
		     -X github.com/prometheus/common/version.Revision=$(REVISION) \
		     -X github.com/prometheus/common/version.BuildUser=$(BUILDUSER) \
		     -X github.com/prometheus/common/version.BuildDate=$(BUILDDATE)

# Distribution
DIST_DIR?=./dist
EXPORTER=filestat_exporter
DIST_ARCHITECTURES=darwin-amd64 linux-amd64 windows-amd64

# Main source files
SRCS = $(wildcard *.go)

# ------------------------------------------------------------------------
# Main targets
# - all: check code and build it
# - build: build exporter for current platform
# - clean: remove build files
# - check: run all following checks
#   - fmt: run formating check
#   - vet: vetting code
# - dist: build distribution packages
#   - dist-linux-amd64/dist-darwin-amd64/...: distribution for arch
# - version: display version number
# - run: launch exporter on sample config
.PHONY: all build clean check dist fmt vet run

all:: vet fmt build

build: $(EXPORTER)

clean:
	@rm -f $(EXPORTER)

check: fmt vet

fmt:
	@$(GO) fmt ./...

vet:
	@$(GO) vet ./...

run:
	@$(GO) run $(SRCS) --log.level=debug '*.*'

version:
	@echo $(VERSION)

DIST_EXPORTER=$(DIST_DIR)/$(EXPORTER)-$(VERSION)
dist: $(foreach ARCH, $(DIST_ARCHITECTURES), $(DIST_EXPORTER).$(ARCH).tar.gz)
dist-%: $(DIST_EXPORTER).$*.tar.gz

# ------------------------------------------------------------------------
# Build exporter


$(EXPORTER): $(SRCS)
	@$(GO) build -ldflags "$(LDFLAGS)" $(SRC)

$(DIST_EXPORTER)/:
	@mkdir -p $@
$(DIST_EXPORTER).%.tar.gz: $(DIST_EXPORTER).%/$(EXPORTER)
	@echo "Packaging $(notdir $@)"
	@cd $(dir $<) ; tar czf $(abspath $@) .
$(DIST_EXPORTER).%/$(EXPORTER): GOOS=$(word 1,$(subst -, ,$*))
$(DIST_EXPORTER).%/$(EXPORTER): GOARCH=$(word 2,$(subst -, ,$*))
$(DIST_EXPORTER).%/$(EXPORTER): $(SRCS)
	@echo "Building $(notdir $@) GOOS=$(GOOS) GOARCH=$(GOARCH)"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -ldflags "$(LDFLAGS)" -o $@ $(SRCS)

