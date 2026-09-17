SHELL = /bin/bash
VERSION := 0.3.0
# One binary. The names span used to install are symlinks to it: it reads the
# name it was invoked under and runs the matching command, which keeps every
# existing script working. Keep LEGACY in sync with internal/cli (tested).
TARGET = span
LEGACY = \
	span-compact \
	span-crossref-fast-snapshot \
	span-crossref-fastproc \
	span-crossref-members \
	span-crossref-snapshot \
	span-crossref-sync \
	span-crossref-table \
	span-doisniffer \
	span-export \
	span-folio \
	span-freeze \
	span-import \
	span-index \
	span-local-data \
	span-mail \
	span-oa-filter \
	span-redact \
	span-tag \
	span-update-labels

PKGNAME = span
MAKEFLAGS := --jobs=$(shell nproc 2>/dev/null || sysctl -n hw.physicalcpu)

.PHONY: all
all: $(TARGET) $(LEGACY)

.PHONY: test
test:
	go test -v -cover ./...

LDFLAGS = -s -w -X github.com/miku/span.AppVersion=$(VERSION)
GOFILES := $(shell find . -name '*.go' -not -path './.git/*' -not -path './build/*') go.mod go.sum

$(TARGET): $(GOFILES)
	go build -ldflags "$(LDFLAGS)" -o $@ ./cmd/span

# The same links the packages install, so a working copy behaves like an
# installed span.
$(LEGACY): $(TARGET)
	ln -sf $(TARGET) $@

.PHONY: clean
clean:
	rm -f $(TARGET) $(LEGACY)
	rm -rf build/
	rm -f $(PKGNAME)_*deb
	rm -f $(PKGNAME)-*rpm
	rm -f coverage.out
	rm -f *.000
	rm -f *.001

# Code quality and performance. staticcheck is pinned in the go.mod tool block,
# so this needs no separately installed binary and matches CI.
.PHONY: lint
lint:
	go vet ./...
	go tool staticcheck ./...

.PHONY: bench
bench:
	go test -v -bench ./...

# nfpm-based packaging (preferred).
SEMVER := $(shell echo $(VERSION) | sed 's/^v//')

# Cross-compiled linux/amd64 binary used for packaging (deb/rpm), kept in a
# separate directory so it never overwrites the native dev binary above and
# the host platform (e.g. macOS arm64) does not leak into the package.
build/span: $(GOFILES)
	@mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $@ ./cmd/span

.PHONY: deb
deb: build/span
	SEMVER=$(SEMVER) GOARCH=amd64 nfpm package -p deb -f nfpm.yaml

.PHONY: rpm
rpm: build/span
	SEMVER=$(SEMVER) GOARCH=amd64 nfpm package -p rpm -f nfpm.yaml

# Docs related, https://github.com/sunaku/md2man
docs/$(PKGNAME).1: docs/$(PKGNAME).md
	md2man-roff docs/$(PKGNAME).md > docs/$(PKGNAME).1

.PHONY: clean-docs
clean-docs:
	rm -f docs/$(PKGNAME).1

# Some lists, refs #13587.
.PHONY: members
members: assets/crossref/members.json
	@echo "Note: Run rm $< manually to rebuild."

assets/crossref/members.json: span
	./span crossref members | jq -rc '.message.items[].prefix[] | {(.value | tostring): .name | gsub("^[[:space:]]+"; "") | gsub("[[:space:]]+$$"; "")}' | jq -s add > $@

.PHONY: names
names: assets/crossref/names.ndj
	@echo "Note: Run rm $< manually to rebuild."

# Primary and other names.
assets/crossref/names.ndj: span
	./span crossref members | jq -rc '.message.items[]| {"primary": .["primary-name"], "names": .["names"]}' > $@

