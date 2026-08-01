package crossrefcmd

import (
	"io"
	"runtime"
	"strings"

	"github.com/miku/span/crossref"
)

// FastSnapshotConfig mirrors the span-crossref-fast-snapshot flags. InputFiles
// and Excludes are the resolved values (main reads the excludes file into
// Excludes).
type FastSnapshotConfig struct {
	InputFiles        []string
	OutputFile        string
	BatchSize         int
	NumWorkers        int
	Verbose           bool
	KeepTempFiles     bool
	SortBufferSize    string
	Excludes          []string
	ShuffleInputFiles bool
	CacheEnabled      bool
	CacheDir          string
	CacheClear        bool
}

// DefaultFastSnapshotConfig returns the default configuration for
// RunFastSnapshot.
func DefaultFastSnapshotConfig() FastSnapshotConfig {
	return FastSnapshotConfig{
		OutputFile:     crossref.DefaultOutputFile,
		BatchSize:      100000,
		NumWorkers:     runtime.NumCPU(),
		SortBufferSize: "25%",
		CacheEnabled:   true,
	}
}

// SnapshotFunc creates a snapshot from options. It is injected so option
// mapping can be tested without running the (heavy) real snapshot.
type SnapshotFunc func(crossref.SnapshotOptions) error

// ParseExcludes reads DOIs to exclude, one per line, from r. As in the
// original, the result is the newline-split content (a trailing empty element
// is possible).
func ParseExcludes(r io.Reader) ([]string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(b), "\n"), nil
}

// Options maps the config to crossref.SnapshotOptions.
func (c FastSnapshotConfig) Options() crossref.SnapshotOptions {
	return crossref.SnapshotOptions{
		InputFiles:        c.InputFiles,
		OutputFile:        c.OutputFile,
		BatchSize:         c.BatchSize,
		NumWorkers:        c.NumWorkers,
		Verbose:           c.Verbose,
		KeepTempFiles:     c.KeepTempFiles,
		SortBufferSize:    c.SortBufferSize,
		Excludes:          c.Excludes,
		ShuffleInputFiles: c.ShuffleInputFiles,
		CacheEnabled:      c.CacheEnabled,
		CacheDir:          c.CacheDir,
		CacheClear:        c.CacheClear,
	}
}

// RunFastSnapshot creates the snapshot using snapshot (defaults to
// crossref.CreateSnapshot when nil).
func RunFastSnapshot(cfg FastSnapshotConfig, snapshot SnapshotFunc) error {
	if snapshot == nil {
		snapshot = crossref.CreateSnapshot
	}
	return snapshot(cfg.Options())
}
