SHELL = /bin/bash
TARGETS = span-import span-export span-tag

# find go-bindata executable on vm
export PATH := /home/vagrant/bin:$(PATH)

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
test: assets deps
	go get github.com/kylelemons/godebug/pretty
	go get github.com/kr/pretty
	go test -v ./...

bench:
	go test -bench .

deps:
	go get ./...

imports:
	go get golang.org/x/tools/cmd/goimports
	goimports -w .

assets: assetutil/bindata.go

assetutil/bindata.go:
	go get -f -u github.com/jteeuwen/go-bindata/...
	go-bindata -o assetutil/bindata.go -pkg assetutil assets/...

vet:
	go vet ./...

cover:
	go test -cover ./...

generate:
	go generate

all: $(TARGETS)

span-import: assets imports deps generate
	go build -o span-import cmd/span-import/main.go

span-export: assets imports deps generate
	go build -o span-export cmd/span-export/main.go

span-tag: assets imports deps generate
	go build -o span-tag cmd/span-tag/main.go

clean:
	rm -f $(TARGETS)
	rm -f span_*deb
	rm -f span-*rpm
	rm -rf ./packaging/deb/span/usr
	rm -f assetutil/bindata.go

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

cloc:
	cloc --max-file-size 1 --exclude-dir assets --exclude-dir assetutil --exclude-dir tmp --exclude-dir fixtures .

# ==== vm-based packaging ====
#
# Required, if development and deployment OS have different versions of libc.
# Examples: CentOS 6.5 has 2.12 (2010-08-03), Ubuntu 14.04 2.19 (2014-02-07).
#
# ----
#
# Initially, setup a CentOS 6.5 machine, install dependencies and git clone:
#
#     $ vagrant up
#     $ make setup
#
# To build an rpm, subsequently run:
#
#     $ make rpm-compatible
#
# If vagrant ssh runs on a port other than 2222, adjust (e.g. to port 2200):
#
#     $ make rpm-compatible PORT=2200
#
# A span-<version>-0.x86_64.rpm file should appear on your host machine, that
# has been built againts CentOS' 6.5 libc.
#
# Cleanup VM:
#
#     $ vagrant destroy --force

PORT = 2222
SSHCMD = ssh -o StrictHostKeyChecking=no -i vagrant.key vagrant@127.0.0.1 -p $(PORT)
SCPCMD = scp -o port=$(PORT) -o StrictHostKeyChecking=no -i vagrant.key

# Helper to build RPM on a RHEL6 VM, to link against glibc 2.12
vagrant.key:
	curl -sL "https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant" > vagrant.key
	chmod 0600 vagrant.key

# Don't forget to vagrant up :) - and add your public key to the guests authorized_keys
setup: vagrant.key
	$(SSHCMD) "sudo yum install -y sudo yum install http://ftp.riken.jp/Linux/fedora/epel/6/i386/epel-release-6-8.noarch.rpm"
	$(SSHCMD) "sudo yum install -y golang git rpm-build gcc-c++"
	$(SSHCMD) "mkdir -p /home/vagrant/src/github.com/miku"
	$(SSHCMD) "cd /home/vagrant/src/github.com/miku && git clone /vagrant/.git span"

rpm-compatible: vagrant.key
	$(SSHCMD) "GOPATH=/home/vagrant go get -f -u github.com/jteeuwen/go-bindata/... golang.org/x/tools/cmd/goimports"
	$(SSHCMD) "cd /home/vagrant/src/github.com/miku/span && git pull origin master && pwd && GOPATH=/home/vagrant make clean rpm"
	$(SCPCMD) vagrant@127.0.0.1:/home/vagrant/src/github.com/miku/span/*rpm .
