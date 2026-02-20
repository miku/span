## Migrate fmt.Errorf %v to %w for error wrapping

### priority

2

### type

chore

Use `%w` instead of `%v` in `fmt.Errorf` calls to preserve error chains. Also adopt `errors.Is`, `errors.As` (for `span.Skip`), and `errors.Join` where applicable.

---

## Migrate logging from logrus to log/slog

### priority

2

### type

chore

Replace logrus and stdlib log usage with `log/slog` (Go 1.21). 21 files import logrus, ~15 use stdlib log. Consolidate the logrus fork (`lytics/logrus` vs `sirupsen/logrus`) as part of this migration. Can be done package-by-package.

---

## Replace interface{} with any and add generics where useful

### priority

3

### type

chore

Rename 81 occurrences of `interface{}` to `any` across 23 files. Consider generic `Set[T comparable]` to replace `container.StringSet` and a typed `FormatMap` in span-import.

---

## Adopt slices and maps packages

### priority

3

### type

chore

Replace manual sort/collect patterns with `slices.Sort`, `slices.Collect`, `maps.Keys` etc. Applies to `container/string.go`, `cmd/span-import`, `formats/finc/intermediate.go`, `filter/filter.go`.

---

## Add context.Context to long-running operations

### priority

2

### type

feature

Add `context.Context` to `parallel.Processor`, `crossref.CreateSnapshot`, and HTTP calls in `folio/folio.go` to enable cancellation, timeouts, and tracing.

---

## Refactor filter registry boilerplate

### priority

3

### type

chore

Replace the 70-line switch in `filter/filter.go:unmarshalFilter` with a registry map to reduce boilerplate and make adding filters simpler.

---

## Refactor crossref/snapshot.go into separate stages

### priority

2

### type

chore

Split the 700+ line `crossref/snapshot.go` into `stage1.go`, `stage2.go`, `stage3.go`. Replace custom worker pool with `errgroup.Group` for structured concurrency. Add `context.Context` for cancellation.

---

## Fix data race in parallel/processor.go

### priority

1

### type

fix

Multiple goroutines write to `wErr` without synchronization. Use `errgroup` or `atomic.Pointer[error]`. Also fix `BytesBatch.Reset()` to use `bb.b = bb.b[:0]` instead of `bb.b = nil` to retain backing array.

---

## Inject version via ldflags instead of manual patching

### priority

3

### type

chore

Version string lives in Makefile, `common.go`, and packaging files. Use `-ldflags "-X ..."` at build time so only the Makefile needs updating.

---

## Modernize tests with subtests and benchmarks

### priority

3

### type

chore

Add `t.Run()` for subtests, use `t.Fatalf` for setup failures, add benchmarks for `parallel/`, `encoding/`, and `filter/` packages.

---

## Triage TODO/FIXME/XXX comments

### priority

3

### type

chore

64 TODO/FIXME/XXX comments across 34 files. Close stale ones, convert actionable ones to issues.

---

## Replace deprecated golint with staticcheck in Makefile

### priority

2

### type

fix

`golint` at `Makefile:53` is deprecated. Replace with `staticcheck` or `golangci-lint`.

---

## Remove obsolete KeyLengthLimit const

### priority

3

### type

chore

`KeyLengthLimit` in `common.go:34` is marked obsolete in a comment. Remove it.
