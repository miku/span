Summary:    Library data tools.
Name:       span
Version:    0.1.308
Release:    0
License:    GPL
ExclusiveArch:  x86_64
BuildRoot:  %{_tmppath}/%{name}-build
Group:      System/Base
Vendor:     Leipzig University Library, https://www.ub.uni-leipzig.de
URL:        https://github.com/miku/span

%description

Library data tools.

%prep

%build

%pre

%install

mkdir -p $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-amsl-discovery $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-check $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-compare $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-crossref-members $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-crossref-snapshot $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-export $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-freeze $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-hcov $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-import $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-local-data $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-oa-filter $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-redact $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-report $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-review $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-tag $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-update-labels $RPM_BUILD_ROOT/usr/sbin
install -m 755 span-webhookd $RPM_BUILD_ROOT/usr/sbin

mkdir -p $RPM_BUILD_ROOT/usr/local/share/man/man1
install -m 644 span.1 $RPM_BUILD_ROOT/usr/local/share/man/man1/span.1

mkdir -p $RPM_BUILD_ROOT/usr/lib/systemd/system
install -m 644 span-webhookd.service $RPM_BUILD_ROOT/usr/lib/systemd/system/span-webhookd.service

mkdir -p $RPM_BUILD_ROOT/var/log
touch $RPM_BUILD_ROOT/var/log/span-webhookd.log

%post

%clean
rm -rf $RPM_BUILD_ROOT
rm -rf %{_tmppath}/%{name}
rm -rf %{_topdir}/BUILD/%{name}

%files
%defattr(-,root,root)

/usr/lib/systemd/system/span-webhookd.service
/usr/local/share/man/man1/span.1
/usr/sbin/span-amsl-discovery
/usr/sbin/span-check
/usr/sbin/span-compare
/usr/sbin/span-crossref-members
/usr/sbin/span-crossref-snapshot
/usr/sbin/span-export
/usr/sbin/span-freeze
/usr/sbin/span-hcov
/usr/sbin/span-import
/usr/sbin/span-local-data
/usr/sbin/span-oa-filter
/usr/sbin/span-redact
/usr/sbin/span-report
/usr/sbin/span-review
/usr/sbin/span-tag
/usr/sbin/span-update-labels
/usr/sbin/span-webhookd

%attr(0644, daemon, daemon) /var/log/span-webhookd.log

%changelog

* Mon Feb 18 2019 Martin Czygan
- 0.1.281 release
- replace span-amsl with span-amsl-discovery

* Wed Jan 23 2019 Martin Czygan
- 0.1.273 release
- add span-amsl api helper

* Tue Jul 10 2018 Martin Czygan
- 0.1.240 release
- include span-review, span-webhookd for index tests

* Wed Feb 17 2016 Martin Czygan
- 0.1.60 release
- first appearance of span-tag, ISIL tagger

* Mon Nov 2 2015 Martin Czygan
- 0.1.53 release
- span-import: sort format name output
- thieme: add subject
- thieme: rework XML parsing
- exporter: add more words to author blacklist
- export: strip Index from author field, refs #6326
- add simple test for TestFromJSONSize
- export: fix date and rawdate, refs #6266
- span-export -list sorted output formats
- embed and reuse export structs
- update LICENCE to GPL
- span-export: allow DOI filter per ISIL as well
- genios: add two more document attributes

* Fri Aug 14 2015 Martin Czygan
- 0.1.51 release
- no new features, just internal refactoring
- XML and JSON sources are now simpler to get started with FromJSON, FromXML
- slight performance gains

* Tue Aug 11 2015 Martin Czygan
- 0.1.50 release
- use a pre-script to purge affected artifacts

* Sat Aug 1 2015 Martin Czygan
- 0.1.48 release
- add -doi-blacklist flag

* Mon Jul 6 2015 Martin Czygan
- 0.1.41 release
- much faster language detection with cld2 (libc sensible)

* Sat Jun 6 2015 Martin Czygan
- 0.1.36 release
- add genios/gbi support

* Mon Jun 1 2015 Martin Czygan
- 0.1.35 release
- initial support for multiple exporters

* Sun Mar 15 2015 Martin Czygan
- 0.1.11 release
- added intermediate schema to the repo

* Thu Feb 19 2015 Martin Czygan
- 0.1.8 release
- import/export

* Thu Feb 19 2015 Martin Czygan
- 0.1.7 release
- first appearance of an intermediate format

* Wed Feb 11 2015 Martin Czygan
- 0.1.2 release
- initial release
