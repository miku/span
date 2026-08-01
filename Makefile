SHELL = /bin/bash
VERSION := 0.2.35
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
MAKEFLAGS := --jobs=$(shell nproc 2>/dev/null || sysctl -n hw.physicalcpu)


.PHONY: all
all: $(TARGETS)

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
.PHONY: test
test:
	# go get github.com/kylelemons/godebug/pretty
	# go get github.com/kr/pretty
	go test -v -cover ./...
	# go mod tidy

$(TARGETS): %: $(wildcard cmd/%/*.go)
	@# CGO_ENABELED required?
	go build -ldflags "-s -w -X github.com/miku/span.AppVersion=$(VERSION)" -o $@ ./cmd/$@

.PHONY: clean
clean:
	rm -f $(TARGETS)
	rm -f $(PKGNAME)_*deb
	rm -f $(PKGNAME)-*rpm
	rm -rf ./packaging/deb/$(PKGNAME)/usr
	rm -f coverage.out
	rm -f *.000
	rm -f *.001

# Code quality and performance.
.PHONY: lint
lint:
	go tool staticcheck ./...

.PHONY: bench
bench:
	go test -v -bench ./...

.PHONY: imports
imports:
	go tool goimports -w .

.PHONY: vet
vet:
	go vet ./...

.PHONY: cover
cover:
	go test -cover ./...

# nfpm-based packaging (preferred).
SEMVER := $(shell echo $(VERSION) | sed 's/^v//')

.PHONY: deb
deb: all
	SEMVER=$(SEMVER) GOARCH=amd64 nfpm package -p deb -f nfpm.yaml

.PHONY: rpm
rpm: all
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

# Live indexing pipelines. Two lanes feed the same live SOLR:
#
#   fast-lane  additive daily upsert of the latest crossref slice (frequent)
#   full-lane  dedup + full reindex over the cached corpus, then sweep stale
#              docs (periodic; the GC pass the fast lane cannot do)
#
# Override any variable on the command line, e.g.:
#   make fast-lane SOLR=http://10.0.0.1:8983/solr/biblio DATE=2026-07-31
#
# Requires solrbulk (https://github.com/miku/solrbulk) on PATH. The filterconfig
# comes from FILTERCONFIG (a frozen zip); if empty, span-crossref-fastproc
# fetches it from FOLIO via the OKAPI_URL and OKAPI_TOKEN environment variables.
SOLR           ?= http://localhost:8983/solr/biblio
SOLRBULK       ?= solrbulk
CROSSREF_CACHE ?= /data/finc/crossref
PREFIX         ?= feed-1-
FILTER         ?= index
CROSSREF_SID   ?= 49
FILTERCONFIG   ?=
FASTPROC_FC     = $(if $(FILTERCONFIG),-f $(FILTERCONFIG),)
# fast lane: a single day (default: yesterday, UTC).
DATE  ?= $(shell date -u -d 'yesterday' +%F 2>/dev/null || date -u -v-1d +%F)
# full lane: a window [SINCE, UNTIL] to top up (default: last 30 days, UTC).
SINCE ?= $(shell date -u -d '30 days ago' +%F 2>/dev/null || date -u -v-30d +%F)
UNTIL ?= $(DATE)

# Fast lane: harvest yesterday's crossref slice and upsert it into the live
# index. Additive only (no deletes) - safe because crossref is append-mostly.
.PHONY: fast-lane
fast-lane: span-crossref-sync span-crossref-fastproc
	./span-crossref-sync -c $(CROSSREF_CACHE) -P $(PREFIX) -f $(FILTER) -p zstd -i d -s $(DATE) -e $(DATE) -q
	./span-crossref-fastproc $(FASTPROC_FC) -o - \
		$(CROSSREF_CACHE)/$(PREFIX)$(FILTER)-$(DATE)-$(DATE).json.zst | \
		$(SOLRBULK) -server $(SOLR) -verbose

# Full lane: top up recent slices, dedup the whole cached corpus to the latest
# version per DOI, reindex it, then print a delete-by-query for records the
# reindex did not touch (last_indexed older than the pass start). The sweep is
# printed for review, not executed - pipe it to sh once you trust it.
.PHONY: full-lane
full-lane: span-crossref-sync span-crossref-fast-snapshot span-crossref-fastproc span-index
	@START=$$(date -u +%Y-%m-%dT%H:%M:%SZ); echo "pass start: $$START"; \
	./span-crossref-sync -c $(CROSSREF_CACHE) -P $(PREFIX) -f $(FILTER) -p zstd -i d -s $(SINCE) -e $(UNTIL) -q && \
	./span-crossref-fast-snapshot -o $(CROSSREF_CACHE)/snapshot.json.zst \
		$(CROSSREF_CACHE)/$(PREFIX)$(FILTER)-*.json.zst && \
	./span-crossref-fastproc $(FASTPROC_FC) -o - $(CROSSREF_CACHE)/snapshot.json.zst | \
		$(SOLRBULK) -server $(SOLR) -verbose && \
	echo "reindex done; review the sweep below, then re-run piped to sh to delete:" && \
	./span-index cleanup -s $(SOLR) --sid $(CROSSREF_SID) --until $$START

.PHONY: update-version
update-version:
	sed -i -e 's@^Version:.*@Version: $(VERSION)@' packaging/deb/span/DEBIAN/control
	sed -i -e 's@^Version:.*@Version:    $(VERSION)@' packaging/rpm/span.spec

