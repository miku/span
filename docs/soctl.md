SOCTL 1 "JULY 2018" "Leipzig University Library" "Manuals"
=========================================================

NAME
----

soctl - a thin solr tool for finc


SYNOPSIS
--------

`soctl` COMMAND [OPTIONS] [FILE]


DESCRIPTION
-----------

The `soctl` outputs index stats and can update record labels, from ERM or
custom queries.

EXAMPLE
-------

Display index stats.

  `soctl status -server solr:8080/solr/biblio`

Pop records and apply ad-hoc fixes.

  `soctl pop -q "source_id:123" | jq -rc '.' | solrbulk

Attach a label:

  `soctl set -label DE-15 -q "issn:1234-5678"

Detach labels from a record:

  `soctl unset -label DE-* -q "source_id:123"



AUTHOR
------

Martin Czygan <martin.czygan@uni-leipzig.de>


SEE ALSO
--------

[FINC](https://finc.info), [AMSL](http://amsl.technology/), [metafacture](https://github.com/culturegraph), jq(1), xmlstarlet(1)
