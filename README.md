Span
====

Span formats.

Godocs: http://godoc.org/github.com/miku/span

Formats
-------

* [CrossRef API](http://api.crossref.org/)
* [OVID](http://rzblx4.uni-regensburg.de/ezeitdata/admin/ezb_export_ovid_v01.xsd)
* Finc

Usage
-----

    $ span
    Usage: span [OPTIONS] CROSSREF.LDJ
      -b=25000: batch size
      -cpuprofile="": write cpu profile to file
      -hspec="": ISIL PATH pairs
      -members="": path to LDJ file, one member per line
      -v=false: prints current program version
      -w=4: workers

----

**Inputs**

* An input LDJ containing all crossref works metadata as [crossref.Document](https://github.com/miku/span/blob/5585dc500d82fcab9c783937d7d567fdffb71fde/crossref/document.go#L46). [Example](http://api.crossref.org/works/56).

Optionally:

* A number of XML files, containing holdings information for various institutions in [ovid format](http://rzblx4.uni-regensburg.de/ezeitdata/admin/ezb_export_ovid_v01.xsd).
* A file containing information about members, in LDJ format. [Example](http://api.crossref.org/members/56).

The [current implementation](https://github.com/miku/span/blob/318c85e649880efb02dacb75a7d5ecb010a1b979/cmd/span/main.go#L38) will not emit documents without any institiution.

One can transform the documents with the `span` tool:

    span -hspec DE-15:file.xml,DE-20:other.xml crossref.ldj

Additionally, if one has a cached file of members API responses, one can
use it as input, so the API does not need to be called at all:

    span -hspec DE-15:file.xml,DE-10:other.xml -members members.ldj crossref.ldj

The output is an LDJ in [finc.Schema](https://github.com/miku/span/blob/5585dc500d82fcab9c783937d7d567fdffb71fde/finc/schema.go#L5).
