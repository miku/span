# Attachments

Currently, `kbart.Holdings` can be loaded (e.g. from a file), which allows to
derive some helper indices, such as `h.SerialNumberMap` which allows to find
the relevant entry for an ISSN.

Trying to move to a tabular form, the following steps seem to be required:

* read config table (span-amsl-discovery -f ...) into memory
* (1) determine if record values (sid, collection) is in table
* (2) if so, check if there is a holding file to consider - if yes, find relevant
  entries from holding file and match agains record values (date, volume,
issue)

For (1), we can map (sid, collection) and (sid, tcid) to indices to entries.

## Exploring attachment table

```python
>>> import pandas as pd
>>> from siskin.sources.amsl import AMSLServiceTab
>>> amslfile = AMSLServiceTab().output().path
>>> names = [
    'ShardLabel',
    'ISIL',
    'SourceID',
    'TechnicalCollectionID',
    'MegaCollection',
    'HoldingsFileURI',
    'HoldingsFileLabel',
    'LinkToHoldingsFile',
    'EvaluateHoldingsFileForLibrary',
    'ContentFileURI',
    'ContentFileLabel',
    'LinkToContentFile',
    'ExternalLinkToContentFile',
    'ProductISIL',
    'DokumentURI',
    'DokumentLabel',
]
>>> df = pd.read_csv(amslfile, compression="gzip", sep="\t", names=names)
>>> df.groupby("SourceID").size().sort_values(ascending=False).head(10)
SourceID
49     239717
126     10101
17        638
26        168
94        154
55        113
48         86
0          67
105        50
68         32
dtype: int64

>>> df.groupby("ISIL").size().sort_values(ascending=False)
ISIL
DE-15                55805
DE-105               33576
DE-82                33408
DE-D275              33376
DE-15-FID            23790
DE-Brt1              16684
DE-Zwi2              11196
DE-14                 8544
DE-Ch1                8432
DE-Zi4                8139
DE-Gla1               8077
DE-540                6267
FID-BBI-DE-23         3776
DE-1156                105
DE-L242                 52
DE-D117                 47
DE-1972                 43
DE-D161                 38
DE-Bn3                  36
DE-1989                 35
DE-Mh31                 28
DE-L152                 28
DE-Kn38                 24
DE-Rs1                  17
DE-D13                  14
DE-Pl11                  9
DE-L229                  7
FID-NORD-DE-8            4
FID-MONTAN-DE-105        3
DE-197                   2
DE-L328                  2
DE-L327                  2
DE-D209a                 1
DE-520                   1
DE-Wim8                  1
DE-Wh1                   1
DE-L334                  1
DE-L330                  1
DE-185                   1
DE-25                    1
DE-291-355               1
DE-L326                  1
DE-D209                  1
DE-L325                  1
DE-B791                  1
DE-L282                  1
DE-L245                  1
DE-L228                  1
DE-D115                  1
DE-Ks17                  1
DE-Gl2                   1
DE-D174                  1
DE-D40                   1
dtype: int64
```

## FOLIO setup

* finc-config (data)
* finc-select (tenants)
* api reference: https://dev.folio.org/reference/api/

Discovery database.

* dual api access

Filters.

* join (filter and collection)
* how to develop

