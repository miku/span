Span
====

Install with

    $ go get github.com/miku/span/cmd/...

or via deb or rpm [packages](https://github.com/miku/span/releases).

Formats
-------

* [CrossRef API](http://api.crossref.org/), works and members
* JATS [Journal Archiving and Interchange Tag Set](http://jats.nlm.nih.gov/archiving/versions.html), with various flavours for JSTOR and others
* [DOAJ](http://doaj.org/) exports
* FINC [Intermediate Format](https://github.com/miku/span/blob/master/schema/README.md)
* Various FINC [SOLR Schema](https://github.com/miku/span/blob/ca8583aaa9b6d5e42b758f25ade8ed3e85532841/finc/solr.go#L4)
* GENIOS Profile XML
* Elsevier Transport
* Thieme TM Style

Also:

* [KBART](http://www.uksg.org/KBART)
* [OVID](http://rzblx4.uni-regensburg.de/ezeitdata/admin/ezb_export_ovid_v01.xsd) holdings
* [Google holdings](http://scholar.google.com/intl/en/scholar/libraries.html)

TODO
----

* Decouple format from source. Things like SourceID and MegaCollection are per source, not format.

Licence
-------

* GPLv3
* This project uses the Compact Language Detector 2 - [CLD2](https://github.com/CLD2Owners/cld2), Apache License Version 2.0
