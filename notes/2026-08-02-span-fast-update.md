# span-fast-update

The span-fast-update tool will be a single, unifying, low maintenance tool to
update the article index of the finc project (at leipzig university library).

The article index is driven by a couple of projects:

* this span toolkit
* @../siskin/ for orchestration, especially @../siskin/siskin/workflows/ai.py

Plus various command line tools, e.g.

* @../solrbulk/...
* @../metha/...
* @../solrdump/...

We want to consolidate the update (e.g. per day) in form of a (mostly)
self-contained span-fast-update tool. This serves as a daily update mechanism,
that will only insert (or upsert) records in the index.

For a fresh update, we will do a full update, but we do not need to do a full
update every day (wears the system and the nvmes).

Create a plan first. Ok to shellout to tools, but also nice to reuse Go code, if possible.


## Plan: span-fast-update

### 1. Goals

* One tool, `span-fast-update`, that performs a **daily incremental update** of
  the article index (SOLR `biblio`), inserting/upserting only records that
  changed since the last run.
* **Low maintenance**: a single Go binary with a small config file; no luigi
  DAG, no dozens of intermediate targets to reason about.
* **Mostly self-contained**: prefer reusing span's own Go packages in-process;
  shell out only where a mature external tool already does the job (harvesting,
  bulk indexing).
* **Idempotent & resumable**: safe to re-run; a crashed run leaves the index
  consistent and the next run picks up from the last high-water mark.
* Does *not* replace the full rebuild — the full pipeline stays for cold starts
  and periodic reconciliation. Fast update is the day-to-day path so we stop
  re-processing terabytes (and wearing the NVMes) every day.

### 2. Where we are today (recap)

The current daily/periodic update is a luigi pipeline in
`../siskin/siskin/workflows/ai.py` + `../siskin/siskin/sources/*.py`. Per
source the shape is:

```
harvest (metha/solrdump/API) → span-import -i <fmt>   # -> intermediate schema (IS), zstd
                             → span-tag -unfreeze cfg  # attach ISILs (licensing), dedup
                             → span-export -o solr5vu3 # -> SOLR docs
```

Sources are concatenated (`AIExport`) and loaded with **solrbulk**. Crossref —
the dominant daily mover — already has a delta fetcher:

```
span-crossref-sync -p zstd -P feed-<n>- -i d -s <begin> -e <date> -c <dir>
```

Licensing config comes from FOLIO via `span-freeze` (a frozen zip consumed by
`span-tag -unfreeze`). SOLR upserts by document `id`, so re-indexing a changed
record simply overwrites it — this is the property the whole "fast update" idea
rests on.

### 3. Design principles

1. **Delta in, upsert out.** Only fetch records changed since the last run;
   let SOLR's overwrite-by-id semantics do the "update". No purge, no full
   reindex.
2. **Stream, don't stage.** Where practical, pipe `import → tag → export →
   index` in-process over an `io.Reader`/channel rather than writing large
   zstd intermediates to disk. Falls back to temp files only when a stage
   genuinely needs a second pass.
3. **Reuse span internals.** `span-import`, `span-tag`, `span-export` are thin
   `cmd/` wrappers over packages (`formats`, `filter`/`tag`, `schema`/export).
   Call those packages directly so the fast-update binary is one process.
4. **Shell out only at the edges.** Harvesting (`metha`) and bulk load
   (`solrbulk`) are mature and out-of-scope to reimplement; invoke them (or
   import solrbulk as a library — it is Go).
5. **Explicit state.** A tiny on-disk state file records, per source, the last
   successfully-indexed high-water mark (date/token) + a run log.

### 4. Architecture

```
                 ┌──────────────── span-fast-update (one binary) ───────────────┐
                 │                                                               │
  state.json ──► │  for each enabled source in config:                          │
                 │    1. fetch delta   (crossref-sync / metha, since watermark)  │
                 │    2. import        (formats pkg  → intermediate schema)      │
                 │    3. tag/license   (tag pkg, -unfreeze frozen FOLIO cfg,     │
                 │                      optional -server dedup vs live index)    │
                 │    4. export        (export pkg   → solr5vu3 docs)            │
                 │    5. index         (solrbulk lib → POST /update, commit)     │
                 │    6. advance watermark, append run log                       │
                 └───────────────────────────────────────────────────────────────┘
                                              │
                                          SOLR biblio
```

Each source is a small struct implementing an interface, e.g.:

```go
type Source interface {
    Name() string
    // Fetch delta since the watermark; returns a reader of raw records
    // plus the new watermark to commit on success.
    Delta(ctx context.Context, since Watermark) (io.ReadCloser, Watermark, error)
    ImportFormat() string // e.g. "crossref", for the formats package
}
```

The core loop (fetch → import → tag → export → index → commit-watermark) is
source-agnostic; adding a source is writing one `Delta`.

### 5. Incremental correctness

* **Overwrite by id.** Exported SOLR docs carry a stable `id`. Re-indexing a
  changed record overwrites the previous version — no duplicates from the
  update itself.
* **Cross-source dedup (DOI collisions).** The full pipeline dedups across
  sources (`groupcover` / `span-update-labels`, and `span-tag -server`). For
  incremental runs, reuse `span-tag`'s on-the-fly dedup: `-server <solr>` +
  `-isi` queries the live index so a newly-arrived record inherits the right
  ISIL labels relative to what is already indexed. This keeps the daily delta
  consistent without a global recompute.
* **Deletions.** Fast update is insert/upsert only. Withdrawals/retractions are
  *not* handled incrementally by design; they are reconciled by the periodic
  full rebuild (or a later, separate `-delete` pass driven by a tombstone
  list). Call this out explicitly as a known gap.

### 6. State & scheduling

* `state.json` (path configurable), one entry per source:
  `{ "crossref": { "watermark": "2026-08-02", "last_run": "...", "docs": N } }`.
* Watermark is only advanced **after** a successful commit for that source, so
  a crash re-fetches the same window (fetch/import/index are all idempotent
  under overwrite-by-id).
* Scheduling stays external (cron/systemd timer): `span-fast-update -config
  fast-update.toml`. The tool is a one-shot; no daemon.

### 7. Reuse map — Go package vs shellout

| Stage    | Approach              | Source |
|----------|-----------------------|--------|
| fetch (crossref) | reuse code | `cmd/span-crossref-sync` logic → package, `-i d` window |
| fetch (OAI, e.g. DOAJ) | shellout | `../metha` (incremental by datestamp) |
| import   | **Go package**        | `span/formats` (what `span-import` wraps) |
| tag/license | **Go package**     | `span/tag` + `-unfreeze` FOLIO freeze |
| export   | **Go package**        | span export (`solr5vu3`, `-with-fullrecord`) |
| index    | **Go library** or shellout | `../solrbulk` (batch, `-commit`, retries) |
| freeze cfg | shellout (infrequent) | `span-freeze` against FOLIO (daily/weekly) |

First milestone can legitimately shell out to the existing `span-import` /
`span-tag` / `span-export` / `solrbulk` binaries to get an end-to-end path
quickly, then pull stages in-process one at a time.

### 8. Config (sketch)

```toml
state_file = "/var/lib/span/fast-update-state.json"
solr       = "http://localhost:8983/solr/biblio"
freeze     = "/etc/span/current.zip"   # FOLIO filterconfig for span-tag -unfreeze

[source.crossref]
enabled   = true
kind      = "crossref-sync"
cache_dir = "/data/span/crossref-sync"
dedup     = true                        # span-tag -server <solr> -isi

[source.doaj]
enabled  = false
kind     = "oai"
endpoint = "https://doaj.org/oai.article"
```

### 9. CLI (sketch)

```
span-fast-update -config fast-update.toml            # run all enabled sources
span-fast-update -config ... -source crossref        # one source
span-fast-update -config ... -dry-run                # fetch+transform, skip indexing
span-fast-update -config ... -since 2026-07-01       # override watermark
span-fast-update -config ... -no-commit              # index without SOLR commit
```

### 10. Full vs fast

* **Fast (daily):** this tool. Crossref delta (+ optionally DOAJ), tag, export,
  upsert. Minutes, not hours; small disk churn.
* **Full (periodic, e.g. weekly/on demand):** existing siskin pipeline
  (`AIUpdate`) — rebuilds snapshot, global dedup, handles deletions, replaces
  index. Fast update watermarks are reset to the rebuild date afterwards.

### 11. Milestones

1. **Skeleton + config + state** — parse config, load/save `state.json`, source
   registry, no-op run.
2. **Crossref end-to-end via shellout** — sync delta → span-import → span-tag
   → span-export → solrbulk; advance watermark. Proves the loop.
3. **Pull import/tag/export in-process** — call `formats`/`tag`/export packages
   directly; stream between stages; drop temp files.
4. **On-the-fly dedup** — wire `span-tag -server`/`-isi` equivalent against the
   live index.
5. **Second source (DOAJ via metha)** — validate the `Source` interface
   generalises.
6. **Ops** — structured run log, metrics (docs/sec, delta size), `-dry-run`,
   Makefile/nfpm packaging like the other `span-*` tools; docs.

### 12. Open questions

* Which sources beyond crossref truly need *daily* freshness vs. a weekly full
  refresh? (Likely only crossref + DOAJ; OLC/OSF/IOS/JSTOR are batchy.)
* Do we reuse `solrbulk` as a library or keep shelling out? (Lib gives one
  binary + shared retry/commit logic; shellout is simpler to start.)
* How often to refresh the FOLIO freeze relative to the daily run?
* Deletion strategy: rely solely on periodic full rebuild, or add a tombstone
  `-delete` pass later?
* Relationship to the planned **articleindexd** consolidation (merging siskin +
  span into one Go service) — is `span-fast-update` a stepping stone toward
  that service, or a component of it?
