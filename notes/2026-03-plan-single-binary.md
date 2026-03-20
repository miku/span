# Plan: Consolidate span-* into a single `span` binary

## Motivation

Currently 24 separate executables are built, packaged, and deployed. This
creates maintenance overhead (Makefile, nfpm.yaml, man pages) and makes
discoverability hard for users. A single `span` binary with subcommands
(`span import`, `span query`, `span tag`, ...) is easier to distribute,
document, and extend.

## Current inventory

| Tool | Purpose | ~LOC | Subcommand |
|------|---------|------|------------|
| span-import | convert source formats to intermediate schema | 317 | `span import` |
| span-export | intermediate schema to SOLR/other formats | 133 | `span export` |
| span-tag | apply filter config to tag records | 275 | `span tag` |
| span-query | read-only queries against index/files | 988 | `span query` |
| span-index | prebuilt SOLR queries (absorbed into span-query) | 504 | **remove** |
| span-report | ISSN/date reports (absorbed into span-query) | 374 | **remove** |
| span-compare | compare two SOLR indices | 476 | `span compare` |
| span-compare-file | compare file vs index (absorbed into span-query) | 307 | **remove** |
| span-crossref-sync | download crossref works API | 389 | `span crossref sync` |
| span-crossref-snapshot | extract/dedupe crossref data | 359 | `span crossref snapshot` |
| span-crossref-fast-snapshot | dedupe daily slices | 93 | `span crossref fast-snapshot` |
| span-crossref-fastproc | daily slice to SOLR-importable | 221 | `span crossref fastproc` |
| span-crossref-members | paginate crossref members API | 171 | `span crossref members` |
| span-crossref-table | TSV from crossref data | 50 | `span crossref table` |
| span-folio | query FOLIO API | 144 | `span folio` |
| span-freeze | freeze URLs into zip | 214 | `span freeze` |
| span-amsl-discovery | create discovery API response | 467 | `span amsl` |
| span-hcov | KBART coverage report | 132 | `span hcov` |
| span-oa-filter | set OA flag from KBART | 189 | `span oa-filter` |
| span-redact | clear fulltext field | 77 | `span redact` |
| span-doisniffer | sniff DOI from SOLR docs | 102 | `span doisniffer` |
| span-local-data | extract fields from JSON | 56 | `span local-data` |
| span-mail | compose/send email | 85 | `span mail` |
| span-update-labels | update x.labels from TSV | 100 | `span update-labels` |

## Subcommand grouping

```
span
  import              convert source formats to intermediate schema
  export              intermediate schema to SOLR/other formats
  tag                 apply filter config for tagging/dedup
  query               read-only queries against index and files
  compare             compare two SOLR indices
  redact              clear fulltext from intermediate schema
  freeze              freeze URLs and content into zip

  crossref
    sync              download from Crossref Works API
    snapshot          extract/dedupe crossref records
    fast-snapshot     dedupe accumulated daily slices
    fastproc          daily slice to SOLR-importable
    members           paginate Crossref Members API
    table             TSV from crossref data

  folio               query FOLIO API
  amsl                create discovery API response
  hcov                KBART coverage report
  oa-filter           set OA flag from KBART
  doisniffer          sniff DOI from SOLR docs
  local-data          extract fields from JSON file
  mail                compose/send email
  update-labels       update x.labels from TSV
```

## Already consolidated (can be removed after transition)

These tools are already absorbed into `span-query`:

- **span-index** -- all tasks moved to span-query
- **span-report** -- issn-report task in span-query
- **span-compare-file** -- isil-file and isil-compare tasks in span-query

## Transition plan

### Phase 1: Framework setup

1. Add `cobra` dependency: `go get github.com/spf13/cobra`
2. Create `cmd/span/main.go` with root command
3. Create `cmd/span/cmd/` directory for subcommand files
4. Start with one subcommand (`query`) to validate the pattern

### Phase 2: Migrate subcommands (incremental)

Migrate tools in batches, ordered by risk (lowest first):

**Batch 1 -- Small, self-contained tools:**
- redact, local-data, crossref-table, doisniffer, mail, hcov, update-labels
- These are small (<150 LOC), have few dependencies, low blast radius

**Batch 2 -- Core pipeline tools:**
- import, export, tag
- These are the data pipeline backbone; test thoroughly

**Batch 3 -- Index/query tools:**
- query (already consolidated), compare
- Remove span-index, span-report, span-compare-file source dirs

**Batch 4 -- Crossref tools:**
- crossref sync, snapshot, fast-snapshot, fastproc, members
- Group under `span crossref` subcommand tree

**Batch 5 -- Integration tools:**
- folio, amsl, freeze, oa-filter

### Phase 3: Backwards compatibility

During transition, keep old binaries working:

1. **Symlink approach**: package installs `span` binary plus symlinks
   (`span-import -> span`). When invoked as `span-import`, the binary
   detects `os.Args[0]` and dispatches to the `import` subcommand.
   This is the busybox pattern.
2. **Deprecation period**: keep symlinks for 2-3 releases, print a
   stderr warning ("span-import is deprecated, use: span import")
3. **Remove symlinks**: after deprecation period, remove from nfpm.yaml

### Phase 4: Cleanup

1. Remove individual `cmd/span-*` directories
2. Simplify Makefile to build single binary
3. Update nfpm.yaml: one binary, optional symlinks
4. Update man pages / documentation
5. Update any scripts/automation that reference old binary names

## Build changes

**Before (Makefile):**
```make
TARGETS = span-import span-export span-tag ... (24 targets)
$(TARGETS): %: cmd/%/main.go
    go build ... -o $@ $<
```

**After (Makefile):**
```make
span: cmd/span/main.go cmd/span/cmd/*.go
    go build ... -o $@ ./cmd/span/
```

## Packaging changes (nfpm.yaml)

**Before:** 24 binary entries.

**After:**
```yaml
contents:
  - src: span
    dst: /usr/local/bin/span
    file_info:
      mode: 0755
  # Compatibility symlinks (remove after deprecation)
  - src: /usr/local/bin/span
    dst: /usr/local/bin/span-import
    type: symlink
  # ... etc
```

## Risks and mitigations

| Risk | Mitigation |
|------|------------|
| Scripts break that call `span-import` etc. | Busybox/symlink pattern for backwards compat |
| Flag incompatibilities (cobra vs stdlib flag) | Keep same flag names; cobra supports POSIX and Go-style |
| Large single binary | Still small by Go standards; upx compression available |
| cobra dependency | Well-maintained, widely used, minimal transitive deps |
| Merge conflicts during transition | Migrate one batch at a time, merge each before starting next |

## Open questions

- Should we keep `span-query`'s `-t task` pattern or convert tasks to
  sub-subcommands (`span query numdocs`)? The latter is more cobra-native
  but `-t` with prefix matching is convenient for interactive use. Could
  support both.
- Some tools (span-crossref-sync) are long-running daemons. Does cobra's
  model fit, or do we need special handling?
- The systemd unit `span-webhookd.service` references a binary that isn't
  in the current cmd/ tree. Investigate whether this is still used.
