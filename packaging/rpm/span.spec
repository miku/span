Summary:    Library data conversions.
Name:       span
Version:    0.1.113
Release:    0
License:    GPL
BuildArch:  x86_64
BuildRoot:  %{_tmppath}/%{name}-build
Group:      System/Base
Vendor:     Leipzig University Library, https://www.ub.uni-leipzig.de
URL:        https://github.com/miku/span

%description

Library data format conversions.

%prep

%build

%pre

%install
mkdir -p $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 span-check $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 span-export $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 span-import $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 span-redact $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 span-tag $RPM_BUILD_ROOT/usr/local/sbin

mkdir -p $RPM_BUILD_ROOT/usr/local/share/man/man1
install -m 644 span.1 $RPM_BUILD_ROOT/usr/local/share/man/man1/span.1

%post

%clean
rm -rf $RPM_BUILD_ROOT
rm -rf %{_tmppath}/%{name}
rm -rf %{_topdir}/BUILD/%{name}

%files
%defattr(-,root,root)

/usr/local/sbin/span-check
/usr/local/sbin/span-export
/usr/local/sbin/span-import
/usr/local/sbin/span-redact
/usr/local/sbin/span-tag
/usr/local/share/man/man1/span.1

%changelog

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
