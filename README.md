Span
====

Span formats.

[![Build Status](https://travis-ci.org/miku/span.svg?branch=master)](https://travis-ci.org/miku/span) [![GoDoc](https://godoc.org/github.com/miku/span?status.svg)](https://godoc.org/github.com/miku/span)

The `span` tools aims at high performance, versatile document conversions between a series of metadata formats.

Formats
-------

* [CrossRef API](http://api.crossref.org/), works and members
* [OVID](http://rzblx4.uni-regensburg.de/ezeitdata/admin/ezb_export_ovid_v01.xsd) holdings
* FINC [Intermediate Format](https://github.com/miku/span/blob/4baf2a67fb057ac37edc2f12f05ece7b93190373/finc/schema.go#L61)
* FINC [SOLR Schema](https://github.com/miku/span/blob/4baf2a67fb057ac37edc2f12f05ece7b93190373/finc/schema.go#L5)
* JATS [Journal Archiving and Interchange Tag Set](http://jats.nlm.nih.gov/archiving/versions.html)

A little toolkit
----------------

* span-import, anything to intermediate schema
* span-export, intermediate schema to finc.SolrSchema

The `span-import` tool should require minimal external information (no holdings file, etc.)
and be mainly concerned with the transformation of fancy source formats into the catch-all
intermediate schema.

The `span-export` tool may include external sources to create output, e.g. holdings
files can be passed in via `-hspec`.

And has a few helpers:

* span-hspec, dump internal holdings data structure
* span-gh-dump, tabularize google holdings file

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
    jats

Import crossref LDJ (with cached members API responses):

    $ span-import -i crossref -members fixture/members.ldj fixtures/crossref.ldj > crossref.is.ldj

Import degruyter XML:

    $ span-import -i jats fixtures/degruyter.ldj > degruyter.is.ldj

Various intermediate schema files may be concatenated for convenience:

    $ cat crossref.is.ldj degruyter.is.ldj > ai.is.ldj

----

Export to a fixed (finc) SOLR schema:

    $ span-export -hspec DE-14:DE-14.xml,DE-15:DE-15.xml ai.is.ldj >  ai.ldj

The exported `ai.ldj` contains all AI record and incorporates all holdings information.
It can be indexed quickly with [solrbulk](https://github.com/miku/solrbulk):

    $ solrbulk ai.ldj

----

Export intermediate schema records to a memcache server with [memcldj](https://github.com/miku/memcldj):

    $ memcldj ai.is.ldj

Adding new sources
------------------

For the moment, a new data source has to implement is the `span.Source` interface:

    // Source can emit records given a reader. What is actually returned is decided
    // by the source, e.g. it may return Converters or Batchers. Dealing with the
    // various types is responsibility of the call site.
    type Source interface {
            Iterate(io.Reader) (<-chan interface{}, error)
    }

Channels in APIs might not be the optimum, though we deal with a kind of unbounded streams here.

Additionally, the the emitted objects must implement `span.Importer` or `span.Batcher`,
which is the transformation business logic:

    // Importer objects can be converted into an intermediate schema.
    type Importer interface {
            ToIntermediateSchema() (*finc.IntermediateSchema, error)
    }
