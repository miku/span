SPAN 1 "JULY 2016" "Leipzig University Library" "Manuals"
=========================================================

NAME
----

span-import, span-tag, span-export, span-check, span-oa-filter,
span-update-labels, span-crossref-snapshot, span-local-data, span-freeze - intermediate
schema tools

SYNOPSIS
--------

`span-import` [`-i` *input-format*] < *file*

`span-tag` [`-c` *config*, `-unfreeze` *file*] < *file*

`span-export` [`-o` *output-format*] < *file*

`span-check` [`-verbose`] < *file*

`span-oa-filter` [`-f` *file*] [`-fc` *file*] [`-xsid` *string*] < *file*

`span-update-labels` [`-f` *file*, `-s` *separator*] < *file*

`span-crossref-snapshot` [`-x` *file*] -o *file* *file*

`span-local-data` < *file*

`span-freeze` -o *file* < *file*

DESCRIPTION
-----------

The `span` tools convert to and from an intermediate schema and support
license tagging and quality assurance.

The intermediate schema is a normalization vehicle, spec:
https://github.com/ubleipzig/intermediateschema

OPTIONS
-------

`-i` *format*
  Input format. `span-import` only.

`-o` *format*
  Output format or file. `span-export`, `span-freeze`, `span-crossref-snapshot` only.

`-c` *config-string* or *config-file*
  Configuration string or path to configuration file. `span-tag` only. See
  EXAMPLE for a CONFIGURATION FILE.

`-list`
  List support formats. `span-import`, `span-export` only.

`-verbose`
  More output. `span-check` only.

`-b` *N*
  Batch size. `span-tag`, `span-check`, `span-export`, `span-crossref-snapshot` only.

`-w` *N*
  Number of workers (defaults to CPU count). `span-tag`, `span-check`, `span-export` only.

`-cpuprofile` *pprof-file*
  Profiling. `span-import`, `span-tag`, `span-crossref-snapshot` only.

`-f` *file*
  File location (ISSN list or ID,ISIL). `span-oa-filter`, `span-update-labels` only.

`-fc` *file*
  File in AMSL FreeContent API format about sources, collections and their OA status, `span-oa-filter` only.

`-s` *sep*
  Field separator. `span-update-labels` only.

`-unfreeze` *file*
  Take a file created with `span-freeze` and use it instead of a filterconfig. `span-tag` only.

`-v`
  Show version.

`-x` *file*
  Filename to DOI to exclude, one per line. `span-crossref-snapshot` only.

`-xsid` *sid*
  Do not apply processing on a given source id. `span-oa-filter` only.

`-z`
  Input is gzip compressed. `span-crossref-snapshot` only.

`-h`
  Show usage.

EXAMPLES
--------

List supported format for conversion to intermediate schema:

  `span-import -list`

Convert DOAJ dump into intermediate schema:

  `span-import -i doaj dump.ldj`

Apply licensing information from a string with streaming input.

  `cat intermediate.file | span-tag -c '{"DE-15": {"any": {}}}'`

Apply licensing information from a configuration file to an intermediate schema file.

  `span-tag -c <(echo '{"DE-15": {"any": {}}})' intermediate.file`

There are a couple of content filters available: `any`, `doi`, `issn`,
`package`, `holdings`, `collection`, `source` and `subject`. These content
filters can be combined with: `or`, `and` and `not`. The configuration can be
seen as an expression forest. The top level keys are the labels, that will be
injected as `x.labels` into the document, if the filter below the key evaluates
to true.

The holdings filter configuration can include a list of URLs. As of 0.1.221 the
the "urls" value supports the `file://` scheme as well.

More complex example for a configuration file:

    {
      "DE-14": {
        "or": [
          {
            "and": [
              {
                "source": [
                  "55"
                ]
              },
              {
                "holdings": {
                  "urls": [
                    "http://www.jstor.org/kbart/collections/asii",
                    "http://www.jstor.org/kbart/collections/as"
                  ]
                }
              }
            ]
          },
          {
            "and": [
              {
                "source": [
                  "49"
                ]
              },
              {
                "holdings": {
                  "urls": [
                    "https://example.com/KBART_DE14",
                    "https://example.com/KBART_FREEJOURNALS"
                  ]
                }
              },
              {
                "collection": [
                  "Turkish Family Physicans Association (CrossRef)",
                  "Helminthological Society (CrossRef)",
                  "International Association of Physical Chemists (IAPC) (CrossRef)",
                  "The Society for Antibacterial and Antifungal Agents, Japan (CrossRef)",
                  "Fundacao CECIERJ (CrossRef)"
                ]
              }
            ]
          }
        ]
      }
    }

  `span-tag -c config.json intermediate.file`

List available export formats:

  `span-export -list`

Export to a SOLR schema:

  `span-export -o solr5vu3 intermediate.file`

Export to Metafacture formeta:

  `span-export -o formeta intermediate.file`

Set OA flag (via KBART-ish file):

  `echo '{"rft.issn": ["1234-1234"], "rft.date": "2000-01-01"}' | span-oa-filter -f <(echo $'online_identifier\n1234-1234')`

Update labels:

  `echo '{"finc.record_id": "1"}' | span-update-labels -f <(echo '1,X,Y')`

Create a snapshot of crossref works API message items:

  `span-crossref-snapshot -o snapshot.ldj.gz messages.ldj.gz`

The `messages.ldj.gz` must contain only the message portion of an crossref API
response - one per line - for example:

  `curl -sL goo.gl/Cq34Bd | jq .message`

Given an intermediate schema file, extract record id, source id, doi and labels
(ISIL). Can be fed into groupcover(1) for deduplication.

  `span-local-data < input.ldj > output.tsv`

Example output:

  `ai-49-aHR0cDovL2R4LmRva...    49    10.2307/3102818    DE-15-FID    DE-Ch1    DE-105`

Freezing a filterconfig
-----------------------

When given a single file containing a number of URLs, it is required to keep
both the file and all URLs it contains for a given point in time. The
`span-freeze` tool is generic, in that it does not assume any format. It will
create a zip file with the following layout:

    /blob
    /mapping.json
    /files/<hash>
    /files/<hash>
    ...

Where `blob` is the original file containing URLs, `mapping.json` is a JSON document
containing a SHA1 to URL mapping and the `files` directory contains all
responses, with the filename being the SHA1 of the URL.

Example usage:

  `span-freeze -o frozen.zip < filterconfig.json`

Example for thawing a configuration. The zip file will be decompressed into a
temporary location and the configuration is modified accordingly before tagging
starts.

  `span-tag -unfreeze frozen.zip < intermediate.file`

FILES
-----

Assets (mostly string to string mappings) are compiled into the executable. To
change these mappings, edit the suitable file under
https://github.com/miku/span/tree/master/assets, commit and recompile.

DIAGNOSTICS
-----------

Any error (like faulty JSON, IO errors, ...) will lead to an immediate halt.

To debug a holdings filter, set `verbose` to `true` to see rejected records and rejection reason:

    {
      "DE-14": {
        "holdings": {
          "verbose": true,
          "urls": [
            "http://www.jstor.org/kbart/collections/asii",
            "http://www.jstor.org/kbart/collections/as"
          ]
        }
      }
    }

Example debugging output, record rejected because it's outside licence coverage:

    2016/07/14 14:29:45 {
        "document": {
            ...
            "finc.record_id": "ai-55-aHR0cDovL3d3dy5qc3Rvci5vcmcvc3RhYmxlLzEwLjE0MzIxL3JoZXRwdWJsYWZmYS4xOC4xLjAxNjE",
            ...
            "rft.atitle": "Review: Depression: A Public Feeling",
            ...
            "rft.issn": [
                "1094-8392",
                "1534-5238"
            ],
            "rft.date": "2015-04-01",
            "doi": "10.14321/rhetpublaffa.18.1.0161",
            ...
        },
        "err": "after coverage interval",
        "issn": "1534-5238",
        "license": {
            "Begin": {
                "Date": "1998-04-01",
                "Volume": "1",
                "Issue": "1"
            },
            "End": {
                "Date": "2012-12-01",
                "Volume": "15",
                "Issue": "4"
            },
            "Embargo": -126144000000000000,
            "EmbargoDisallowEarlier": false
        }
    }


BUGS
----

Please report bugs to https://github.com/miku/span/issues.

AUTHOR
------

Martin Czygan <martin.czygan@uni-leipzig.de>

SEE ALSO
--------

[FINC](https://finc.info), [AMSL](http://amsl.technology/), [intermediate schema](https://github.com/ubleipzig/intermediateschema), [metafacture](https://github.com/culturegraph), jq(1), xmlstarlet(1)
