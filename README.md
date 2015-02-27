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

Usage
-----

    $ span
    Usage: span [OPTIONS] CROSSREF.LDJ
      -allow-empty-institutions=false: keep records, even if no institutions is using it
      -b=25000: batch size
      -cpuprofile="": write cpu profile to file
      -hspec="": ISIL PATH pairs
      -hspec-export=false: export a single combined holdings map as JSON
      -ignore=false: skip broken input record
      -members="": path to LDJ file, one member per line
      -v=false: prints current program version
      -verbose=false: print debug messages
      -w=4: workers

Inputs and Outputs
------------------

The `span` tools recognizes the following inputs at the moment:

* An input LDJ containing all crossref works metadata, one [crossref.Document](https://github.com/miku/span/blob/5585dc500d82fcab9c783937d7d567fdffb71fde/crossref/document.go#L46) per line. [Example API response](http://api.crossref.org/works/56). The [CrossrefItems](https://github.com/miku/siskin/blob/75bd2e51de9a38c9c6b5fd9dd611f1a23c866cc2/siskin/sources/crossref.py#L126) task creates such an output.

And, optionally:

* A number of XML files, containing holdings information for various institutions in [OVID](http://rzblx4.uni-regensburg.de/ezeitdata/admin/ezb_export_ovid_v01.xsd) format.
* A file containing information about [members](https://github.com/miku/span/blob/aa59d6468bad530fbf680c529e341b76e033386c/crossref/api.go#L23), in LDJ format. [Example API response](http://api.crossref.org/members/56). The [CrossrefGenericItems](https://github.com/miku/siskin/blob/75bd2e51de9a38c9c6b5fd9dd611f1a23c866cc2/siskin/sources/crossref.py#L331) can create such an output.

Example usage with two institutional holding files:

    $ span -hspec DE-15:file.xml,DE-20:other.xml crossref.ldj

Additionally, if one has a cached file of members API responses, one can
use it as input. This way the API does not need to be called at all:

    $ span -hspec DE-15:file.xml,DE-10:other.xml -members members.ldj crossref.ldj

The output is an LDJ in [finc.SolrSchema](https://github.com/miku/span/blob/aa59d6468bad530fbf680c529e341b76e033386c/finc/schema.go#L5),
which can be indexed into SOLR either via JSON update URL or with tools like [solrbulk](https://github.com/miku/solrbulk).

Notes
-----

Two separate problems: wrapping formats and processing them. Wrapping formats could be factored out into an own package.

The command line interface should accept input and output types:

    $ span -f crossref -t finc.schema -hspec ... -members ... crossref.ldj > output.ldj
    $ span -f jats -t finc.schema in.xml > output.ldj

The input type must have a To[OutputSchema] method. Implement conversions as needed.
