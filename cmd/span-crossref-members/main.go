// span-crossref-members fetches crossref members api. It will merely paginate
// through the api responses and will output one response per line.  Create
// mapping from DOI to name: span-crossref-members | jq -rc '.message.items[].prefix[] |
// {(.value|tostring): .name}' | jq -s add > assets/crossref/members.json
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/internal/cmd/crossrefcmd"
)

var (
	offset     = flag.Int("offset", 0, "offset")
	rows       = flag.Int("rows", 20, "rows to fetch per request")
	base       = flag.String("base", "http://api.crossref.org/members", "base url")
	sleep      = flag.Duration("sleep", 1*time.Second, "time to sleep between requests")
	silent     = flag.Bool("q", false, "suppress logging output")
	version    = flag.Bool("version", false, "output version")
	retryCount = flag.Int("retry", 3, "retry count on HTTP 500 and similar errors")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-crossref-members [options]\n\n")
		fmt.Fprintf(os.Stderr, "Paginates through the crossref members API and emits one JSON response per\n")
		fmt.Fprintf(os.Stderr, "line. Useful for building DOI-prefix to publisher-name mappings.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *silent {
		log.SetOutput(io.Discard)
	}
	if *version {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	cfg := crossrefcmd.MembersConfig{
		Offset:     *offset,
		Rows:       *rows,
		Base:       *base,
		Sleep:      *sleep,
		RetryCount: *retryCount,
	}
	if err := crossrefcmd.RunMembers(cfg, nil, w); err != nil {
		log.Fatal(err)
	}
}
