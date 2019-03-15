PWD ?= $(shell pwd)
GO ?= $(shell command -v go 2>/dev/null)
GOFLAGS ?= -buildmode=c-archive -ldflags '-linkmode external'
TUMBLRTVDIR ?= $(PWD)/tumblrtv2
PACKR ?= $(shell command -v packr 2>/dev/null)
EXPORTS ?= CGO_ENABLED=1 GO111MODULE=on

.PHONY: packr-bin
packr-bin:
ifeq ($(strip $(PACKR)),)
	@unset GO111MODULE && \
		$(GO) get github.com/gobuffalo/packr/packr && \
		echo "packr installed, rerunning make" && \
		make && exit
endif

.PHONY: pack
pack: packr-bin
	@export $(EXPORTS) && \
		$(PACKR) build $(GOFLAGS) -o $(TUMBLRTVDIR)/go.a

.PHONY: archive
archive:
	@export $(EXPORTS) && \
		$(GO) build $(GOFLAGS) -o $(TUMBLRTVDIR)/go.a

.DEFAULT_GOAL:=pack
