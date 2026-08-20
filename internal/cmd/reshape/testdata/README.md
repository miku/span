# Format golden test data

One directory per input format, named exactly as in the `formats` registry in
`reshape.go`:

    <format>/input.<ext>       a small, real sample: one to five records
    <format>/golden.ndjson     the intermediate schema those records convert to

`golden_test.go` walks the registry, runs each format over its `input.*` and
compares the result byte for byte. It also checks, for every record, that
`finc.id`, `finc.source_id` and `version` are populated, and that two runs
produce identical bytes.

## Adding a format

1. Drop a real sample record or two in `<format>/input.<ext>`. The extension is
   free (the test globs `input.*`); use the one the vendor ships.
2. Generate the golden file and read it:

       go test ./internal/cmd/reshape -run TestGolden -update

3. Remove the format's line from `formatsWithoutSample` in `golden_test.go`.

Keep samples small. They exist to pin the field mapping, not to benchmark. If a
record carries anything that should not be committed, redact the values but keep
the structure — several of the samples below are already redacted upstream.

## Reviewing a change

After an intentional change to a format, regenerate and read the diff. Golden
files are one long line per record, which `git diff` renders poorly; the test's
own failure output is field level (`~ rft.jtitle: "old" -> "new"`), so running
the test before regenerating is usually the faster way to see what moved.

## Provenance

| format | source |
|---|---|
| `crossref` | first 5 lines of `fixtures/crossref.ldj` |
| `degruyter` | `fixtures/jats.xml`, a redacted De Gruyter JATS record (DOI prefix 10.14315) |
| `doaj-legacy` | first 3 lines of `fixtures/doaj.ldj` (Elasticsearch style DOAJ dump) |
| `dummy` | synthetic; `dummy` is the documented minimal example format |
| `hhbd` | first 2 records of `fixtures/hhdb.xml`, with `<Record>` lowercased to `<record>` — see note |
| `zvdd-mets` | `fixtures/sample.mets.xml`, an OAI-PMH GetRecord response |

Note on `hhbd`: `fixtures/hhdb.xml` uses capitalised `<Record>` elements, but
`hhbd.Record` is tagged `xml:"record"`, so `span-import -i hhbd` converts zero
records from that file as it stands. The sample here is the standard OAI-PMH
lowercase spelling, which is what the parser expects. Worth confirming which
spelling the live harvest actually produces.

## Formats still without a sample

Everything listed in `formatsWithoutSample` in `golden_test.go`. Those subtests
skip rather than fail, so the suite stays green while the list shrinks.

## Export side

`internal/cmd/export/testdata` holds the other end: `input.is` is the
concatenation of the golden files here, and each exporter's output is pinned
against it. Adding a sample here therefore widens the export coverage too, but
the two need to be regenerated in order:

    go test ./internal/cmd/reshape -run TestGolden -update
    go test ./internal/cmd/export  -run TestGolden -update

The second command rebuilds `input.is` from these goldens itself; `TestInputIsCurrent`
fails if it was forgotten.
