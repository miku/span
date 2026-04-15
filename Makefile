SHELL = /bin/bash
VERSION := 0.2.32
TARGETS = \
          span-amsl-discovery \
		  span-compare \
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
		  span-hcov \
		  span-index \
		  span-import \
		  span-query \
		  span-mail \
		  span-local-data \
		  span-oa-filter \
		  span-redact \
		  span-report \
		  span-tag \
		  span-update-labels

PKGNAME = span
MAKEFLAGS := --jobs=$(shell nproc)


.PHONY: all
all: $(TARGETS)

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
.PHONY: test
test:
	# go get github.com/kylelemons/godebug/pretty
	# go get github.com/kr/pretty
	go test -v -cover ./...
	# go mod tidy

$(TARGETS): %: cmd/%/main.go
	@# CGO_ENABELED required?
	go build -ldflags "-s -w -X github.com/miku/span.AppVersion=$(VERSION)" -o $@ $<
	@$(if $(shell which upx 2>/dev/null),upx -qqq --best --lzma $@,)

.PHONY: clean
clean:
	rm -f $(TARGETS)
	rm -f span-*.upx
	rm -f $(PKGNAME)_*deb
	rm -f $(PKGNAME)-*rpm
	rm -rf ./packaging/deb/$(PKGNAME)/usr
	rm -f coverage.out
	rm -f *.000
	rm -f *.001

# Code quality and performance.
.PHONY: lint
lint:
	staticcheck ./...

.PHONY: bench
bench:
	go test -v -bench ./...

.PHONY: imports
imports:
	go get golang.org/x/tools/cmd/goimports
	goimports -w .

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

.PHONY: update-version
update-version:
	sed -i -e 's@^Version:.*@Version: $(VERSION)@' packaging/deb/span/DEBIAN/control
	sed -i -e 's@^Version:.*@Version:    $(VERSION)@' packaging/rpm/span.spec

