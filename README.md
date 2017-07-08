Span
====

The span tools convert to and from an intermediate schema and support license
tagging and quality assurance.

The intermediate schema is a normalization vehicle, spec: https://github.com/ubleipzig/intermediateschema

----

Install with

    $ go get github.com/miku/span/cmd/...

or via deb or rpm [packages](https://github.com/miku/span/releases).

Formats
-------

* [CrossRef API](http://api.crossref.org/), works and members
* JATS [Journal Archiving and Interchange Tag Set](http://jats.nlm.nih.gov/archiving/versions.html), with various flavours for JSTOR and others
* [DOAJ](http://doaj.org/) exports
* FINC [Intermediate Format](https://github.com/ubleipzig/intermediateschema)
* Various FINC [SOLR Schema](https://github.com/finc/index/blob/master/schema.xml)
* GENIOS Profile XML
* Elsevier Transport
* Thieme TM Style
* [Formeta](https://github.com/culturegraph)
* IEEE IDAMS Exchange V2.0.0

Also:

* [KBART](http://www.uksg.org/KBART)

Ideas for span 0.2.0
--------------------

TODO:

* Do not require recompilation for mapping updates (allow various sources)
* Decouple format from source. Things like SourceID and MegaCollection are per source, not format.

DONE:

* Reuse more generic code, e.g. [parallel](http://github.com/miku/parallel)
* Make conversions a simpler with [xmlstream](https://github.com/miku/xmlstream)

Licence
-------

* GPLv3
* This project uses the Compact Language Detector 2 - [CLD2](https://github.com/CLD2Owners/cld2), Apache License Version 2.0
