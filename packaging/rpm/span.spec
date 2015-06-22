Summary:    Library data conversions.
Name:       span
Version:    0.1.38
Release:    0
License:    MIT
BuildArch:  x86_64
BuildRoot:  %{_tmppath}/%{name}-build
Group:      System/Base
Vendor:     Leipzig University Library, https://www.ub.uni-leipzig.de
URL:        https://github.com/miku/span

%description

Library data conversions.

%prep
# the set up macro unpacks the source bundle and changes in to the represented by
# %{name} which in this case would be my_maintenance_scripts. So your source bundle
# needs to have a top level directory inside called my_maintenance _scripts
# %setup -n %{name}

%build
# this section is empty for this example as we're not actually building anything

%install
# create directories where the files will be located
mkdir -p $RPM_BUILD_ROOT/usr/local/sbin

# put the files in to the relevant directories.
# the argument on -m is the permissions expressed as octal. (See chmod man page for details.)
install -m 755 span-export $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 span-gh-dump $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 span-import $RPM_BUILD_ROOT/usr/local/sbin


%post
# the post section is where you can run commands after the rpm is installed.
# insserv /etc/init.d/my_maintenance

%clean
rm -rf $RPM_BUILD_ROOT
rm -rf %{_tmppath}/%{name}
rm -rf %{_topdir}/BUILD/%{name}

# list files owned by the package here
%files
%defattr(-,root,root)
/usr/local/sbin/span-export
/usr/local/sbin/span-gh-dump
/usr/local/sbin/span-import


%changelog
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
