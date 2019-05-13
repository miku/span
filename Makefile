SHELL = /bin/bash
TARGETS = span-import span-export span-tag span-redact span-check span-oa-filter span-update-labels span-crossref-snapshot span-local-data span-freeze span-review span-compare span-webhookd span-report span-hcov span-amsl-discovery span-crossref-members span-crossref-pubprefix
PKGNAME = span

.PHONY: test lint bench deps imports assets vet clean deb rpm cloc clean-docs all

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
test: assets deps
	go get github.com/kylelemons/godebug/pretty
	go get github.com/kr/pretty

	go test ./...

lint:
	golint -set_exit_status ./...

bench:
	go test -v -bench ./...

deps:
	go get -v ./...

imports:
	go get golang.org/x/tools/cmd/goimports
	goimports -w .

assets: assetutil/bindata.go

assetutil/bindata.go:
	go get -u github.com/jteeuwen/go-bindata
	go-bindata -o assetutil/bindata.go -pkg assetutil assets/...

vet:
	go vet ./...

cover:
	go test -cover ./...

all: assets deps $(TARGETS)

$(TARGETS): %: cmd/%/main.go
	go build -ldflags=-linkmode=external -o $@ $<

clean:
	rm -f $(TARGETS)
	rm -f $(PKGNAME)_*deb
	rm -f $(PKGNAME)-*rpm
	rm -rf ./packaging/deb/$(PKGNAME)/usr
	rm -f assetutil/bindata.go

# Just a shortcut.
members: assets/crossref/members.json
	@echo "Note: Run rm $< manually to rebuild."

assets/crossref/members.json:
	span-crossref-members | jq -rc '.message.items[].prefix[] | {(.value | tostring): .name | gsub("[[:space:]]+$$"; "")}' | jq -s add > $@

deb: all
	mkdir -p packaging/deb/$(PKGNAME)/usr/sbin
	cp $(TARGETS) packaging/deb/$(PKGNAME)/usr/sbin
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

cloc:
	cloc --max-file-size 1 --exclude-dir vendor --exclude-dir assets --exclude-dir assetutil --exclude-dir tmp --exclude-dir fixtures .

docs/$(PKGNAME).1: docs/$(PKGNAME).md
	md2man-roff docs/$(PKGNAME).md > docs/$(PKGNAME).1

clean-docs:
	rm -f docs/$(PKGNAME).1

