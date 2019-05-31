SPAN 1 "JULY 2016" "Leipzig University Library" "Manuals"
=========================================================

NAME
----

span-import, span-tag, span-export, span-check, span-oa-filter,
span-update-labels, span-crossref-snapshot, span-local-data, span-freeze,
span-review, span-webhookd, span-hcov, span-amsl-discovery - intermediate
schema and integration tools

SYNOPSIS
--------

`span-import` [`-i` *input-format*] < *file*

`span-tag` [`-c` *config*, `-unfreeze` *file*, `-server` *url*, `-prefs` *prefs*] < *file*

`span-export` [`-o` *output-format*] < *file*

`span-check` [`-verbose`] < *file*

`span-oa-filter` [`-f` *file*] [`-fc` *file*] [`-xsid` *string*] [`-oasid` *string*] < *file*

`span-update-labels` [`-f` *file*, `-s` *separator*] < *file*

`span-crossref-snapshot` [`-x` *file*] -o *file* *file*

`span-local-data` < *file*

`span-freeze` -o *file* < *file*

`span-review` [`-server` *url*] [`-span-config` *file*] [`-c` *file*] [`-a`] [`-t`] [`-ticket` *number*]

`span-webhookd` [`-addr` *hostport*] [`-logfile` *file*] [`repo-dir` *path*] [`-span-config` *file*] [`-token` *token*] [`-trigger-path` *path*]

`span-hcov` `-f` *file* `-server` *url*

`span-amsl-discovery` `-live` *URL* [`-allow-empty`] [`-verbose`]

`span-crossref-members` [`-base` *URL*] [`-offset` *N*] [`-rows` *N*] [`-q`] [`-sleep` *duration*] [`-email` *addr*]

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
  Configuration string or path to configuration file. `span-tag` example in
  EXAMPLE for a CONFIGURATION FILE. `span-review` details in INDEX REVIEW.

`-list`
  List supported formats. `span-import`, `span-export` only.

`-verbose`
  More output. `span-check` only.

`-b` *N*
  Batch size. `span-tag`, `span-check`, `span-export`, `span-crossref-snapshot` only.

`-w` *N*
  Number of workers (defaults to CPU count). `span-tag`, `span-check`, `span-export` only.

`-cpuprofile` *pprof-file*
  Profiling. `span-import`, `span-tag`, `span-crossref-snapshot` only.

`-memprofile` *pprof-file*
  Profiling. `span-import`, `span-tag`, `span-export` only.

`-f` *file*
  File location (ISSN list or ID,ISIL). `span-oa-filter`, `span-update-labels` only.

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

`-addr` *hostport*
  Hostport to listen on. `span-webhookd` only.

`-logfile` *file*
  Logfile to log to. `span-webhookd`, `span-import` only.

`-repo-dir` *path*
  Local repo clone. `span-webhookd` only.

`-span-config` *path*
  Path to span config. `span-review`, `span-webhookd` only.

`-token` *token*
  GitLab API token. `span-webhookd` only.

`-a`
  Emit ascii table. `span-review` only.

`-t`
  Emit textile table for redmine. `span-review` only.

`-server` *url*
  Location of SOLR, including scheme, host, port and core. `span-review` only.

`-ticket` *id*
  Post review results into a Redmine ticket. `span-review` only.

`-trigger-path` *path*
  Path trigger (default "trigger"), `span-webhookd` only.

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

`-email` *addr*
  Use email address in query parameter for API etiquette, `span-crossref-members` only.

`-h`
  Show usage.

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

INDEX REVIEWS
-------------

Since 0.1.241 it is possible to run slightly automated SOLR index reviews. The
two tools are `span-review` for reviews and `span-webhookd` for automatically
running a review on commits in GitLab. These tools are experimental and might
change in the future.

Start the webhook receiver:

  `span-webhookd`

Or use the service shipped with the distribution packages.

  `servicectl span-webhookd start`

The service requires `/var/log/span-webhookd.log` to be writable by `daemon`.

The default port is 8080 (change this in SPAN CONFIG). The server listens on
all interfaces. The default URL is: `http://0.0.0.0:8080/trigger`. Enter this
URL in GitLab *settings/integrations*.

The review file location is hardcoded at the moment, `docs/review.yaml`.
Example config file:

```
# Review configuration, refs #12756.
#
# Proposed workflow:
#
# 1. Edit this file via GitLab at
# https://git.sc.uni-leipzig.de/miku/span/blob/master/docs/review.yaml. Add,
# edit or remove rules, update ticket number. If done, commit.
# 2. A trigger will run an index review based on these rules.
# 3. Find the results in your ticket, in case the ticket number was valid.

# The solr server to query, including scheme, port and collection, e.g.
# "http://localhost:8983/solr/biblio". If "auto", then the current testing solr
# server will be figured out automatically.
solr: "auto"

# The ticket number of update. Set this to "NA" or anything non-numeric to
# suppress ticket updates.
ticket: "NA"

# If set to "fail" an empty result set will be marked as failure.
# Otherwise a empty result set will - most of the time - not be considered a violation.
zero-results-policy: "fail"

# Allowed keys: [Query, Facet-Field, Value, ...] checks if all values of field
# contain only given values.
allowed-keys:
    - ["source_id:30", "format", "eBook", "ElectronicArticle"]
    - ["source_id:30", "format_de15", "Book, eBook", "Article, E-Article"]
    - ["source_id:48", "language", "German", "English"]
    - ["source_id:49", "facet_avail", "Online", "Free"]
    - ["source_id:55", "facet_avail", "Online", "Free"]

# All records: [Query, Facet-Field, Value, ...] checks if all record contain
# only the given values.
all-records:
    - ["source_id:28", "format", "ElectronicArticle"]
    - ["source_id:28", "format_de15", "Article, E-Article"]
    - ["source_id:28", "facet_avail", "Online", "Free"]
    - ["source_id:28", "access_facet", "Electronic Resources"]
    - ["source_id:28", "mega_collection", "DOAJ Directory of Open Access Journals"]
    - ["source_id:28", "finc_class_facet", "not assigned"]
    - ["source_id:30", "facet_avail", "Online", "Free"]
    - ["source_id:30", "access_facet", "Electronic Resources"]
    - ["source_id:30", "mega_collection", "SSOAR Social Science Open Access Repository"]

# MinRatio: Query, Facet-Field, Value, Ratio (Percent), checks if the given
# value appears in a given percentage of documents.
min-ratio:
    - ["source_id:49", "facet_avail", "Free", 0.8]
    - ["source_id:55", "facet_avail", "Free", 2.2]
    - ["source_id:105", "facet_avail", "Free", 0.5]

# MinCount: Query, Facet-Field, Value, Min Count. Checks, if the given value
# appears at least a fixed number of times.
min-count:
    - ["source_id:89", "facet_avail", "Free", 50]
```

SPAN CONFIG
-----------

The span config file is used by `span-review` and `span-webhookd`, since they
access various external systems: SOLR, Redmine, GitLab, Nginx. Default location
is `~/.config/span/span.json`, with `/etc/span/span.json` as fallback. The
`span-webhookd` service will not start, if no config file can be found.

```
{
  "gitlab.token": "adszuDZZ778sdsiuDsd-R4",
  "whatislive.url": "http://example.com/whatislive",
  "redmine.baseurl": "https://projects.example.com",
  "redmine.apitoken": "d41d8cd98f00b204e9800998ecf8427e",
  "port": 8080
}
```

COVERAGE REPORT
---------------

A simple coverage report can be generated with `span-hcov` tool.

```
$ span-hcov -f kbart.txt -server 10.1.1.1:8085/solr/biblio
```

This will calculate the ratio of ISSN overlap between holdings and index.

Example report (might change in the future):

```
{
  "coverage_pct": "83.29%",
  "date": "2018-09-24T14:42:46.565617857+02:00",
  "holdings": 22122,
  "holdings_file": "tmp/MFHB_ALkbart_2018-08-23.txt",
  "holdings_only_count": 3697,
  "holdings_only": [
    "0000-0600",
    "0000-3600",
    "0001-0196",
    "0001-4672",
    ...
    "8756-7113",
    "8756-8160"
  ],
  "index": 156708,
  "index_url": "http://172.18.113.7:8085/solr/biblio",
  "intersection": 18425
}
```

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

BUGS
----

Please report bugs to https://github.com/miku/span/issues.

AUTHOR
------

Martin Czygan <martin.czygan@uni-leipzig.de>

SEE ALSO
--------

[FINC](https://finc.info), [AMSL](http://amsl.technology/), [intermediate schema](https://github.com/ubleipzig/intermediateschema), [metafacture](https://github.com/culturegraph), jq(1), xmlstarlet(1)
