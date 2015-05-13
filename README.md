Span
====

Span formats.

[![Build Status](https://travis-ci.org/miku/span.svg?branch=master)](https://travis-ci.org/miku/span) [![GoDoc](https://godoc.org/github.com/miku/span?status.svg)](https://godoc.org/github.com/miku/span)

The `span` tools aim at high performance, versatile document conversions between a series of metadata formats.

Install with

    $ go get github.com/miku/span/cmd/...

or via [packages](https://github.com/miku/span/releases).

Formats
-------

* [CrossRef API](http://api.crossref.org/), works and members
* JATS [Journal Archiving and Interchange Tag Set](http://jats.nlm.nih.gov/archiving/versions.html)
* [DOAJ](http://doaj.org/) exports
* [OVID](http://rzblx4.uni-regensburg.de/ezeitdata/admin/ezb_export_ovid_v01.xsd) holdings
* [Google holdings](http://scholar.google.com/intl/en/scholar/libraries.html)
* FINC [Intermediate Format](https://github.com/miku/span/blob/master/schema/README.md)
* FINC [SOLR Schema](https://github.com/miku/span/blob/ca8583aaa9b6d5e42b758f25ade8ed3e85532841/finc/solr.go#L4)

A toolkit approach
------------------

* `span-import`, anything to intermediate schema
* `span-export`, intermediate schema to anything (finc.SolrSchema only, at the moment)

The `span-import` tool should require minimal external information (no holdings file, etc.)
and be mainly concerned with the transformation of fancy source formats into the catch-all
intermediate schema.

The `span-export` tool may include external sources to create output, e.g. holdings
files can be passed in via `-hspec`.

There are a few helpers:

* `span-hspec`, dump internal holdings data structure
* `span-gh-dump`, tabularize google holdings file

Usage
-----

    $ span-import -h
    Usage of span-import:
      -i="": input format
      -l=false: list formats
      -log="": if given log to file
      -members="": path to LDJ file, one member per line
      -v=false: prints current program version
      -w=4: number of workers

    $ span-export -h
    Usage of span-export:
      -hspec="": ISIL PATH pairs
      -v=false: prints current program version

    $ span-hspec -h
    Usage of span-hspec:
      -hspec="": ISIL PATH pairs
      -v=false: prints current program version

    $ span-gh-dump -h
    Usage of span-gh-dump:
      -v=false: prints current program version

Examples
--------

List available formats:

    $ span-import -l
    crossref
    degruyter
    jstor
    doaj

Import crossref LDJ (with cached members API responses) or DeGruyter XML (preprocessed into a single file):

    $ span-import -i crossref -members members.ldj crossref.ldj > crossref.is.ldj
    $ span-import -i jats degruyter.ldj > degruyter.is.ldj

Concat for convenience:

    $ cat crossref.is.ldj degruyter.is.ldj > ai.is.ldj

Export intermediate schema records to a memcache server with [memcldj](https://github.com/miku/memcldj):

    $ memcldj ai.is.ldj

Export to a fixed (finc) SOLR schema:

    $ span-export -hspec DE-14:DE-14.xml,DE-15:DE-15.xml ai.is.ldj > ai.ldj

The exported `ai.ldj` contains all AI record and incorporates all holdings information.
It can be indexed quickly with [solrbulk](https://github.com/miku/solrbulk):

    $ solrbulk ai.ldj

TODO
----

* clearer holdings file handling
* support for files or URLs as lookup tables (e.g. classification, languages, ...)

Adding new sources
------------------

For the moment, a new data source has to implement is the [span.Source](https://github.com/miku/span/blob/ca8583aaa9b6d5e42b758f25ade8ed3e85532841/common.go#L36) interface:

```go
// Source can emit records given a reader. What is actually returned is decided
// by the source, e.g. it may return Importer or Batcher object.
// Dealing with the various types is responsibility of the call site.
// Channel will block on slow consumers and will not drop objects.
type Source interface {
        Iterate(io.Reader) (<-chan interface{}, error)
}
```

Channels in APIs might not be the optimum, though we deal with a kind of unbounded streams here.

Additionally, the the emitted objects must implement [span.Importer](https://github.com/miku/span/blob/ca8583aaa9b6d5e42b758f25ade8ed3e85532841/common.go#L22)
or [span.Batcher](https://github.com/miku/span/blob/ca8583aaa9b6d5e42b758f25ade8ed3e85532841/common.go#L16),
which is the transformation business logic:

```go
// Importer objects can be converted into an intermediate schema.
type Importer interface {
        ToIntermediateSchema() (*finc.IntermediateSchema, error)
}
```
