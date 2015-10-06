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
The tools expect a single input *file* and produce a single output *file* (stdin and stdout, respectively).

----

Why in Go?

Linux shell scripts have no native XML or JSON support, Python is a bit too
slow for the casual processing of 100M or more records, Java is a bit too
verbose - which is why we chose Go. Go comes with XML and JSON support in the
standard library, nice concurrency primitives and simple single static-binary
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
* `span-export`, intermediate schema to anything

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

Import [crossref LDJ](https://github.com/miku/siskin/blob/3d34d786f054ca153be37a514e53eea420748a8f/siskin/sources/crossref.py#L138) (with [cached members](https://github.com/miku/siskin/blob/3d34d786f054ca153be37a514e53eea420748a8f/siskin/sources/crossref.py#L224) API responses) or DeGruyter XML ([preprocessed](https://github.com/miku/siskin/blob/3d34d786f054ca153be37a514e53eea420748a8f/siskin/sources/degruyter.py#L59) into a single file):

    $ span-import -i crossref -members members.ldj crossref.ldj > crossref.is.ldj
    $ span-import -i jats degruyter.ldj > degruyter.is.ldj

Concat for convenience:

    $ cat crossref.is.ldj degruyter.is.ldj > ai.is.ldj

Export intermediate schema records to a memcache server with [memcldj](https://github.com/miku/memcldj):

    $ memcldj ai.is.ldj

Export to finc 1.3 SOLR 4 schema:

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
// Source can emit records given a reader. The channel is of type []Importer,
// to allow the source to send objects over the channel in batches for
// performance (1000 x 1000 docs vs 1000000 x 1 doc).
type Source interface {
        Iterate(io.Reader) (<-chan []Importer, error)
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

Licence
-------

* GPL
* [CLD2](https://github.com/CLD2Owners/cld2): Compact Language Detector 2, Dick Sites dsites@google.com, Apache License Version 2.0

----

TODO
----

* maybe factor out importer interface (like exporter)
* docs: add example files for each supported data format

A filtering pipeline.

The final processing step from an intermediate schema to an export format
includes various decisions.

* Should an ISIL be attached to a record or not?
* Should a record be excluded, due to an expired or deleted DOI?

Provide a middleware-ish processing interface (similar to flow, metafacture)?

    pl := Pipeline{}
    pl.Add(DOIFilter)
    pl.Add(ISILAttacher)
    pl.Add(QualityAssuranceTests)
    pl.Add(Exporter)

    err := pl.Run(input)

Done
----

* decouple batching (performance) from record stream generation (content)
* write wrappers around common inputs like XML, JSON, CSV ...
