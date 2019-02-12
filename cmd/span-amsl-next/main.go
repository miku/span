package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"

	"github.com/sethgrid/pester"
	log "github.com/sirupsen/logrus"
)

var (
	live       = flag.String("live", "https://example.technology", "AMSL live base url")
	staging    = flag.String("staging", "https://example.technology", "AMSL staging base url")
	allowEmpty = flag.Bool("allow-empty", false, "allow empty responses from api")
)

// Discovery API response (now defunkt).
type Discovery struct {
	ContentFileLabel               string `json:"contentFileLabel,omitempty"`
	ContentFileURI                 string `json:"contentFileURI,omitempty"`
	DokumentLabel                  string `json:"DokumentLabel,omitempty"`
	DokumentURI                    string `json:"DokumentURI,omitempty"`
	EvaluateHoldingsFileForLibrary string `json:"evaluateHoldingsFileForLibrary"`
	ExternalLinkToContentFile      string `json:"externalLinkToContentFile,omitempty"`
	HoldingsFileLabel              string `json:"holdingsFileLabel,omitempty"`
	HoldingsFileURI                string `json:"holdingsFileURI,omitempty"`
	ISIL                           string `json:"ISIL"`
	LinkToContentFile              string `json:"linkToContentFile,omitempty"`
	LinkToHoldingsFile             string `json:"linkToHoldingsFile,omitempty"`
	MegaCollection                 string `json:"megaCollection"`
	ProductISIL                    string `json:"productISIL"`
	ShardLabel                     string `json:"shardLabel"`
	SourceID                       string `json:"sourceID"`
	TechnicalCollectionID          string `json:"technicalCollectionID"`
}

// MetadataUsage entry from metadata_usage and metadata_usage_concat endpoints.
type MetadataUsage struct {
	ISIL                  string `json:"ISIL"`
	MegaCollection        string `json:"megaCollection"`
	ProductISIL           string `json:"productISIL"`
	ShardLabel            string `json:"shardLabel"`
	SourceID              string `json:"sourceID"`
	TechnicalCollectionID string `json:"technicalCollectionID"`
}

// ContentFiles entry from contentfiles endpoint.
type ContentFiles struct {
	ContentFileLabel      string `json:"contentFileLabel"`
	ContentFileURI        string `json:"contentFileURI"`
	LinkToContentFile     string `json:"linkToContentFile"`
	MegaCollection        string `json:"megaCollection"`
	TechnicalCollectionID string `json:"technicalCollectionID"`
}

// HoldingsFiles entry from holdingsfiles endpoint.
type HoldingsFiles struct {
	DokumentLabel string `json:"DokumentLabel"`
	DokumentURI   string `json:"DokumentURI"`
	ISIL          string `json:"ISIL"`
	LinkToFile    string `json:"LinkToFile"`
}

// HoldingsFileConcat entry from holdings_file_concat endpoint.
type HoldingsFileConcat struct {
	ISIL                  string `json:"ISIL"`
	MegaCollection        string `json:"megaCollection"`
	ProductISIL           string `json:"productISIL"`
	ShardLabel            string `json:"shardLabel"`
	SourceID              string `json:"sourceID"`
	TechnicalCollectionID string `json:"technicalCollectionID"`
}

// SeparatedFields splits s on given separator and trims whitespace.
func SeparatedFields(s, sep string) (result []string) {
	for _, v := range strings.Split(s, sep) {
		result = append(result, strings.TrimSpace(v))
	}
	return
}

// ReaderCounter is counter for io.Reader
type ReaderCounter struct {
	count int64
	r     io.Reader
}

// NewReaderCounter function for create new ReaderCounter
func NewReaderCounter(r io.Reader) *ReaderCounter {
	return &ReaderCounter{r: r}
}

// Read keeps count.
func (counter *ReaderCounter) Read(buf []byte) (int, error) {
	n, err := counter.r.Read(buf)
	atomic.AddInt64(&counter.count, int64(n))
	return n, err
}

// Count function return counted bytes.
func (counter *ReaderCounter) Count() int64 {
	return atomic.LoadInt64(&counter.count)
}

// readFrom helper to decode data from reader into value, returning number ob
// bytes read.
func readFrom(v interface{}, r io.Reader) (int64, error) {
	rc := NewReaderCounter(r)
	if err := json.NewDecoder(rc).Decode(v); err != nil {
		return 0, err
	}
	return rc.Count(), nil
}

// MetadataUsageResponse group many metadata usage items.
type MetadataUsageResponse []MetadataUsage

// ReadFrom populates a response with MetadataUsage items.
func (resp *MetadataUsageResponse) ReadFrom(r io.Reader) (int64, error) {
	return readFrom(resp, r)
}

// ContentFilesResponse group many content files items.
type ContentFilesResponse []ContentFiles

// ReadFrom populates a response with MetadataUsage items.
func (resp *ContentFilesResponse) ReadFrom(r io.Reader) (int64, error) {
	return readFrom(resp, r)
}

// HoldingsFiles group many holdings files items.
type HoldingsFilesResponse []HoldingsFiles

// ReadFrom populates a response with MetadataUsage items.
func (resp *HoldingsFilesResponse) ReadFrom(r io.Reader) (int64, error) {
	return readFrom(resp, r)
}

// HoldingsFileConcatResponse group many holdings files (concat) items.
type HoldingsFileConcatResponse []HoldingsFileConcat

// ReadFrom populates a response with MetadataUsage items.
func (resp *HoldingsFileConcatResponse) ReadFrom(r io.Reader) (int64, error) {
	return readFrom(resp, r)
}

// fetchLocation fetches a HTTP location into a io.ReaderFrom.
func fetchLocation(location string, r io.ReaderFrom) (int64, error) {
	log.Printf("fetching %s", location)
	resp, err := pester.Get(location)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return r.ReadFrom(resp.Body)
}

// fetchFrom constructs a link and fetches the response into given io.ReaderFrom.
func fetchFrom(base, kind string, r io.ReaderFrom) (int64, error) {
	loc := fmt.Sprintf("%s/outboundservices/list?do=%s", base, kind)
	return fetchLocation(loc, r)
}

func main() {
	flag.Parse()

	var (
		mur MetadataUsageResponse
		hcr HoldingsFileConcatResponse
		hfr HoldingsFilesResponse
		cfr ContentFilesResponse
	)

	// Where we get the data from.
	fetchlist := []struct {
		base string        // This can be live or staging.
		kind string        // Name of the query.
		r    io.ReaderFrom // Struct wrapped in a ReaderFrom.
	}{
		{*live, "metadata_usage", &mur},
		{*staging, "holdings_file_concat", &hcr}, // TODO(miku): Change this to *live.
		{*live, "holdingsfiles", &hfr},
		{*live, "contentfiles", &cfr},
	}

	for _, ff := range fetchlist {
		if _, err := fetchFrom(ff.base, ff.kind, ff.r); err != nil {
			log.Fatal(err)
		}
	}

	// For debugging.
	sizes := map[string]int{
		"metadata_usage":       len(mur),
		"holdings_file_concat": len(hcr),
		"holdingsfiles":        len(hfr),
		"contentfiles":         len(cfr),
	}
	for k, v := range sizes {
		log.Printf("%s: %d", k, v)
		if v == 0 && !*allowEmpty {
			log.Fatal("empty response from %s", k)
		}
	}

	// This should contain an approximation of the discovery endpoint.
	var updates []Discovery

	for i, mu := range mur {
		if i%10000 == 0 {
			log.Printf("%d of %d done", i, len(mur))
		}
		if mu.MegaCollection == "" {
			log.Printf("skipping empty megaCollection in #L%d", i)
			continue
		}

		// Defunkt update.
		update := Discovery{
			ISIL:                  mu.ISIL,
			MegaCollection:        mu.MegaCollection,
			ProductISIL:           mu.ProductISIL,
			ShardLabel:            mu.ShardLabel,
			SourceID:              mu.SourceID,
			TechnicalCollectionID: mu.TechnicalCollectionID,
		}
		// Merge fields from content file.
		for _, cf := range cfr {
			if cf.MegaCollection != mu.MegaCollection {
				continue
			}
			update.ContentFileLabel = cf.ContentFileLabel
			update.ContentFileURI = cf.ContentFileURI
			update.LinkToContentFile = cf.LinkToContentFile
			break
		}

		// Default is negative.
		update.EvaluateHoldingsFileForLibrary = "no"

		// Incorporate new holdings file concat response.
		for _, hc := range hcr {
			if hc.MegaCollection != mu.MegaCollection {
				continue
			}
			// ISIL is a list, separated by semicolons.
			for _, isil := range SeparatedFields(hc.ISIL, ";") {
				if isil != mu.ISIL {
					continue
				}
				update.ProductISIL = hc.ProductISIL
				update.ShardLabel = hc.ShardLabel
				break
			}
			update.EvaluateHoldingsFileForLibrary = "yes"
		}

		// Add link to content file.
		if update.ContentFileURI != "" {
			if strings.HasPrefix(update.ContentFileURI, "http://amsl") {
				update.LinkToContentFile = fmt.Sprintf(
					"%s/OntoWiki/files/get?setResource=%s", *live, update.ContentFileURI)
			}
		}

		// No holding file required? Next item.
		if update.EvaluateHoldingsFileForLibrary == "no" {
			updates = append(updates, update)
			continue
		}

		for _, hf := range hfr {
			if hf.ISIL != mu.ISIL {
				continue
			}
			// Create a new item for each holding file.
			ndoc := Discovery{
				ContentFileLabel:               update.ContentFileLabel,
				ContentFileURI:                 update.ContentFileURI,
				EvaluateHoldingsFileForLibrary: update.EvaluateHoldingsFileForLibrary,
				ISIL:                           update.ISIL,
				LinkToContentFile:              update.LinkToContentFile,
				MegaCollection:                 update.MegaCollection,
				ProductISIL:                    update.ProductISIL,
				ShardLabel:                     update.ShardLabel,
				SourceID:                       update.SourceID,
				TechnicalCollectionID:          update.TechnicalCollectionID,
				LinkToHoldingsFile:             hf.LinkToFile,
				DokumentLabel:                  hf.DokumentLabel,
				DokumentURI:                    hf.DokumentURI,
			}
			if hf.DokumentURI != "" {
				ndoc.LinkToHoldingsFile = fmt.Sprintf(
					"%s/OntoWiki/files/get?setResource=%s", *live, hf.DokumentURI)
			}
			updates = append(updates, ndoc)
		}
	}

	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()

	if err := json.NewEncoder(bw).Encode(updates); err != nil {
		log.Fatal(err)
	}
}
