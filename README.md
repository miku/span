Span
====

Span formats.

[![Build Status](https://travis-ci.org/miku/span.svg?branch=master)](https://travis-ci.org/miku/span) [![GoDoc](https://godoc.org/github.com/miku/span?status.svg)](https://godoc.org/github.com/miku/span)

The `span` command line tools aim at high performance, versatile document conversions
between a series of metadata formats.

The goal is to quickly move between input formats, such as stored API
responses, XML-ish or line-delimited JSON bibliographic data and output
formats, such as [finc](https://finc.info) [intermediate
format](https://github.com/miku/span/tree/master/schema) or formats, that can
be directly imported into SOLR or elasticsearch.

As a non-goal, the `span` tools do not care, how you obtain your input data.
The tools expect input *files* and produces *files*. Even more: Bibliographic
input data must be contained in a *single file* (even if it is 100G
in size) and the output will be a single file (stdin and stdout, respectively).

----

Why Go?

Linux shell scripts have no native XML or JSON support, Python is a bit too
slow for the casual processing of 100M or more records, Java is a bit too
verbose - which is why we choose Go. Go comes with XML and JSON support in the
standard library, nice concurrency primitives and simple single static- binary
deployments.

----

Install with

    $ go get github.com/miku/span/cmd/...

or via deb or rpm [packages](https://github.com/miku/span/releases).

Formats
-------

* [CrossRef API](http://api.crossref.org/), works and members
* JATS [Journal Archiving and Interchange Tag Set](http://jats.nlm.nih.gov/archiving/versions.html), with various flavours for JSTOR and others
* [DOAJ](http://doaj.org/) exports
* [OVID](http://rzblx4.uni-regensburg.de/ezeitdata/admin/ezb_export_ovid_v01.xsd) holdings
* [Google holdings](http://scholar.google.com/intl/en/scholar/libraries.html)
* FINC [Intermediate Format](https://github.com/miku/span/blob/master/schema/README.md)
* FINC [SOLR Schema](https://github.com/miku/span/blob/ca8583aaa9b6d5e42b758f25ade8ed3e85532841/finc/solr.go#L4)
* GENIOS Profile XML

A toolkit approach
------------------

* `span-import`, anything to intermediate schema
* `span-export`, intermediate schema to anything (finc.SolrSchema only, at the moment)

The `span-import` tool should require minimal external information (no
holdings file, etc.) and be mainly concerned with the transformation of fancy
source formats into the catch-all intermediate schema.

The `span-export` tool may include external sources to create output, e.g. holdings files.

Usage
-----

    $ span-import -h
    Usage of span-import:
      -cpuprofile="": write cpu profile to file
      -i="": input format
      -list=false: list formats
      -log="": if given log to file
      -members="": path to LDJ file, one member per line
      -v=false: prints current program version
      -verbose=false: more output
      -w=4: number of workers

    $ span-export -h
    Usage of span-export:
      -any=[]: ISIL
      -b=20000: batch size
      -cpuprofile="": write cpu profile to file
      -dump=false: dump filters and exit
      -f=[]: ISIL:/path/to/ovid.xml
      -l=[]: ISIL:/path/to/list.txt
      -list=false: list output formats
      -o="solr413": output format
      -skip=false: skip errors
      -source=[]: ISIL:SID
      -v=false: prints current program version
      -w=4: number of workers

Examples
--------

List available formats:

    $ span-import -list
    doaj
    genios
    crossref
    degruyter
    jstor

Import crossref LDJ (with cached members API responses) or DeGruyter XML (preprocessed into a single file):

    $ span-import -i crossref -members members.ldj crossref.ldj > crossref.is.ldj
    $ span-import -i jats degruyter.ldj > degruyter.is.ldj

Concat for convenience:

    $ cat crossref.is.ldj degruyter.is.ldj > ai.is.ldj

Export intermediate schema records to a memcache server with [memcldj](https://github.com/miku/memcldj):

    $ memcldj ai.is.ldj

Export to a fixed (finc) SOLR schema:

    $ span-export -o solr413 -f DE-14:DE-14.xml -f DE-15:DE-15.xml ai.is.ldj > ai.ldj

The exported `ai.ldj` contains all aggregated index record and incorporates
all holdings information. It can be indexed quickly with
[solrbulk](https://github.com/miku/solrbulk):

    $ solrbulk ai.ldj

Adding new sources
------------------

*This is work/simplification-in-progress.*

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

Channels in APIs [might](http://www.informit.com/articles/article.aspx?p=2359758) not be the optimum, though we deal with a kind of unbounded streams here.

Additionally, the the emitted objects must implement [span.Importer](https://github.com/miku/span/blob/ca8583aaa9b6d5e42b758f25ade8ed3e85532841/common.go#L22)
or [span.Batcher](https://github.com/miku/span/blob/ca8583aaa9b6d5e42b758f25ade8ed3e85532841/common.go#L16),
which is the transformation business logic:

```go
// Importer objects can be converted into an intermediate schema.
type Importer interface {
        ToIntermediateSchema() (*finc.IntermediateSchema, error)
}
```

The exporters need to implement the `finc.Exporter` interface:

```go
// ExportSchema encapsulate an export flavour. This will most likely be a
// struct with fields and methods relevant to the exported format. For the
// moment we assume, the output is JSON. If formats other than JSON are
// requested, move the marshalling into this interface.
type ExportSchema interface {
  // Convert takes an intermediate schema record to export. Returns an
  // error, if conversion failed.
  Convert(IntermediateSchema) error
  // Attach takes a list of strings (here: ISILs) and attaches them to the
  // current record.
  Attach([]string)
}
```

TODO
----

* decouple batching (performance) from record stream generation (content)
* write wrappers around common inputs like XML, JSON, CSV ...
* maybe factor out importer interface (like exporter)
* docs: add example files for each supported data format
