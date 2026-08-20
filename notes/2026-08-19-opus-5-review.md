# Repo review: design, streamlining, performance

Date: 2026-08-19. Scope: full repo (~40k LOC Go), read against the two earlier
notes (`2026-06-10-fable-5-review.md`, `2026-08-02-span-fast-update.md`).
All claims below were checked against the current tree; measurements are from
this machine (Apple A18 Pro, 6 physical cores, darwin/arm64).

Baseline health, verified: `go vet ./...`, `staticcheck ./...` and
`go mod tidy -diff` are all clean, and `go test ./...` passes. That is a
better starting point than the June note describes — most of it has been acted
on (see §1).

---

## 0. What span is, structurally

Worth stating plainly, because the rest of the note leans on it.

span is a **hub-and-spoke converter** wrapped in a **Unix pipeline**.

* The hub is `finc.IntermediateSchema` (IS). 22 source formats convert *into*
  it; 2 export formats convert *out of* it. This is the same shape as pandoc's
  AST or a compiler IR, and it is the single best decision in the codebase:
  adding a source is one package plus one line in a registry, and it composes
  with every downstream stage for free.
* The composition unit is a **process**: ndjson on stdin, ndjson on stdout.
  `span-import -i crossref | span-tag -c cfg | span-export -o solr5vu3`.
* Every stage has the same internal shape: `parallel.NewProcessor(r, w, f)`
  where `f` is `func(lineno int64, b []byte) ([]byte, error)`.

Compared with how data-processing libraries are usually built (Beam/Spark/Arrow
/dbt, or in Go: `errgroup` + typed channels + `iter.Seq`), span differs in four
ways that are worth examining rather than defending:

| | span today | common practice |
|---|---|---|
| stage interface | `[]byte -> []byte` | typed records, `iter.Seq2[T, error]` |
| adjacent stages | separate processes, full JSON round-trip each | fused; parse once |
| output order | not preserved, not deterministic | deterministic, or explicitly opted out |
| bad records | abort mid-stream, partial output | dead-letter + threshold |

Three of those four (fusion, determinism, bad records) are also **prerequisites
for the `span-fast-update` plan** in the 2026-08 note. That plan wants in-process
composition, watermarks, and idempotent re-runs; the current stage design fights
all three. So the refactors below are not cosmetic — they are the groundwork for
the tool that note describes.

The internal restructuring already done (every command's logic extracted to
`internal/cmd/<x>` as `Run(cfg, r, w) error`, with the thin `cmd/*/main.go`
kept for the binary) was exactly the right move and makes everything below
cheaper. It is what makes `span-fast-update` a couple of function calls rather
than a rewrite.

---

## 1. Status of the June 2026 review

Most of it landed. Recording this so the list is not re-derived a third time.

**Done:** format registry unified (`internal/cmd/reshape`, single `formats` map
driving `-list`, dispatch and existence check — the `thieme-tm`/`doaj-api` drift
is gone); `parallel` `wErr` race fixed (mutex-guarded `setErr`/`getErr`);
trailing-line-without-newline fixed; `BytesBatch.Reset` keeps capacity;
`licensing.Entry.ParseDates` added and called eagerly from
`kbart/holdings.go:46`, so the date cache survives the value copies into the
lookup maps; `span-hcov` removed; `RandomCollection` error message fixed;
`slices.Concat` in `HoldingsFilter.Apply`; `go.sum`/`pflag` hygiene fixed;
staticcheck nits gone.

**Still open, still worth doing:** `xio.FileReader` still carries
`TODO(miku): Throw this out` (`xio/io.go:141`); `container.StringSlice` still
duplicates `xflag.Array`; `langutil.go` is still 15,960 lines in the root
package; the GPL header template is still uninstantiated ("some open source
application", "along with Foobar") in 10 files; `span.KeyLengthLimit`
(self-described as obsolete since 2017) is still load-bearing in six format ID
paths — either delete it or drop the "obsolete" comment, but it should not be
both; `solrutil.PrependHTTP` still matches bare `http` prefixes;
`solrutil.Index.Select` still builds its link inline instead of via
`selectLink`; `Tagger.Tag` still iterates a map (see §3);
`tag.preferencePosition` still calls `strings.Fields(c.Prefs)` on every
invocation.

---

## 2. The main performance finding: stage framing, not stage work

**`parallel.Processor` batches on the way in but not on the way out.** Workers
send *one result per record* over an unbuffered `out` channel to a single
writer goroutine (`parallel/processor.go:149`, `167-168`). Input batching of
10–20k is therefore undone immediately: a 100M-record run performs 100M
unbuffered channel handoffs, all serialised through one goroutine.

Measured, `span-export -o solr5vu3` over 300k IS records (270 MB), 3 runs each:

```
w=1     2.71  2.86  2.69     (~2.75s)
w=6     1.54  1.55  1.58     (~1.56s)   1.8x on 6 cores  (~30% efficiency)
w=8     1.39  1.64  1.44     (~1.49s)
w=16    1.32  1.18  1.40     (~1.30s)   2.1x
```

`span-import -i crossref` behaves the same way (1.16s → 0.62s, 1.9x).

Two things stand out. Six workers on six cores buys 1.8×, and **oversubscribing
to 16 workers still helps** — a workload that were genuinely CPU-bound would
flatten or regress there. The CPU profile says the same thing outright:

```
29.32%  syscall.rawsyscalln
25.30%  runtime.usleep            <- scheduler spin
 9.24%  runtime.pthread_cond_wait
 ...
12.65%  finc.(*Solr5Vufind3).convert   (cum)  <- the actual work
```

~64% of samples are in syscall, spin and wait. The transform is a footnote.

To size the ceiling, I prototyped a variant where workers accumulate a batch's
results into one `bytes.Buffer` and send **one `[]byte` per batch**, with small
buffered channels. On a cheap (memcpy) transform, where framing is all there is:

```
current      256 MB/s
batched-out 1352 MB/s     5.3x
```

That 5.3× is the framing headroom, not the expected end-to-end gain — real
stages do real work. But given that the profile shows framing dominating even
the JSON-heavy export stage, a meaningful fraction of it is recoverable.

**Recommendation.** Rework `parallel.Processor` to batch the output side:
workers write into a per-batch buffer and send once per batch; give `queue` and
`out` a small buffer (`NumWorkers`). This is contained — the exported API
(`NewProcessor`, `BatchSize`, `NumWorkers`, `Run`, `RunWorkers`) need not
change, so all call sites benefit without edits. Add a `-race` test and a
benchmark in the same change.

Do this one first. It is the highest ratio of measured win to blast radius in
the repo, and it makes §3 nearly free.

---

## 3. Output is not deterministic — verified

Two independent sources of run-to-run variation:

1. `parallel.Processor` does not preserve input order (documented, deliberate).
2. `filter.Tagger.Tag` (`filter/filter.go:60`) iterates `FilterMap`, a Go map,
   so the order of `is.Labels` is randomised per record per run.

Verified for (1) — same binary, same input, twice:

```
run1 (w=8)  37e627609b01472d18d0745540e26652
run2 (w=8)  ad7532662a226600e6c3da7ea02dfa76
run3 (w=1)  bf2b9ca83e5a7d89846b8c0d480c791a
```

Today this mostly costs diff noise (and makes `span-compare-file` harder to
reason about). It becomes a real constraint under the fast-update plan: "has
this record changed since yesterday?" is the cheapest possible incremental
filter, and it requires byte-stable output. Content-addressed caching, golden
tests over fixtures, and reproducible builds of the index all want the same
property.

Both fixes are cheap:

* Sort `FilterMap` keys once at load and iterate a slice. Deterministic
  `Labels`, no measurable cost, done.
* Order-preserving output falls out of §2 almost for free: batches are already
  numbered, so tag each with a sequence number and have the writer hold
  out-of-order batches in a small map until their turn. The buffering is bounded
  by `NumWorkers` batches. Worth making it the default and keeping an opt-out
  (`Processor.Unordered = true`) rather than the other way round.

---

## 4. Error semantics: partial output on a single bad record

Current behaviour in `parallel.Run`: a transform error is recorded, that record
is dropped, the producer stops enqueuing **at the next batch boundary**, and
`Run` returns the error. Records already handed off still get written.

Measured — 200,001 IS records with exactly one malformed line inserted at
position 100,001:

```
exit code 1
stdout: 119,999 records   (of 200,000 good ones)
```

So the tool does fail loudly, which is right. But it leaves **119,999 records of
plausible-looking ndjson on stdout**, and how far past the error it gets is
arbitrary (batch-boundary dependent). In the shell-pipeline composition style
this repo is built around, `a | b | c > out` reports only `c`'s status unless
`set -o pipefail` is set, so a truncated intermediate is easy to miss — and the
truncation is silent, because a short file looks exactly like a small one.

For vendor metadata, one malformed record in a multi-million-record dump is the
normal case, not an exceptional one. Aborting the batch is the wrong default.
Standard practice is a **dead-letter channel plus a tolerance threshold**:

```go
type Processor struct {
    // ...
    MaxErrors  int       // 0 = fail on first (current behaviour)
    ErrWriter  io.Writer // rejected records go here, with lineno + reason
}
```

Then `span-import -max-errors 100 -rejects bad.ndjson` becomes the normal
invocation, the run completes, and the rejects are inspectable. Records dropped
by design (`span.Skip`, `-drop-dangling`) should flow through the same
accounting.

Related: **nothing counts records.** No stage reports how many came in, went
out, were skipped or were rejected. Every stage can silently drop records
(`span.Skip` in ~20 formats, `DropDangling` in tag, error-drop in `parallel`),
and today the only way to notice is to `wc -l` both sides. For a pipeline about
to move to unattended daily incremental runs, a one-line summary per stage to
stderr — `in=N out=M skipped=S rejected=R elapsed=T` — is the single cheapest
operability improvement available, and it is what makes a watermark advance
trustworthy.

---

## 5. Streamlining: a typed stage, and stage fusion

### 5a. The three stages are the same twenty lines

`reshape.processJSON`, `tag.Run` and `export.Run` each write:

```go
parallel.NewProcessor(r, w, func(_ int64, b []byte) ([]byte, error) {
    var v T
    if err := json.Unmarshal(b, &v); err != nil { return b, err }
    out, err := transform(v)          // the only line that differs
    if err != nil { ... }
    bb, err := json.Marshal(out)
    bb = append(bb, '\n')
    return bb, nil
})
```

The differences between the three copies are accidental, not intentional:
`export` returns `b` on unmarshal error while `reshape` returns `nil`; only
`reshape` handles `span.Skip`; only `export` logs the offending record. Any
change to skip/error/logging policy has to be made three times and currently
has not been.

With generics this collapses to one helper:

```go
// parallel.JSONLines adapts a typed record transform into a TransformerFunc,
// centralising unmarshal, skip handling, marshal and the trailing newline.
func JSONLines[In, Out any](f func(In) (Out, error)) TransformerFunc
```

Each stage then supplies only its transform. Uniform error handling, uniform
skip handling, uniform accounting hook for §4, ~60 lines deleted. `span.Skip`
handling lives in exactly one place, which matters for §6.

### 5b. Fusion: the pipeline parses each record three times

`import | tag | export` currently performs, per record: parse source JSON →
marshal IS → **parse IS → marshal IS** → **parse IS** → marshal Solr. Two full
JSON round-trips exist purely as the inter-process wire format. Given that
`Solr5Vufind3.convert` is 12.65% of the export stage's own profile, the
serialisation around it is not a rounding error.

Every mature data-processing framework fuses adjacent narrow transformations for
exactly this reason. Once §5a exists, fusion is nearly free: a fused stage is
function composition over `finc.IntermediateSchema`, and

```go
pipeline := parallel.JSONLines(compose(importFn, tagFn, exportFn))
```

parses once and serialises once. **This is precisely what milestone 3 of the
fast-update plan asks for** ("pull import/tag/export in-process; stream between
stages; drop temp files"), so it is worth building the typed helper with that
milestone in mind rather than as a separate cleanup. The separate binaries keep
working — fusion is an additional composition, not a replacement.

### 5c. One binary

The `internal/cmd` extraction already did the hard part: every command is a
library with a `Run`. `internal/cmd/index/main.go` is a good template — a
`subcommand{name, short, run}` table and a 22-line `cmd/span-index/main.go`.
Applying that shape to a single `span` binary is now mostly mechanical and would
retire 20 `main.go` files, 20 Makefile targets and 20 nfpm entries. The
2026-03 plan describes this; it is much closer to reach than when it was
written.

While doing that, fix the boilerplate bug it would eliminate: several mains do
`defer pprof.StopCPUProfile()` and then `log.Fatal(err)` (e.g.
`cmd/span-export/main.go:56,79`). `log.Fatal` calls `os.Exit`, deferred
functions do not run, so **the CPU profile is truncated on exactly the failing
runs you wanted to profile**, and `-memprofile` never writes at all. The fix is
the standard `func run() error` + `os.Exit(1)` in `main` shape, applied once in
the shared entry point instead of 20 times.

---

## 6. `span.Skip` is matched by type assertion, not `errors.As`

All three call sites use `if _, ok := err.(span.Skip); ok`
(`reshape.go:169,196,239`). The codebase wraps errors with `%w` in 62 places.
The day someone writes `fmt.Errorf("record %s: %w", id, span.Skip{...})` — a
completely natural thing to do — that record stops being skipped and instead
aborts the run (§4: with partial output). Nothing currently triggers it; it is a
tripwire, not a live bug.

Use `errors.As`, and give `Skip` a pointer-free `As`-compatible shape. Best done
inside the §5a helper so it exists once.

---

## 7. No `context.Context` anywhere

Verified: zero occurrences in the repo. Consequences:

* **No timeouts on outbound HTTP.** `tag.droppableLabels` calls bare
  `http.Get` (`internal/cmd/tag/tag.go:192`) — `http.DefaultClient`, no
  timeout. On the `-server` dedup path this is one synchronous request *per
  record*, and a hung SOLR hangs the run indefinitely. Same in
  `solrutil.decodeLink:165` and `xio/io.go:54`.
* **No cancellation.** Ctrl-C during a multi-hour run cannot unwind cleanly or
  advance a watermark.
* The fast-update plan's own interface sketch is
  `Delta(ctx context.Context, ...)`, so this has to be introduced regardless.

Also on that hot path, and cheap to fix while there: `DefaultTransport` allows
only 2 idle connections per host, so N workers hammering one SOLR mostly tear
down and re-establish connections. A package-level `*http.Client` with a
timeout and `MaxIdleConnsPerHost = NumWorkers` is a few lines. DOIs repeat
across sources, so a small LRU in front of `droppableLabels` would cut the
request count further. And `preferencePosition` re-tokenises `c.Prefs` with
`strings.Fields` on every call, twice per (label × doc) — parse it once into a
`map[string]int` at config load (still open from June).

---

## 8. The biggest maintainability gap: `formats/` is untested

**20 of 22 format packages have no test file at all.** Only `crossref` and
`ceeol` have any. `formats/finc` — the intermediate schema itself and the
`Solr5Vufind3` exporter, i.e. the definition of the hub and the shape of every
document in the production index — has none.

This is inverted relative to risk. The format packages are the highest-churn,
most-adversarial code in the repo (vendor XML with bad entities, mixed
encodings, missing dates, `XXX: Is it?` guesses at genre mapping — there are ten
such markers in `formats/`). They are also the easiest to test, because the
whole contract is a pure function: bytes in, `IntermediateSchema` out.

The infrastructure already exists — `fixtures/` has real samples and
`schema/fixtures/{0.9,1.0}` holds schema versions. What is missing is one
table-driven golden test:

```
formats/testdata/<name>/input.<ext>
formats/testdata/<name>/golden.ndjson
```

with a single test that walks the `formats` registry, runs each format over its
input, and compares against golden (`-update` flag to regenerate). One test
function covers all 22 formats, adding a format means adding two files, and
schema changes surface as a reviewable golden diff instead of a surprise in the
index. Seed it from real records already in `fixtures/`.

This is more valuable than any performance work here. It is also what makes the
performance work above safe to do: §2, §3 and §5 all change how records move
through the system, and right now there is very little that would notice if a
field quietly stopped being populated.

---

## 9. Build and tooling bugs

Both verified by running them.

**Incremental builds do nothing.** The static pattern rule

```make
$(TARGETS): %: $(wildcard cmd/%/*.go)
```

expands `$(wildcard ...)` at parse time with a literal `%`, which matches
nothing, so every target has **no prerequisites**. Once a binary exists, make
considers it current forever:

```
$ touch internal/cmd/tag/tag.go && make span-tag
make: `span-tag' is up to date.
```

Editing source and re-running `make` silently ships the old binary. This is a
daily-friction bug and the kind that costs an afternoon when it bites. It also
misses `internal/cmd/**` entirely, which is now where the logic lives — so a
correct prerequisite list is not just `cmd/%/*.go`. Simplest robust fix is to
let `go build` do the staleness analysis (declare the targets `.PHONY`), since
Go's own build cache already makes no-op rebuilds fast.

**`make lint` invokes a binary that does not exist:** `golanglint-ci run ./...`
— transposed, should be `golangci-lint`. Given `go.mod` already has a `tool`
block with `staticcheck` and `goimports`, the tidiest fix is to point `lint` at
`go tool staticcheck ./...` plus `go vet`, which are known to pass today, and
drop the dependency on an uninstalled binary.

**Minor `go.mod` items:** both `mvdan/xurls` and `mvdan.cc/xurls` are present
(the former indirect — one is the old import path); `golang.org/x/sync` is
already available for `errgroup` if §2/§4 want it. Also, JSON is split between
`segmentio/encoding/json` (49 files) and stdlib `encoding/json` (4 files:
`freeze/unfreeze.go`, `internal/cmd/crossrefcmd/sync.go`,
`internal/cmd/folio/folio.go`). The split looks incidental rather than chosen;
worth either unifying or leaving a one-line comment where stdlib is deliberate,
since the two differ subtly on `html` escaping and number handling.

---

## 10. Smaller items

* `reshape.processXML:177` constructs `json.NewEncoder(w)` inside the per-record
  loop. Hoist it out — one allocation per record on the XML path for no reason.
* `parallel.Processor.BatchMemoryLimit` defaults to 32 GB and is checked *after*
  the record is appended; the log message ("trim batch to ...") describes
  trimming that does not happen. Either implement the limit or drop it.
* `filter.Cache` (`filter/holdings.go:80`) is a package-level mutable singleton.
  It makes tests order-dependent and concurrent config loads unsafe. Hanging it
  off the `Tagger` is a contained change and removes the `HoldingsFilter.Names`
  / `CachedValues` double bookkeeping (the struct currently carries both a key
  list *and* a resolved map, and `Apply` uses the global rather than either).
* `oafilter.go:34` — `XXX: This can take up significant memory (e.g. 40% of
  16G)`. If that is a known operational cliff it deserves either a fix or a
  documented bound, not an XXX.
* `internal/cmd/index` is the good template. Worth pointing new subcommands at
  it explicitly in `docs/` so the pattern spreads rather than being rediscovered.

---

## 11. Suggested order

Sequenced so each step makes the next one safer or cheaper.

1. **Golden tests for `formats/`** (§8). Not because it is the most exciting,
   but because everything below moves records around and there is currently
   little that would catch a regression. One test harness, seeded from existing
   fixtures.
2. **Makefile fixes** (§9). Ten minutes, and until it is done every "I tested
   that" below is suspect.
3. **`parallel.Processor`: batch the output side** (§2). Measured win, exported
   API unchanged, all stages benefit. Land with a `-race` test and a benchmark.
4. **Determinism** (§3): sorted `FilterMap` iteration, plus order-preserving
   output, which is nearly free once (3) is done. Verify with the md5 check
   above.
5. **Error tolerance + counters** (§4). `MaxErrors`/`ErrWriter`, and the
   one-line per-stage summary. This is what makes unattended daily runs
   trustworthy.
6. **Typed stage helper** (§5a) and `errors.As` for `Skip` (§6). Deletes the
   triplicated boilerplate and gives (5) a single place to hook accounting.
7. **`context.Context` + a real HTTP client** (§7). Required by the fast-update
   plan anyway; do it as that work starts rather than speculatively.
8. **Fusion + single binary** (§5b, §5c) — as `span-fast-update` is built, since
   that tool is the first consumer of both.

Steps 1–5 are independently valuable and none of them require agreeing on the
articleindexd/fast-update direction. Steps 6–8 are where this review and the
2026-08 plan converge: the typed stage, cancellation and fusion are the same
work, whether it is framed as cleanup or as building the new tool.
