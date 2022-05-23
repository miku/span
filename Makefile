SHELL = /bin/bash
TARGETS = span-check \
		  span-compare \
		  span-crossref-snapshot \
		  span-crossref-sync \
		  span-export \
		  span-freeze \
		  span-hcov \
		  span-import \
		  span-local-data \
		  span-oa-filter \
		  span-redact \
		  span-report \
		  span-review \
		  span-tag \
		  span-update-labels \
		  span-webhookd \
          span-amsl-discovery \
          span-crossref-members \
          span-crossref-sync \
          span-folio \
          span-genios-modules \
          span-tagger

PKGNAME = span

.PHONY: all assets bench clean clean-docs cloc deb imports lint members names rpm test vet

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
test:
	# go get github.com/kylelemons/godebug/pretty
	# go get github.com/kr/pretty
	go test -v -cover ./...
	# go mod tidy

all: $(TARGETS)

$(TARGETS): %: cmd/%/main.go
	go build -ldflags="-w -s -linkmode=external" -o $@ $<

clean:
	rm -f $(TARGETS)
	rm -f $(PKGNAME)_*deb
	rm -f $(PKGNAME)-*rpm
	rm -rf ./packaging/deb/$(PKGNAME)/usr

# Code quality and performance.
lint:
	golint -set_exit_status ./...

bench:
	go test -v -bench ./...

imports:
	go get golang.org/x/tools/cmd/goimports
	goimports -w .

vet:
	go vet ./...

cover:
	go test -cover ./...

# Packaging related.
deb: all
	mkdir -p packaging/deb/$(PKGNAME)/usr/local/bin
	cp $(TARGETS) packaging/deb/$(PKGNAME)/usr/local/bin
	mkdir -p packaging/deb/$(PKGNAME)/usr/local/share/man/man1
	cp docs/$(PKGNAME).1 packaging/deb/$(PKGNAME)/usr/local/share/man/man1
	mkdir -p packaging/deb/$(PKGNAME)/usr/lib/systemd/system
	cp packaging/span-webhookd.service packaging/deb/$(PKGNAME)/usr/lib/systemd/system/
	cd packaging/deb && fakeroot dpkg-deb --build $(PKGNAME) .
	mv packaging/deb/$(PKGNAME)_*.deb .

rpm: all
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/rpm/$(PKGNAME).spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	cp docs/$(PKGNAME).1 $(HOME)/rpmbuild/BUILD
	cp packaging/span-webhookd.service $(HOME)/rpmbuild/BUILD
	./packaging/rpm/buildrpm.sh $(PKGNAME)
	cp $(HOME)/rpmbuild/RPMS/x86_64/$(PKGNAME)*.rpm .

# Docs related, https://github.com/sunaku/md2man
docs/$(PKGNAME).1: docs/$(PKGNAME).md
	md2man-roff docs/$(PKGNAME).md > docs/$(PKGNAME).1

clean-docs:
	rm -f docs/$(PKGNAME).1

# Some lists, refs #13587.
members: assets/crossref/members.json
	@echo "Note: Run rm $< manually to rebuild."

assets/crossref/members.json: span-crossref-members
	span-crossref-members | jq -rc '.message.items[].prefix[] | {(.value | tostring): .name | gsub("^[[:space:]]+"; "") | gsub("[[:space:]]+$$"; "")}' | jq -s add > $@

names: assets/crossref/names.ndj
	@echo "Note: Run rm $< manually to rebuild."

# Primary and other names.
assets/crossref/names.ndj: span-crossref-members
	span-crossref-members | jq -rc '.message.items[]| {"primary": .["primary-name"], "names": .["names"]}' > $@

assets/genios/dbmap.generated.json:
	# This is here to document the command, mainly (siskin v0.78.2, 344ca56b72a99074c71e45154ba32089c4f2e015 or later).
	taskdo GeniosLatestReloadList --kind all
	taskcat GeniosLatestReloadList --kind all | span-genios-modules | jq --sort-keys . > $@

