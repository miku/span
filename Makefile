TARGETS = span

span: imports
	go build -o span cmd/span/main.go

imports:
	goimports -w .

clean:
	rm -f $(TARGETS)
	rm -f span_*deb
	rm -f span-*rpm

test:
	go test -cover ./...

deb: $(TARGETS)
	mkdir -p packaging/deb/span/usr/sbin
	cp $(TARGETS) packaging/deb/span/usr/sbin
	cd packaging/deb && fakeroot dpkg-deb --build span .
	mv packaging/deb/span_*.deb .

rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/rpm/span.spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	./packaging/rpm/buildrpm.sh span
	cp $(HOME)/rpmbuild/RPMS/x86_64/span*.rpm .
