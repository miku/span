package cli

import (
	"errors"
	"io"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/spf13/pflag"
)

// openInputs returns the concatenation of the named files, or stdin when
// there are none. The returned function closes all files.
func openInputs(names []string, stdin io.Reader) (io.Reader, func(), error) {
	if len(names) == 0 {
		return stdin, func() {}, nil
	}
	var (
		files   []*os.File
		readers []io.Reader
	)
	closeAll := func() {
		for _, f := range files {
			f.Close()
		}
	}
	for _, name := range names {
		f, err := os.Open(name)
		if err != nil {
			closeAll()
			return nil, nil, err
		}
		files = append(files, f)
		readers = append(readers, f)
	}
	return io.MultiReader(readers...), closeAll, nil
}

// profile holds the optional -cpuprofile and -memprofile flags.
type profile struct {
	cpu, mem string
}

func (p *profile) addFlags(fs *pflag.FlagSet, withMem bool) {
	fs.StringVar(&p.cpu, "cpuprofile", "", "write cpu profile to file")
	if withMem {
		fs.StringVar(&p.mem, "memprofile", "", "write heap profile to file (go tool pprof -png --alloc_objects program mem.pprof > mem.png)")
	}
}

// run calls f, recording a CPU profile around it and a heap profile after
// it, if requested.
func (p *profile) run(f func() error) (err error) {
	if p.cpu != "" {
		cf, err := os.Create(p.cpu)
		if err != nil {
			return err
		}
		defer cf.Close()
		if err := pprof.StartCPUProfile(cf); err != nil {
			return err
		}
		defer pprof.StopCPUProfile()
	}
	if err := f(); err != nil {
		return err
	}
	if p.mem == "" {
		return nil
	}
	mf, err := os.Create(p.mem)
	if err != nil {
		return err
	}
	runtime.GC()
	return errors.Join(pprof.WriteHeapProfile(mf), mf.Close())
}
