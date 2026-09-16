# Cleanup Notes

> Code and design cleanup opportunities for span. The codebase targets Go 1.24
> but largely uses patterns from the Go 1.11-1.17 era. Most items below are
> incremental and can be adopted file-by-file.

## 1. Error handling

Only 3 files in the entire codebase use `%w` for error wrapping; zero use of
`errors.Is`, `errors.As`, or `errors.Join`.

**Wrap with `%w`** -- most `fmt.Errorf` calls pass errors via `%v`, which
discards the error chain. Wrapping lets callers inspect causes when needed.

```go
// before
return fmt.Errorf("error in stage 1: %v", err)

// after
return fmt.Errorf("stage 1: %w", err)
```

**Use `errors.Join`** -- `crossref/snapshot.go` and other places with
multi-stage cleanup return only the first error. `errors.Join` (Go 1.20)
preserves all of them.

**Use `errors.As` for Skip** -- `span.Skip` is type-asserted in several `cmd/`
files via `if _, ok := err.(span.Skip); ok`. The idiomatic replacement is
`errors.As`, which works through wrapped errors.

## 2. Structured logging with `log/slog`

21 files import logrus, ~15 more use stdlib `log`. Some cmd files alias both
under the same name. `log/slog` (Go 1.21) is now the standard approach and
removes the external dependency.

```go
// before
log.Printf("[%s] fetching: %s", is.ID, link)

// after
slog.Info("fetching", "id", is.ID, "link", link)
```

Migrating one package at a time is fine; logrus and slog can coexist during
transition.

## 3. Replace `interface{}` with generics or `any`

81 occurrences of `interface{}` across 23 files.

**Quick win** -- rename `interface{}` to `any` (Go 1.18 alias). No behavior
change, just readability.

**`container.StringSet` -> `Set[T comparable]`** -- the current type is only
usable with strings and reimplements iteration, intersection, and difference by
hand. A generic set backed by `map[T]struct{}` covers all uses and removes
duplicate set logic elsewhere.

**`FormatMap` in span-import** -- currently `map[string]func() interface{}`.
Could become `map[string]func() IntermediateSchemaer` or use a small generic
registry, removing the type assertions in `processXML`/`processJSON`.

## 4. Adopt `slices` and `maps` packages

Neither `slices` nor `maps` are imported anywhere. Several patterns would
benefit:

```go
// container/string.go -- before
func (set *StringSet) SortedValues() (values []string) {
    for k := range set.Set {
        values = append(values, k)
    }
    sort.Strings(values)
    return values
}

// after
func (set *StringSet) SortedValues() []string {
    v := slices.Collect(maps.Keys(set.Set))
    slices.Sort(v)
    return v
}
```

Also applies to: `cmd/span-import` (sorting format names),
`formats/finc/intermediate.go` (dedup helpers), `filter/filter.go` (key
extraction).

## 5. `context.Context` for long-running operations

Neither `parallel.Processor` nor `crossref.CreateSnapshot` accept a context.
Adding it enables cancellation, timeouts, and tracing without changing the
public API shape much (context is typically the first parameter).

`folio/folio.go` makes HTTP requests without context -- `http.NewRequestWithContext`
is the standard replacement.

## 6. Filter registry boilerplate

`filter/filter.go:unmarshalFilter` is a 70-line switch that repeats the same
unmarshal-and-return pattern 11 times. A small registry map would collapse this:

```go
var registry = map[string]func() Filter{
    "any":        func() Filter { return &AnyFilter{} },
    "doi":        func() Filter { return &DOIFilter{} },
    // ...
}

func unmarshalFilter(name string, raw json.RawMessage) (Filter, error) {
    ctor, ok := registry[name]
    if !ok {
        return nil, fmt.Errorf("unknown filter: %s", name)
    }
    f := ctor()
    if err := json.Unmarshal(raw, f); err != nil {
        return nil, err
    }
    return f, nil
}
```

## 7. `crossref/snapshot.go` complexity

At 700+ lines with three processing stages, external tool invocations (`sort`,
`zstd`, `filterline`), and manual concurrency, this is the most complex file in
the project.

Possible steps:

* Split into `stage1.go`, `stage2.go`, `stage3.go` within the same package.
* Replace the custom worker pool with `errgroup.Group` for structured
  concurrency and proper error propagation.
* Consider `context.Context` for cancellation (snapshot creation can take hours).

The shell pipeline in stage 2 is intentional (leveraging coreutils sort
parallelism) and probably not worth replacing, but it could be behind an
interface so it's testable.

## 8. `parallel/processor.go` concurrency

The processor silently races on `wErr` -- multiple goroutines write to the same
error variable without synchronization. The comment says "we don't care about
synchronisation" but this is technically a data race. Using `errgroup` or an
`atomic.Pointer[error]` would fix it cleanly.

Also: `BytesBatch.Reset()` sets `bb.b = nil` -- `bb.b = bb.b[:0]` retains the
backing array and avoids re-allocation.

## 9. Version management

The version string lives in three places: `Makefile:2`, `common.go:31`, and
`packaging/*/control|spec`. The Makefile has an `update-version` target that
patches them with sed.

An alternative: inject at build time via `-ldflags`:

```makefile
$(TARGETS): %: cmd/%/main.go
    go build -ldflags "-X github.com/miku/span.AppVersion=$(VERSION)" -o $@ $<
```

This reduces the places to update to one (the Makefile) and makes `common.go`
the single declaration site with a default value.

## 10. Test modernization

* No use of `t.Run()` for subtests -- adding it gives named sub-cases and
  allows `-run` filtering.
* Table-driven tests use `t.Errorf` instead of `t.Fatalf` for setup failures --
  a failed setup should not continue to assertions.
* No benchmarks outside of `crossref/` -- `parallel/`, `encoding/`, and
  `filter/` are performance-sensitive and would benefit from them.

## 11. Minor items

| Item | Location | Notes |
|------|----------|-------|
| 64 TODO/FIXME/XXX comments | 34 files | Triage: close stale ones, convert actionable ones to issues |
| `sort.Strings` → `slices.Sort` | 6+ call sites | Drop `sort` import where possible |
| `io.ReadAll` available | several files still use `ioutil` patterns | Already migrated, but double-check |
| `range N` (Go 1.22) | `for i := 0; i < p.NumWorkers; i++` | Minor, but nice in new code |
| Unused `KeyLengthLimit` const | `common.go:34` | Comment says "obsolete"; remove |
| `golint` in Makefile | `Makefile:53` | Deprecated; replace with `staticcheck` or `golangci-lint` |
| logrus fork | `github.com/lytics/logrus` imported alongside `github.com/sirupsen/logrus` | Consolidate to one (or migrate both to slog) |
