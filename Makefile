SHELL = /bin/bash
VERSION := 0.2.19
TARGETS = \
          span-amsl-discovery \
		  span-compare \
          span-crossref-members \
		  span-crossref-fast-snapshot \
		  span-crossref-snapshot \
          span-crossref-sync \
		  span-crossref-table \
		  span-doisniffer \
		  span-export \
          span-folio \
		  span-freeze \
		  span-hcov \
		  span-import \
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
	go build -o $@ $<

.PHONY: clean
clean:
	rm -f $(TARGETS)
	rm -f $(PKGNAME)_*deb
	rm -f $(PKGNAME)-*rpm
	rm -rf ./packaging/deb/$(PKGNAME)/usr
	rm -f coverage.out

# Code quality and performance.
.PHONY: lint
lint:
	golint -set_exit_status ./...

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

# Packaging related.
.PHONY: deb
deb: all
	mkdir -p packaging/deb/$(PKGNAME)/usr/local/bin
	cp $(TARGETS) packaging/deb/$(PKGNAME)/usr/local/bin
	mkdir -p packaging/deb/$(PKGNAME)/usr/local/share/man/man1
	cp docs/$(PKGNAME).1 packaging/deb/$(PKGNAME)/usr/local/share/man/man1
	mkdir -p packaging/deb/$(PKGNAME)/usr/lib/systemd/system
	cp packaging/span-webhookd.service packaging/deb/$(PKGNAME)/usr/lib/systemd/system/
	cd packaging/deb && fakeroot dpkg-deb -Zzstd --build $(PKGNAME) .
	mv packaging/deb/$(PKGNAME)_*.deb .


.PHONY: rpm
rpm: all
	# on deb based distros, you may need:
	# sudo rpm --initdb && sudo chmod -R a+r /var/lib/rpm/
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	mkdir -p $(HOME)/rpmbuild/SOURCES/span
	cp ./packaging/rpm/$(PKGNAME).spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/SOURCES/span
	cp docs/$(PKGNAME).1 $(HOME)/rpmbuild/SOURCES/span
	cp packaging/span-webhookd.service $(HOME)/rpmbuild/SOURCES/span
	./packaging/rpm/buildrpm.sh $(PKGNAME)
	cp $(HOME)/rpmbuild/RPMS/x86_64/$(PKGNAME)-$(VERSION)*.rpm .

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
	sed -i -e 's@AppVersion =.*@AppVersion = "$(VERSION)"@' common.go
	sed -i -e 's@^Version:.*@Version: $(VERSION)@' packaging/deb/span/DEBIAN/control
	sed -i -e 's@^Version:.*@Version:    $(VERSION)@' packaging/rpm/span.spec

