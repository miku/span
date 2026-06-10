# Repo review: refactoring, readability, performance notes

Date: 2026-06-10. Scope: full repo scan (~35k LOC Go excl. tests/fixtures),
plus `go vet`, `staticcheck`, `go mod tidy -diff`, and the core package test
suite (all passing). Findings ordered by severity.

## Bugs (found by reading + staticcheck)

### 1. `span-hcov` is broken: infinite recursion + dead `-l` flag

`cmd/span-hcov/main.go:109` — `normalizeSerialNumbers` ends with
`return normalizeSerialNumbers(result)` instead of `return result`,
i.e. unconditional infinite recursion (staticcheck SA5007). Every code
path through the tool calls this, so the tool crashes with a stack
overflow on any non-trivial input.

`cmd/span-hcov/main.go:72` — the ISSN list built from `-l` into `ilist`
is unconditionally overwritten by `ilist, err := indexSerialNumbers(*server)`
(staticcheck SA4006). The `-l` flag is documented as "overrides -f" but
actually does nothing. Also, when `-l` is used, `hlist` stays empty, so
`coveragePct` divides by zero (`hset.Size() == 0` → NaN).

Suggested fix: `return result`; rename/restructure so the index list and
the holdings list cannot be confused (e.g. `holdingsList` vs `indexList`),
and treat `-l` as a replacement for the holdings side as documented.

### 2. `solrutil.RandomCollection` formats the wrong variable in its error

`solrutil/index.go:332` — `fmt.Errorf("source id %s has not collections", vals)`
prints the (empty) slice instead of `sid`. Should be `sid` (and "has no
collections").

### 3. `parallel.Processor` has a data race on `wErr`

`parallel/processor.go:121` — `wErr` is written by N worker goroutines
and the writer goroutine and read by the producer without any
synchronization. The comment claims this is fine ("only one way to
toggle"), but it is a data race per the Go memory model (visible under
`-race`; `parallel` is used by nearly every cmd: span-tag, span-import,
span-export, ...). Fix is cheap: `atomic.Pointer[error]` or a
`sync.Once`-guarded setter, or restructure with `errgroup`.

Related issues in the same `Run` loop:

- A final line without a trailing newline is silently dropped:
  `ReadBytes` returns data *and* `io.EOF` together; the code `break`s on
  EOF before using `b` (processor.go:166-169).
- Workers send the transformed value to `out` even when `f` returned an
  error, so a `nil` (or partial) result is written to the output stream
  after an error occurred.
- `BytesBatch.Reset` sets `b = nil`, so the capacity from
  `NewBytesBatchCapacity(p.BatchSize)` is lost after the first batch;
  every batch reallocates. Use `bb.b = bb.b[:0]`... but note `Slice()`
  already copies, so plain reslicing is safe.

### 4. `strutil.Truncate` off-by-one and UTF-8 unsafe

`strutil/helper.go:19` — `Truncate("abc", 3)` returns `"abc..."` (longer
than the input) because the guard is `len(s) < length`, and `s[:length]`
can split a multi-byte rune. Use `if len(s) <= length { return s }` and
truncate on rune boundary (or document byte semantics).

### 5. `span-import` dispatch switch disagrees with `FormatMap`

`cmd/span-import/main.go:247-279` — the kind dispatch switch lists
`"thieme-tm"` and `"doaj-api"`, neither of which exists in `FormatMap`,
so they fail at runtime with "unknown format name" *after* being
accepted by the switch; meanwhile `"doaj-legacy"` exists in the map but
is not dispatchable at all. See refactoring #1 below for the structural
fix (single registry with a kind field).

### 6. Module hygiene: `go vet ./...` currently fails

`go.sum` is missing the entry for `github.com/google/go-cmp` (imported
by `doi/sniffer_test.go`), so `go vet ./...` and `go test ./doi/` fail
out of the box. Additionally `spf13/pflag` is a direct dependency
(cmd/span-index) but is marked `// indirect`. One `go mod tidy` fixes
both (verified with `go mod tidy -diff`).

### 7. Possible aliasing bug in `HoldingsFilter.Apply`

`filter/holdings.go:183` — `append(is.ISSN, is.EISSN...)` may write into
the backing array of `is.ISSN` if it has spare capacity. Here it's
read-only afterwards, so it's latent rather than live, but the idiom is
a trap; `slices.Concat(is.ISSN, is.EISSN)` is explicit and safe.

## Performance

### 1. KBART date-parse cache is thrown away in the span-tag hot path

`licensing/entry.go` caches parsed `FirstIssueDate`/`LastIssueDate` in
`entry.parsed`, populated lazily by `begin()`/`end()`. But
`HoldingsFilter.Apply` / `covers` (filter/holdings.go:163,186) iterate
`SerialNumberMap[issn]` which yields `licensing.Entry` *by value* — the
cache is filled in a copy and discarded after each record. With millions
of records × up to 25 `time.Parse` layout attempts per boundary per
match, this is significant repeated work in span-tag's hottest loop.

Options: store `[]*licensing.Entry` in the cache maps (entries become
shared, cache fills once — also removes any future race concern since
population then happens once at load time, ideally eagerly right after
`ReadFrom`), or precompute `parsed` for all entries when the holdings
file is registered.

Note while touching this: the lazy cache mutation via pointer receivers
is also not goroutine-safe; if entries become shared pointers, populate
the cache eagerly at load time rather than lazily under concurrency.

### 2. `parseWithGranularity` does a redundant O(n) lookup

`licensing/entry.go:329-342` — inside the loop over `datePatterns` it
calls `getGranularity(dfmt.layout)`, which linearly re-scans
`datePatterns` to find the granularity that is already in hand as
`dfmt.granularity`. Use `g = dfmt.granularity` and delete
`getGranularity` entirely.

### 3. `span-tag` re-tokenizes the prefs string per comparison

`cmd/span-tag/main.go:80-89` — `preferencePosition` calls
`strings.Fields(*prefs)` on every invocation, and it is invoked twice
per (label × doc) pair in `DroppableLabels`. Parse `*prefs` once at
startup into a `map[string]int`.

### 4. `crossref/snapshot.go` minor wins

- `processFile` (snapshot.go:417-421): `scanner.Text()` allocates a
  string per line, then `[]byte(line)` copies it back for
  `json.Unmarshal`. Unmarshal from `scanner.Bytes()` and only convert to
  string for the callback (or pass bytes through).
- `processFile` allocates a 100MB scanner buffer per file
  (`MaxScanTokenSize`), i.e. workers × 100MB up front. Start small and
  let `Buffer(buf, max)` grow: `scanner.Buffer(make([]byte, 64*1024), MaxScanTokenSize)`.
- `isZstdCompressed` opens the file and decompresses 64 bytes; checking
  the 4-byte zstd magic (`28 B5 2F FD`) is enough. In fact the call site
  is decidable statically: the index file is *always* written through
  `zstd.NewWriter` in stage 1, so the non-zstd branch of
  `identifyLatestVersions` is dead code.

### 5. `xio` readers buffer entire payloads in memory

`LinkReader`, `ZipContentReader`, `ZipOrPlainLinkReader` slurp the whole
remote file/archive into a `bytes.Buffer` before the first `Read`. Fine
for small holdings files; worth a doc comment at least (only
`ZipContentReader` has one). If any large inputs flow through these,
stream instead.

## Streamlining / refactoring

### 1. span-import: single source of truth for formats

`FormatMap` (names → factory) and the giant switch in `main` (names →
xml/json/text kind) must be kept in sync by hand and have already
drifted (bug #5). Replace with one registry:

```go
type format struct {
    kind string        // "xml", "json", "text", "special"
    new  func() any
}
var formats = map[string]format{ ... }
```

`-list`, dispatch, and the existence check then all derive from one map,
and adding a format is a one-line change.

### 2. Consolidate on one `parallel` package

`doi/sniffer.go` imports the external `github.com/miku/parallel` while
the other 10+ commands use the in-repo `span/parallel` fork. Switch
sniffer.go to the local package and drop the external dependency (the
local fork is also where the race fix from bug #3 will land).

### 3. Deduplicate the batched-write idiom in snapshot.go

`readCacheToIndex` (snapshot.go:204-249) and the closure in
`extractMinimalInfo` (snapshot.go:537-596) both implement "accumulate
lines in a bytes.Buffer, flush under mutex every BatchSize entries,
final flush". Extract a small `batchedWriter` type
(`Write(line)`, `Flush()`, holds buf/mutex/zw/batchSize) and both call
sites shrink considerably.

### 4. `groupLineNumbersByFile` cleanup (snapshot.go:733-818)

- Temp files are closed twice: once via the per-file `defer` registered
  inside the loop (line 774) and again in the explicit loop at line
  798-802 — the second close returns an error on already-closed files on
  some platforms; currently it works by accident because `*os.File.Close`
  on a closed file returns `ErrClosed` which would abort the function.
  Drop the in-loop `defer` and keep the explicit close.
- `linesRead` is incremented but never used; delete.
- The guard `if len(parts) < 2 || len(line) < 2 || ...` followed by
  `if len(parts) != 2` is confusing; one check (`len(parts) != 2 ||
  strings.HasPrefix(line, "#")`) with a `continue`-vs-error decision
  expressed once would read better.
- The per-input-file `sort -n` loop at the end runs sequentially with
  `--parallel NumCPU` each; many small files would benefit more from
  sorting K files concurrently with `--parallel 1`.

### 5. `processFilesParallel` keeps working after the first error

`crossref/snapshot.go:442-470` — when a worker hits an error it exits,
but the other workers keep draining the (fully buffered) file channel,
so a failing 500-file run still processes ~all files before reporting.
`golang.org/x/sync/errgroup` with a shared `context` cancel gives
fail-fast semantics and deletes ~25 lines of channel plumbing
(`errgroup` is already an indirect dependency).

Also: `SortFilesBySize` (snapshot.go:903) never returns a non-nil error —
drop the error from its signature and the caller's dead error branch.

### 6. `solrutil.Index.Select` duplicates link construction

`index.go:121-128` builds `Server + "/select?" + vs.Encode()` inline
while `selectLink`/`FacetLink` exist for the same purpose. Add a
`link(vs url.Values)` helper used by all three. Similarly,
`FacetKeysFunc` and `FacetKeys` are near-identical — implement
`FacetKeys` as `FacetKeysFunc(q, f, func(string,int) bool { return true })`.

`PrependHTTP` (index.go:338) matches any prefix "http", including
"httpfoo.example.com"; check `^https?://` instead.

### 7. `xio.FileReader` is marked for deletion — delete it

`xio/io.go:141` carries `TODO(miku): Throw this out` since long ago.
Its only in-repo use is inside `ZipOrPlainLinkReader.fill`, where a
plain `os.Open`+`io.Copy` suffices. Same file: `UserHomeDir` duplicates
`os.UserHomeDir` from the standard library; `ReadLines`,
`SetFromFilename`/`LoadSet`, and `container.NewStringSetReader` are
three implementations of "read lines into a collection" — one
`iter.Seq[string]` line iterator could back all of them. Both also share
the same subtle bug as parallel.Run: a last line without `\n` is dropped
(`ReadString` returns data with `io.EOF`).

### 8. `container.StringSet` modernization

Go 1.23+ makes most of `container/string.go` one-liners over
`maps`/`slices`; the type is still fine as sugar, but:

- `NewStringSet` has the XXX "make the zero value usable" — doing that
  (lazy init in `Add`) removes a class of nil-map panics.
- `Add`'s doc comment ("returns true if added") describes a return value
  that doesn't exist.
- `StringSlice` duplicates `xflag.Array` (xflag package exists for
  exactly this); keep one.

### 9. Repeated per-cmd boilerplate

Every command repeats the same ~30 lines: version flag handling,
cpu/mem profile setup, multi-file `io.MultiReader` stdin fallback,
buffered stdout writer (see span-tag/main.go:148-224 and
span-import/main.go:201-246, span-export, span-oa-filter, ...). A tiny
internal `cli` helper package (e.g. `RunWithProfiles(fn)`,
`InputReader(flag.Args())`) would cut a few hundred lines across 24
cmds and make new tools cheaper to add. The new span-index subcommand
layout (cmd/span-index/main.go) is a good template for multi-command
binaries; the 2026-03 single-binary plan note points the same way.

### 10. Filter package niceties

- `firstKey` (filter/filter.go:129) unmarshals the whole JSON fragment
  into `map[string]any` just to read the single top-level key — for
  holdings filters this re-parses a potentially large embedded config.
  A `json.Decoder` token peek does it without materializing the value.
- `Tagger.Tag` iterates `FilterMap` (a map) so label order is random;
  output `Labels` order differs run-to-run, which makes diffs of tagged
  output noisy. Sorting keys once and iterating a slice gives
  deterministic output for free.
- Package-level mutable singleton `filter.Cache` (holdings.go:79) makes
  tests and concurrent loads fragile; consider hanging the cache off the
  Tagger.

### 11. Licensing readability

`licensing/entry.go` `Covers`/`CoversDate`/`containsDate` overlap;
`CoversDate` = `Covers` minus volume/issue. Could collapse into one
method with optional args or document the intended entry points.
`findInt` converting via `ParseInt(..., 32)` then `int(...)` is fine but
`Atoi` alone with a bounds note would read simpler.

## Smaller readability items

- staticcheck S1025 at cmd/span-compare/main.go:272 (`fmt.Sprintf` on a
  string) and S1038 at formats/elsevier/dataset.go:464,474
  (`log.Println(fmt.Sprintf(...))` → `log.Printf`).
- `crossref/snapshot.go:252` typos in doc comment: "Tihs", "continously";
  snapshot.go:41 "numebers"; entry.go:328 "recorgnized";
  zvdd `DublicCoreRecord` (typo'd type name, exported).
- License headers still read "This file is part of some open source
  application" / "along with Foobar" — fill in the project name once,
  repo-wide (the GPL header template was never instantiated).
- `span.KeyLengthLimit` (common.go:38) is self-described as obsolete
  since 2017 — check usages and delete.
- `GenFincID` (common.go:54) takes a `sid` parameter but hardcodes the
  `"ai"` prefix despite the doc claiming an "arbitrary prefix".
- `xio.CountReader`/`WriteCounter` use `sync/atomic` int ops; fine, but
  `atomic.Int64` types are the current idiom and prevent accidental
  unsynchronized access.
- `langutil.go` is 16k lines of generated tables in the root package;
  moving it to e.g. `langutil/` (own package, with the `awk` generation
  command kept as `//go:generate`) would make the root package readable
  and shrink compile units for consumers that don't need it.

## Suggested order of attack

1. `go mod tidy` (unbreaks `go vet`/`go test ./doi`), fix span-hcov,
   `RandomCollection` error, staticcheck nits — small, zero-risk PRs.
2. `parallel`: fix the `wErr` race + trailing-line drop (add a `-race`
   test), then migrate `doi/sniffer.go` off the external fork.
3. span-import format registry unification (fixes the thieme-tm/doaj-api
   drift).
4. Holdings entry pointer/eager-parse change — biggest measurable
   speedup for span-tag; benchmark before/after with a fixture.
5. Opportunistic: snapshot.go batched-writer extraction and errgroup
   migration next time that file is touched.
