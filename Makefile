SHELL = /bin/bash
VERSION := 0.2.37
TARGETS = \
		  span-compact \
		  span-compare-file \
          span-crossref-members \
		  span-crossref-fast-snapshot \
		  span-crossref-fastproc \
		  span-crossref-snapshot \
          span-crossref-sync \
		  span-crossref-table \
		  span-doisniffer \
		  span-export \
          span-folio \
		  span-freeze \
		  span-index \
		  span-import \
		  span-mail \
		  span-local-data \
		  span-oa-filter \
		  span-redact \
		  span-report \
		  span-tag \
		  span-update-labels

PKGNAME = span
MAKEFLAGS := --jobs=$(shell nproc 2>/dev/null || sysctl -n hw.physicalcpu)

.PHONY: all
all: $(TARGETS)

.PHONY: test
test:
	go test -v -cover ./...

LDFLAGS = -s -w -X github.com/miku/span.AppVersion=$(VERSION)
GOFILES := $(shell find . -name '*.go' -not -path './.git/*' -not -path './build/*') go.mod go.sum

$(TARGETS): %: $(GOFILES)
	go build -ldflags "$(LDFLAGS)" -o $@ ./cmd/$@

.PHONY: clean
clean:
	rm -f $(TARGETS)
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

# Cross-compiled linux/amd64 binaries used for packaging (deb/rpm), kept in a
# separate directory so they never overwrite the native dev binaries above and
# the host platform (e.g. macOS arm64) does not leak into the package.
BUILD_TARGETS = $(addprefix build/,$(TARGETS))

$(BUILD_TARGETS): build/%: $(GOFILES)
	@mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $@ ./cmd/$*

.PHONY: deb
deb: $(BUILD_TARGETS)
	SEMVER=$(SEMVER) GOARCH=amd64 nfpm package -p deb -f nfpm.yaml

.PHONY: rpm
rpm: $(BUILD_TARGETS)
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

assets/crossref/members.json: span-crossref-members
	span-crossref-members | jq -rc '.message.items[].prefix[] | {(.value | tostring): .name | gsub("^[[:space:]]+"; "") | gsub("[[:space:]]+$$"; "")}' | jq -s add > $@

.PHONY: names
names: assets/crossref/names.ndj
	@echo "Note: Run rm $< manually to rebuild."

# Primary and other names.
assets/crossref/names.ndj: span-crossref-members
	span-crossref-members | jq -rc '.message.items[]| {"primary": .["primary-name"], "names": .["names"]}' > $@

