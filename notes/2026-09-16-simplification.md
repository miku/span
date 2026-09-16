# Simplification plan

Date: 2026-09-16. Goal: make span smaller, easier to follow and more robust.
It is a small tool for a handful of data jobs, and the code should look like
one.

This note is about **reduction**: what to delete, merge or flatten. For
performance, determinism and error handling, see `2026-08-19-opus-5-review.md`.
This note only brings in items from there when they also shrink the code.
Everything below was checked against the tree at `834c4b6e` (v0.2.37).

---

## 0. Where we are

| | |
|---|---|
| Go, non-test, excluding `langutil.go` | 18,866 lines |
| `langutil.go` (4 generated maps) | 15,960 lines |
| tests | 5,683 lines |
| `cmd/*/main.go` (21 binaries) | 1,983 lines |
| top-level packages | 20 (+ 22 `formats/*`, 16 `internal/cmd/*`) |
| direct dependencies | 21 |
| `docs/` / `fixtures/` in git | 48 MB / 71 MB |

`go test ./...` passes. Extracting each command into `internal/cmd/<x>` as
`Run(cfg, r, w)` is done. That was the hard part, and most of what follows is
now mechanical.

What makes span hard to follow today is not any one big file. It is **too many
small, overlapping pieces**:

* 21 binaries, some of which do the same thing;
* about 10 tiny helper packages with one or two users each;
* two or three ways to do the same basic task: parse dates, write files
  atomically, parse flags, encode JSON, make HTTP clients, dedupe crossref;
* the same boilerplate in every `main.go` (profiling, opening files,
  `log.Fatal`).

---

## 1. Delete: dead code (no risk)

Found with `deadcode ./...`. Nothing outside tests reaches any of these.

**Whole package**

* `encoding/csv`: nothing imports it.

**Functions and types**

* `filter/builder.go`: the whole programmatic builder (`NewTagger`, `NewTree`,
  `Any`, `Source`, `Collection`, `Subject`, `Pkg`, `DOI`, `ISSN`, `ISBN`,
  `Holdings`, `HoldingsWithOpts`, `Or`, `And`, `Not`). Only
  `builder_test.go` uses it. Delete both files (≈430 lines), or move the few
  cases that test the *filters* (not the builder) into `filter_test.go`.
* `xio`: `CountReader`, `AtomicDownload`, `SetFromFilename`, `UserHomeDir`,
  `SavedReaders.{Save,Remove}`.
* `container`: `NewStringSetReader`, `StringSlice` (a copy of `xflag.Array`).
* `atomic`: `Compress`, `WriteFile`.
* `strutil`: `Intersection`, `Overlap`.
* `span.GenFincID` (`common.go`). (An earlier version of this note also
  listed `span.DetectLang3`; that was wrong, `jats`, `ieee` and `zvdd` use it.)
* `licensing.Entry.containsDate`, `kbart.Holdings.Filter`,
  `tsv.NewDecoderSeparator`, `doi.NewSniffer`, `parallel.NewBytesBatch`,
  `assetutil.MustLoadStringSet`, `crossref.DefaultSnapshotOptions`,
  `folio.New`, `elsevier.Shipment.String`.
* Eleven `DefaultConfig()` functions in `internal/cmd/*` that nothing calls.
  Five were used nowhere and are deleted. The other six (`compact`,
  `comparefile`, `crossrefcmd` members/stage1, `export`, `tag`) are only
  used by tests to build configs. They stay until §3 makes the flag setup read
  its defaults from them.

**Verify:** `go build ./... && go test ./... && deadcode ./...` shows an empty
list.

---

## 2. Delete: duplicate commands

`span-index` already absorbed these. They are still built and packaged, and
siskin uses neither:

| binary | replaced by | code that goes |
|---|---|---|
| `span-report` | `span-index report` | `cmd/span-report`, `internal/cmd/report` (≈490 lines) |
| `span-compare-file` | `span-index compare` | `cmd/span-compare-file`, `internal/cmd/comparefile` (≈470 lines) |

Before deleting, compare the report names (`basic`, `json`, `fast`, `faster`
vs the `span-index report` registry). Port any report that is actually used
and drop the rest.

**Probably unused, check with operations first** (siskin does not reference
them; cron jobs or manual use might):

* `span-mail` + `mail` + `internal/cmd/mailcmd`: sending a mail is a
  `sendmail`/`curl smtp://` one-liner.
* `span-local-data`: its own doc comment says "something jq can do as well".
  siskin has one use; replace it with `jq -r`.
* `span-folio`: debugging aid; `span-freeze -folio` covers the production
  path.
* `span-crossref-snapshot` (3-stage, shells out to `sort`/`filterline`/awk,
  ≈275-line `main.go` + 932-line `crossref/snapshot.go`) vs
  `span-crossref-fast-snapshot` vs `span-compact`. These are **three ways to
  dedupe records**. siskin still uses the first two. Pick one for crossref
  (the fast one, see its 2026-03 numbers), move the siskin task over, then
  delete the other.
* Import formats that siskin never calls: `dblp`, `elsevier-tar`, `highwire`,
  `imslp`, `mediarep-dim`, `olms`, `olms-mets`, `ssoar`, `zvdd`, `zvdd-mets`,
  `hhbd`, `ceeol-marcxml`, `doaj-legacy`, `dummy`. Some siskin tasks pass
  `-i {format}`, so this list is a **candidate** list only. Each format
  dropped removes a package and often a dependency (`imslp` → `etree`,
  `ssoar` → `monday`).
* Export `formeta` + `encoding/formeta`: only used in 2016/2017 workshop
  material in siskin/docs.

Deleting source formats is the one irreversible-feeling step. It is cheap to
undo with git, but ask the people who run the index first.

---

## 3. One binary, stdlib-only dispatch

The 2026-03 plan proposed cobra. **Don't add it.** `internal/cmd/index/main.go`
already has the needed pattern in about 60 lines: a
`[]subcommand{name, short, run}` table, `help`, `version`. Promote that to
`cmd/span/main.go`:

```
span import | export | tag | redact | oa-filter | update-labels | doisniffer
span freeze
span index  query | select | compare | report | cleanup
span crossref  sync | snapshot | fastproc | members | table
span compact
```

Rules for every subcommand (write them down once in the package doc):

1. `func run(args []string, stdin io.Reader, stdout io.Writer) error`, with its
   own `flag.FlagSet`. No package-level flag vars, no `log.Fatal`, no
   `os.Exit`. This also fixes the known bug where `defer
   pprof.StopCPUProfile()` is skipped by `log.Fatal`.
2. `main` does the shared work **once**: `-cpuprofile`/`-memprofile`,
   `-v`, turning `[file...]` args into a reader (decompressing `.gz`/`.zst`
   by extension; today every siskin task writes `<(zstd -cd ...)`), and
   the exit code.
3. One flag library. Right now `flag` (20 mains) and `pflag` (`span-index`)
   are mixed. Pick stdlib `flag` and drop `pflag`; `span-index` only uses
   `StringP` for `-s/--server`, and stdlib flags accept both `-s` and `--s`.

Backwards compatibility: install `span-import` etc. as symlinks and dispatch
on `filepath.Base(os.Args[0])` (the busybox pattern). siskin keeps working
unchanged, and nfpm lists one binary plus symlinks.

What goes away: 21 `main.go` files (≈2,000 lines → ≈150), 21 Makefile
targets, 21 nfpm entries, the `update-version` sed target.

---

## 4. Fewer helper packages

Most of the small top-level packages exist for one caller. Put code next to
the code that uses it, and keep a package only when it has a real,
reusable job.

| package | users | proposal |
|---|---|---|
| `dateutil` (+ `jinzhu/now`) | crossref sync only | move into `internal/cmd/crossrefcmd`. Day/week/month intervals are ~30 lines with `time.AddDate`, so drop `jinzhu/now` |
| `xflag` | 3 mains | `Array` → inline or `flag.Func`; `Date` → `flag.Func` + `time.Parse`; `UserPassword` → use it where it is used |
| `container` | 19 importers | replace `StringSet` with `map[string]struct{}` + `slices.Sorted(maps.Keys(m))`. `MapDefault`/`MapSliceDefault` are one-line lookups. Package gone |
| `strutil` | 5 | 3 live funcs; move to where they are used (`filter`, `formats/finc`) |
| `assetutil` | 5 formats | fold into `formats/finc` (or a small `assets` loader next to `assets.go`) |
| `atomic` + `dchest/safefile` | 2 + 2 | **two** atomic-write implementations. Keep one small `CreateTemp` + `Rename` helper, drop `safefile` |
| `xio` | 7 | most of it is URL/zip readers for holdings files. Move those into `licensing/kbart` or `filter` and delete `FileReader` (marked "throw this out" since 2021) |
| `doi` | doisniffer only | move into `internal/cmd/doisniffer` |
| `encoding/tsv` | `kbart` only | move into `licensing/kbart` (or use `encoding/csv` with `Comma='\t'`, `LazyQuotes`) |
| `encoding/formeta` | formeta export | delete with the export (§2) or move under `formats/finc` |
| `mail` | mailcmd | delete (§2) |
| `folio` vs `internal/cmd/folio`, `freeze` vs `internal/cmd/freeze` | | library + command wrapper, fine, but `cmd/span-freeze/main.go` still builds its own `http.Client` and `safefile`. Move that into `freeze` |

Target top level: `formats/`, `filter/`, `licensing/`, `parallel/`,
`crossref/`, `freeze/`, `folio/`, `solrutil/`, `internal/cmd/`, `cmd/span/`.

### `langutil.go`: 15,960 → about 50 lines + data

* `ISO639NameToThreeLower` can be computed exactly from `ISO639NameToThree`
  (checked: 7,849 entries, 0 missing, 0 differences after `strings.ToLower`).
  Build it at init; that removes 7,850 lines right away.
* Next, move the tables into a `go:embed`ded TSV (the ISO 639-3 table
  it came from, as the awk comments say) and fill all maps at init. Go
  source becomes code only; the data stays diffable.
* Move it out of the root package, e.g. into `formats/finc/lang.go`.
  `span.LanguageIdentifier` has 2 callers.

### Root package `span`

After the above, `common.go` holds `AppVersion`, `Skip` and
`KeyLengthLimit`. Remove `KeyLengthLimit` (see the 08-19 note, it is
"obsolete" but still used in 6 places: replace it with an explicit truncation
where needed, or drop the truncation). Also fix the GPL header template in 10
files ("Some open source application", "along with Foobar"). The repo is
MIT per `nfpm.yaml`/`LICENSE`, so check which license applies and use one SPDX
line.

---

## 5. One way to do each thing

Having a single way per task cuts both code and surprises.

* **JSON:** `segmentio/encoding/json` in 49 files, stdlib in 4. Use stdlib
  everywhere except the hot paths (`parallel` stages: import/tag/export) and
  keep segmentio behind one small internal alias there. Or use stdlib
  everywhere and measure; Go 1.25 JSON is much faster than it was when
  segmentio was adopted.
* **Dates:** `araddon/dateparse` (3 places), `jinzhu/now`, `monday`, and
  hand-written layouts. Flags and config should take `2006-01-02` only.
  Keep fuzzy parsing only inside the formats that get messy vendor dates.
* **HTTP:** `http.Get` (tag dedup, solrutil, xio, crossref members),
  `http.DefaultClient` (freeze), `pester.New()` (sync, folio, freeze). Add
  one `newHTTPClient(timeout)` with a timeout and a simple retry loop, and
  drop `pester`. The per-record `http.Get` in `tag.droppableLabels` has no
  timeout today, so this is a robustness fix too.
* **Logging:** stdlib `log` in about 50 files, `log/slog` in 3
  (`span-import`, `span-crossref-snapshot`, `crossref/snapshot.go`), and
  `span-crossref-snapshot` imports both. Either is fine; pick one. Libraries (`formats`, `filter`, `parallel`)
  should not log. They should return errors and let the command decide.
* **Skip handling:** `err.(span.Skip)` appears in 4 places (reshape ×3,
  fastproc). Use one `errors.As` check in the shared stage helper (below).
* **Shelling out:** `sort`, `zstd`, `filterline`, awk fallback, `clam`
  templates. Keep external `sort` where it clearly wins (compact, snapshot).
  Use `klauspost/compress` in-process for compression everywhere and drop
  `clam`.

### Stage helper

`reshape` (JSON path), `tag`, `export`, `redact`, `update-labels`,
`oa-filter` and `fastproc` each repeat the same unmarshal → transform →
marshal → newline lambda around `parallel.NewProcessor`, with small
accidental differences. One generic adapter:

```go
func JSONLines[In, Out any](f func(In) (Out, error)) parallel.TransformerFunc
```

handles decode errors, `Skip` and the trailing newline in one place. Each
command then only supplies `f`. This also gives the "in/out/skipped" counters
from the 08-19 note a single hook.

---

## 6. Dependencies

With §1–§5 done, these become unnecessary:

| dep | why it goes |
|---|---|
| `google/go-cmp` | one test; `reflect.DeepEqual` or `slices.Equal` |
| `jinzhu/now` | `dateutil` inlined |
| `spf13/pflag` | stdlib `flag` |
| `dchest/safefile` | one atomic-write helper |
| `sethgrid/pester` | one HTTP client with retry |
| `miku/clam` | in-process compression, `exec.Command` |
| `adrg/xdg` | `os.UserCacheDir()` (3 users) |
| `mvdan.cc/xurls` | one use in freeze; a simple `https?://` regexp is enough for config values |
| `goodsign/monday`, `beevik/etree`, `shantanubhadoria/go-roman` | only if `ssoar`, `imslp`, `ceeol` roman numerals go / are rewritten |

This brings us from 21 direct dependencies to about 11.

---

## 7. Repository hygiene

* **`docs/` is 48 MB**, mostly 2018 `.xlsx` reports (39 MB), a 2023 crossref
  member dump (6.7 MB) and profiling PNGs. None is documentation. Move them
  out of the tree (they stay in git history). Keep `span.md`, `span.1` and
  the current docs.
* **`fixtures/` is 71 MB.** Go tests only use `fixtures/frozen.zip`. The rest
  was used to seed `internal/cmd/reshape/testdata`. Move the few still-needed
  files into `testdata/` next to their test and delete `fixtures/`.
  `hhdb.xml` alone is 45 MB, and the golden test uses 2 records from it.
* **Stale docs:** `README.md` and `docs/span.md` describe `make fast-lane` /
  `make full-lane`, but the Makefile has neither target. Either restore the
  targets or remove the section.
* **Packaging:** `nfpm.yaml` is "preferred", but `packaging/deb/.../control`,
  `packaging/rpm/span.spec`, `buildrpm.sh` and the `update-version` target
  are still there. Remove the legacy packaging.
* **Notes vs docs:** there are review/plan notes in both `notes/` and `docs/`
  (`docs/2026-02-16-cleanup-notes.md`, `docs/FastUpdate.md`,
  `docs/span-icg-issues.md`). Keep dated notes in `notes/` and user-facing
  docs in `docs/`.
* `reports/*.py`, `scripts/crossref_members_table.py`,
  `assets/finc/subjects.py`: one-off scripts. Keep only those still run,
  otherwise delete them.
* CI (`.gitlab-ci.yml`): has a `curl heise.de` proxy probe and commented-out
  cache paths. Reduce it to `go vet`, `go tool staticcheck`, `go test -race`.

---

## 8. Robustness gained along the way

The reduction work above fixes these, so no separate effort is needed:

* no `log.Fatal` in libraries (`parallel/processor.go`, `encoding/csv`) or
  commands, so deferred cleanup (profiles, temp files, atomic writes) runs;
* timeouts on all outbound HTTP (§5);
* `errors.As` for `Skip` (§5);
* `filter.Cache` package-level global (`filter/holdings.go:80`) → a field on
  `Tagger`. This is a small change once the builder code (§1) is gone.
* fewer copies of the same logic, so fixes only need to happen in one place.

---

## 9. Order

Each step can ship on its own, is checked by `go test ./...` plus the
reshape/export golden tests, and makes the next step smaller.

1. **Dead code** (§1) + `langutil` lower-map derivation. Pure deletion.
   Done 2026-09-16: −8.3k lines (mostly the lowercase language map), no
   dependency removed.
2. **Repo hygiene** (§7): move `docs` binaries and unused `fixtures` out,
   remove legacy packaging, fix stale docs. −110 MB checkout.
   Done 2026-09-16: `fixtures/` gone (`frozen.zip`, `z.zip`, `kbart.txt.gz`
   moved to `testdata/` next to their tests), `docs/` 48 MB → 1.1 MB,
   `packaging/`, `reports/`, `scripts/`, `docs/soctl.md` removed, dated notes
   moved from `docs/` to `notes/`, lane docs no longer mention make targets,
   CI runs vet + staticcheck + `go test -race`. The kbart tests had been
   skipping for years (fixture was gitignored); they now run, with expected
   counts corrected to what the tracked file contains (checked with grep).
3. **Delete duplicate commands** (§2 table): `span-report`,
   `span-compare-file`.
4. **Fold helper packages** (§4): `container`, `strutil`, `dateutil`,
   `xflag`, `doi`, `encoding/tsv`, `atomic`/`safefile`, `xio`. One package
   per commit; each is a move + import rewrite.
5. **One way per task** (§5): HTTP client, stdlib flags, JSON decision,
   `JSONLines` helper with `errors.As` Skip.
6. **Single `span` binary** (§3) with symlinks for old names. Makefile and
   nfpm get simpler. Check with siskin by running its pipeline commands
   through the symlinks.
7. **Confirm and prune** (§2, second half): unused formats, crossref snapshot
   variant, mail/local-data/folio. This needs input from operations; do it last.

Rough end state: ~11k lines of Go (from ~35k incl. `langutil`), 1 binary,
~10 top-level packages, ~10 direct dependencies, same pipeline behaviour,
checked by golden tests.

## Non-goals

* Changing the intermediate schema or the Solr output. Golden files must not
  change in steps 1–6.
* Adding new dependencies (cobra, errgroup is fine as `x/sync` is already
  indirect).
* The fast-update / articleindexd work. This plan makes it cheaper but does
  not start it.
