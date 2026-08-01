SPAN 1 "JULY 2016" "Leipzig University Library" "Manuals"
=========================================================

NAME
----

span-import, span-tag, span-export, span-oa-filter,
span-update-labels, span-crossref-snapshot, span-local-data, span-freeze,
span-amsl-discovery - intermediate
schema and integration tools

SYNOPSIS
--------

`span-import` [`-i` *input-format*] < *file*

`span-tag` [`-c` *config*, `-unfreeze` *file*, `-server` *url*, `-prefs` *prefs*] < *file*

`span-tagger` [`-db` *file*, `-f`, `-v`, `-debug`] < *file*

`span-export` [`-o` *output-format*] < *file*

`span-oa-filter` [`-f` *file*] [`-fc` *file*] [`-xsid` *string*] [`-oasid` *string*] < *file*

`span-update-labels` [`-f` *file*, `-s` *separator*] < *file*

`span-crossref-snapshot` [`-x` *file*] [`-S` *SIZE*] -o *file* *file*

`span-local-data` < *file*

`span-freeze` -o *file* < *file*

`span-amsl-discovery` `-live` *URL* [`-allow-empty`] [`-verbose`]

`span-crossref-members` [`-base` *URL*] [`-offset` *N*] [`-rows` *N*] [`-q`] [`-sleep` *duration*]

`span-crossref-sync` [`-P` *prefix*] [`-i` *interval] [`-p` *compress-program*] [`-s` *date*] [`-e` *date*] [`-E` *numerrors*]


DESCRIPTION
-----------

The `span` tools convert to and from an intermediate schema and support
license tagging and quality assurance.

The intermediate schema is a normalization vehicle, spec:
https://github.com/ubleipzig/intermediateschema

OPTIONS
-------

This section is correct, but incomplete. Consult `-h` for further flags.

`-i` *format*
  Input format. `span-import` only.

`-o` *format*
  Output format or file. `span-export`, `span-freeze`, `span-crossref-snapshot` only.

`-c` *config-string* or *config-file*
  Configuration string or path to configuration file. `span-tag` example in
  EXAMPLE for a CONFIGURATION FILE.

`-list`
  List supported formats. `span-import`, `span-export` only.

`-verbose`
  More output. `span-import` only.

`-b` *N*
  Batch size. `span-tag`, `span-import`, `span-export`, `span-crossref-snapshot` only.

`-w` *N*
  Number of workers (defaults to CPU count). `span-tag`, `span-export` only.

`-cpuprofile` *pprof-file*
  Profiling. `span-import`, `span-tag`, `span-crossref-snapshot` only.

`-memprofile` *pprof-file*
  Profiling. `span-import`, `span-tag`, `span-export` only.

`-f` *file*
  File location (ISSN list or ID,ISIL). `span-oa-filter`, `span-update-labels` only.

`-f`
  Flatten output to table. `span-amsl-discovery` only.

`-fc` *file*
  File in AMSL FreeContent API format about sources, collections and their OA status, `span-oa-filter` only.

`-s` *sep*
  Field separator. `span-update-labels` only.

`-unfreeze` *file*
  Take a file created with `span-freeze` and use it instead of a filterconfig. `span-tag` only.

`-v` or `-version`
  Show version.

`-x` *file*
  Filename to DOI to exclude, one per line. `span-crossref-snapshot` only.

`-xsid` *sid*
  Do not apply processing on a given source id. `span-oa-filter` only.

`-oasid` *sid*
  Set `x.oa` to true for all records of a given source id. `span-oa-filter` only.

`-z`
  Input is gzip compressed. `span-crossref-snapshot` only.

`-logfile` *file*
  Logfile to log to. `span-import` only.

`-base` *url*
  API base URL (default "http://api.crossref.org/members"), `span-crossref-members` only.

`-offset` *N*
  Offset to start fetching data from (default 0), `span-crossref-members` only.

`-rows` *N*
  Rows to fetch (default 20), `span-crossref-members` only.

`-sleep` *duration*
  Sleep between requests (default 1s), `span-crossref-members` only.

`-q`
  Suppress logging output, `span-crossref-members` only.

`-h`
  Show usage.

`-E` *numerrors*
  Number of errors to tolerate during processing. `span-crossref-snapshot` only.

`-S` *size*
  Buffer size, passed to sort. `span-crossref-snapshot` only.

EXAMPLES
--------

List supported formats for conversion to intermediate schema:

  `span-import -list`

Convert DOAJ OAI harvest to intermediate schema:

  `span-import -i doaj-oai harvest.xml`

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

Update labels, for example after a deduplication run with groupcover(1):

  `echo '{"finc.id": "1"}' | span-update-labels -f <(echo '1,X,Y')`

Create a snapshot of crossref works API message items -- more details in https://git.io/fjeih:

  `span-crossref-snapshot -o snapshot.ldj.gz messages.ldj.gz`

The `messages.ldj.gz` must contain only the message portion of an crossref API
response - one per line - for example:

  `curl -sL goo.gl/Cq34Bd | jq .message`

Given an intermediate schema file, extract id, source id, doi and labels
(ISIL). Can be fed into groupcover(1) for deduplication.

  `span-local-data < input.ldj > output.tsv`

Example output:

  `ai-49-aHR0cDovL2R4LmRva...    49    10.2307/3102818    DE-15-FID    DE-Ch1    DE-105`

Freezing a filterconfig
-----------------------

When given a single file containing a number of URLs, it is required to keep
both the file and all URLs it contains for a given point in time (#12021). The
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

The freeze tool is generic, albeit of limited utility:

  `curl -sL https://www.heise.de | span-freeze -b -o heise.zip`

NEXT ITERATION TAGGING
----------------------

In order to simplify processing, we try to get rid of the flexible tree
structure (of the filterconfig) and use a tabular approach turning AMSL into an
sqlite3 database.

  `span-amsl-discovery -db amsl.db -live https://live.example.technology`

A new program, called `span-tagger` for the moment, can take this DB and use it
to attach records based on the data in both the database and linked holding
files.

The queries results are cached, otherwise the process would be too slow for
millions of records. Referenced (holding) files are downloaded into
`$HOME/.cache/span/` (or whatever your `XDG_CACHE_HOME` is) on the fly. The
various cases are condensed into a single `switch` statement.

  `span-tagger -db amsl.db < input.is > output.is`

Similar to `span-tag`, we can let the data flow into the index through pipes.

  `taskcat AIIntermediateSchema | span-tagger -db amsl.db | span-export | solrbulk -server ...`

FILES
-----

Assets (mostly string to string mappings) are compiled into the executable. To
change these mappings, edit the suitable file under
https://github.com/miku/span/tree/master/assets, commit and recompile.

DIAGNOSTICS
-----------

Any error (like faulty JSON, IO errors, ...) will lead to an immediate halt.
The packages might contain executables in test, that are not mentioned at all
in this man page.

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
            "finc.id": "ai-55-aHR0cDovL3d3dy5qc3Rvci5vcmcvc3RhYmxlLzEwLjE0MzIxL3JoZXRwdWJsYWZmYS4xOC4xLjAxNjE",
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

AMSL DISCOVERY API COMPAT
-------------------------

In December 2018, the AMSL discovery API, required for licensing via span-tag,
has been shut down. In order to not have to rewrite too much code at this
point, we rebuild a discovery-like response from the existing endpoints:
*metadata_usage*, *holdingsfiles*, *contentfiles* and the new
*holdings_file_concat*.

At the moment (Feb 2019), the following command writes a discovery API like
JSON response to stdout:

`span-amsl-discovery -live https://live.example.technology`

A tabular output of the API can be generated with `-f`, like:

`span-amsl-discovery -f live https://live.example.technology | head -10 | cut -f 1-5`

    UBL-main        DE-1972 0       sid-0-col-zdb176dch     Digital Concert Hall
    UBL-main        DE-Mh31 0       sid-0-col-zdb176dch     Digital Concert Hall
    UBL-main        DE-105  0       lfer    Lizenzfreie Online-Ressourcen
    UBL-main        DE-14   0       lfer    Lizenzfreie Online-Ressourcen
    UBL-main        DE-15   0       lfer    Lizenzfreie Online-Ressourcen
    UBL-main        DE-15-FID       0       lfer    Lizenzfreie Online-Ressourcen
    UBL-main        DE-1972 0       lfer    Lizenzfreie Online-Ressourcen
    UBL-main        DE-Bn3  0       lfer    Lizenzfreie Online-Ressourcen
    UBL-main        DE-Brt1 0       lfer    Lizenzfreie Online-Ressourcen
    UBL-main        DE-Ch1  0       lfer    Lizenzfreie Online-Ressourcen


DEDUPLICATION AGAINST SOLR
--------------------------

Since 0.1.285, preliminary support for deduplication (DOI) against SOLR to shorten time-to-index. Basically:

    $ cat file.is | span-tag -unfreeze $(taskoutput AMSLFilterConfigFreeze) -server example.com/solr/biblio -w 64 -b 2000 -verbose > tagged.is

This will take an untagged intermediate schema file, attach all ISIL according
to config (AMSL) and post-process the document by looking up the DOI in the
given index, checking whether we have a higher prio source for a document and
ISIL - if so, drop the label, then serialize.

A hacky way around the fact, that SOLR only supports single document updates, if *all* fields are stored:

1. Drop the source, collection or whatever set from the index.
2. Find the associated intermediate schema files, run span-tag ... -server ... and span-export.
3. Reindex with solrbulk(1).

If we could generate smaller updates (daily, weekly) per source (or
collection), then a live-updater could be feasible, albeit generating extra
load on server (https://i.imgur.com/fkQNGIr.png).

INDEXING PIPELINES
------------------

Two composable pipelines feed the same live SOLR, wired up as make targets (see
the Makefile for the tunable variables). Both need solrbulk(1) on PATH and a
filterconfig, supplied as a frozen zip (`FILTERCONFIG`) or fetched from FOLIO
via the `OKAPI_URL` and `OKAPI_TOKEN` environment variables.

The *fast lane* harvests the most recent crossref slice and upserts it into the
index. It is additive only and runs often, e.g. as a daily cron:

    $ make fast-lane

which is roughly:

    $ span-crossref-sync -c CACHE -P PREFIX -i d -s DATE -e DATE -q
    $ span-crossref-fastproc CACHE/PREFIX-index-DATE-DATE.json.zst -o - | solrbulk -server SOLR

Because crossref is append-mostly, upserting the daily slice keeps the index
current without a full rebuild. The fast lane never deletes; disappearances are
reconciled by the full lane.

The *full lane* deduplicates the whole cached corpus to the latest version per
DOI, reindexes it, then removes records the reindex did not touch. It runs
periodically, e.g. weekly:

    $ make full-lane

which is roughly:

    $ span-crossref-sync -c CACHE -P PREFIX -i d -s SINCE -e UNTIL -q
    $ span-crossref-fast-snapshot -o CACHE/snapshot.json.zst CACHE/PREFIX-index-*.json.zst
    $ span-crossref-fastproc CACHE/snapshot.json.zst -o - | solrbulk -server SOLR
    $ span-index cleanup -s SOLR --sid 49 --until PASS_START | sh

The final step is the garbage collector the fast lane cannot be: after a full
pass every live crossref record has a fresh `last_indexed`, so anything older
than the pass start is stale. `span-index cleanup` only prints the
delete-by-query (scoped to the crossref source id and bounded by time) - review
it, then pipe to sh to execute. Incremental upserts and this `last_indexed`
sweep are mutually exclusive: the sweep is only correct after a pass that
touches every live record, which is why deletes belong to the full lane.

Override any variable on the command line, e.g.:

    $ make fast-lane SOLR=http://10.0.0.1:8983/solr/biblio DATE=2026-07-31
    $ make full-lane FILTERCONFIG=/etc/span/filterconfig.zip

BUGS
----

Please report bugs to https://github.com/miku/span/issues.

AUTHOR
------

Martin Czygan <martin.czygan@uni-leipzig.de>

SEE ALSO
--------

[FINC](https://finc.info), [AMSL](http://amsl.technology/), [intermediate schema](https://github.com/ubleipzig/intermediateschema), [metafacture](https://github.com/culturegraph), jq(1), xmlstarlet(1)
