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

Addings data sources
--------------------

The following kinds of data shapes are supported at the moment:

* A stream of XML, containing zero, one or more records, identified by an XML
tag. Moderately fast.
* Newline delimited JSON, containing zero, one or more records, one record per
line. Fast.
* Single records of arbitrary shape. Slow.

Use span, if
[metafacture](https://github.com/culturegraph/metafacture-core/wiki) or
[jq](https://stedolan.github.io/jq/) or a Python snippet are not sufficient.

Steps:

* Add a new subpackage for your format, e.g. [dummy](https://github.com/miku/span/tree/master/formats/dummy).
* Add a [struct](https://github.com/miku/span/blob/9f07e35be39c184686b05e759b4d826b1de1a905/formats/dummy/example.go#L12-L15) representing the original record (XML, JSON, bytes).
* Implement the conversion functions required, e.g. [ToIntermediateSchema](https://github.com/miku/span/blob/9f07e35be39c184686b05e759b4d826b1de1a905/formats/dummy/example.go#L17-L22)
* Add an entry into the [format map](https://github.com/miku/span/blob/f89ef0337249ba5f75d05d8a3db7c85b5c389eaa/cmd/span-import/main.go#L59) for span-import
* [Decide](https://github.com/miku/span/blob/9f07e35be39c184686b05e759b4d826b1de1a905/cmd/span-import/main.go#L202),
which kind of source this is (XML stream, newline delimited JSON, single
records, something else)
* Recompile and ship.

Ideas for span 0.2.0
--------------------

TODO:

* Do not require recompilation for mapping updates (allow various sources)
* Decouple format from source. Things like SourceID and MegaCollection are per source, not format.
* Allow loadable assets from ~/.config/span/maps, some specified location or a single JSON file.

More taggable formats:

Let formats implement a single function interface, that takes a filter value.
How to make this filter generic? Or use Tag as an adapeter?

```go
// Tag alters the document by applying a filter to it.
func (doc *Document) Tag(f Filter) error {
    // Adapt.
    record = FilterRecord{
	ISSN: doc.ISSN,
	Title: doc.CombinedTitle(),
	// ...
    }
    // Apply.
    return f.Apply(record)
}
```

DONE:

* Reuse more generic code, e.g. [parallel](http://github.com/miku/parallel)
* Make conversions a simpler with [xmlstream](https://github.com/miku/xmlstream)

Licence
-------

* GPLv3
* This project uses the Compact Language Detector 2 - [CLD2](https://github.com/CLD2Owners/cld2), Apache License Version 2.0
