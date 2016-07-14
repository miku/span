SPAN 1 "JULY 2016" "Leipzig University Library" "Manuals"
=========================================================

NAME
----

span-import, span-tag, span-check, span-export - intermediate schema tools

SYNOPSIS
--------

`span-import` [`-i` *input-format*] *file*

`span-tag` [`-c` *config-file*] *file*

`span-check` [`-verbose`] *file*

`span-export` [`-o` *output-format*] *file*

DESCRIPTION
-----------

The `span` tools support metadata processing by supplying commands for data conversion
to and from intermediate schema, license tagging and quality assurance.

The intermediate schema is a normalization vehicle. Spec: https://github.com/ubleipzig/intermediateschema

OPTIONS
-------

`-i` *format*
  Input format. `span-import` only.

`-o` *format*
  Output format. `span-export` only.

`-c` *config-file*
  Path to configuration file. `span-tag` only. See EXAMPLE for a CONFIGURATION FILE.

`-list`
  List support formats. `span-import`, `span-export` only.

`-verbose`
  More output. `span-import`, `span-check` only.

`-b` *N*
  Batch size (default N=20000). `span-tag`, `span-check`, `span-export` only.

`-w` *N*
  Number of workers (defaults to CPU count). `span-tag`, `span-check`, `span-export` only.

`-cpuprofile` *pprof-file*
  Profiling. `span-import`, `span-tag` only.

`-log` *log-file*
  If given log to file. `span-import` only.

`-v`
  Show version.

`-h`
  Show usage.

EXAMPLES
--------

List supported format for conversion to intermediate schema:

  `span-import -list`

Convert DOAJ dump into intermediate schema:

  `span-import -i doaj dump.ldj`

Apply licensing information from a configuration file to an intermediate schema file.

  `span-tag -c <(echo '{"DE-15": {"any": {}}})' intermediate.file`

There are a couple of content filters available: `any`, `doi`, `issn`,
`package`, `holdings`, `collection`, `source`. These content filters can be
combined with: `or`, `and` and `not`. The top level keys are the labels, that
will be injected as `x.labels` into the document, if the filter evaluates to
true.

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

Export in a SOLR schema:

  `span-export -o solr5vu3v11 intermediate.file`

FILES
-----

Assets (mostly string to string mappings) are compiled into the executable. To
change these mappings, edit the suitable file under
https://github.com/miku/span/tree/master/assets, commit and recompile.

ENVIRONMENT
-----------

`GOMAXPROCS`
  The GOMAXPROCS variable limits the number of operating system threads that can
  execute user-level Go code simultaneously.

DIAGNOSTICS
-----------

Any input error, e.g. faulty JSON, any write error, etc., will lead to an
immediate halt.

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

[FINC](https://finc.info), [AMSL](http://amsl.technology/), [intermediate schema](https://github.com/ubleipzig/intermediateschema), jq(1), xmlstarlet(1)
