// The span-amsl-discovery tool will create a discovery (now defunkt) like API
// response from available AMSL endpoints, refs #14456, #14415.
package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/sethgrid/pester"
	log "github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
)

var (
	live       = flag.String("live", "https://example.technology", "AMSL live base url")
	allowEmpty = flag.Bool("allow-empty", false, "allow empty responses from api")
	flatten    = flag.Bool("f", false, "flatten output into a TSV")
	dbFile     = flag.String("db", "", "write data into an sqlite3 database")
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

// readFrom decodes data from reader into value, returning bytes read.
func readFrom(r io.Reader, v interface{}) (int64, error) {
	rc := span.NewReaderCounter(r)
	if err := json.NewDecoder(rc).Decode(v); err != nil {
		return 0, err
	}
	return rc.Count(), nil
}

// MetadataUsageResponse group many metadata usage items.
type MetadataUsageResponse []MetadataUsage

// ReadFrom populates a response with MetadataUsage items.
func (resp *MetadataUsageResponse) ReadFrom(r io.Reader) (int64, error) {
	return readFrom(r, resp)
}

// ContentFilesResponse group many content files items.
type ContentFilesResponse []ContentFiles

// ReadFrom populates a response with MetadataUsage items.
func (resp *ContentFilesResponse) ReadFrom(r io.Reader) (int64, error) {
	return readFrom(r, resp)
}

// HoldingsFiles group many holdings files items.
type HoldingsFilesResponse []HoldingsFiles

// ReadFrom populates a response with MetadataUsage items.
func (resp *HoldingsFilesResponse) ReadFrom(r io.Reader) (int64, error) {
	return readFrom(r, resp)
}

// HoldingsFileConcatResponse group many holdings files (concat) items.
type HoldingsFileConcatResponse []HoldingsFileConcat

// ReadFrom populates a response with MetadataUsage items.
func (resp *HoldingsFileConcatResponse) ReadFrom(r io.Reader) (int64, error) {
	return readFrom(r, resp)
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

// slugifyTabs removes tabs from all fields.
func slugifyTabs(vs []string) (result []string) {
	for _, v := range vs {
		c := strings.ReplaceAll(v, "\t", " ")
		result = append(result, c)
		if v != c {
			log.Printf("[note] removed tab from %s", v)
		}
	}
	return result
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
The generated TSV (via -f) fields are:

1	ShardLabel
2	ISIL
3	SourceID
4	TechnicalCollectionID
5	MegaCollection
6	HoldingsFileURI
7	HoldingsFileLabel
8	LinkToHoldingsFile
9	EvaluateHoldingsFileForLibrary
10	ContentFileURI
11	ContentFileLabel
12	LinkToContentFile
13	ExternalLinkToContentFile
14	ProductISIL
15	DokumentURI
16	DokumentLabel
`)
	}
	flag.Parse()

	var (
		mur MetadataUsageResponse
		hcr HoldingsFileConcatResponse
		hfr HoldingsFilesResponse
		cfr ContentFilesResponse
	)

	// Where we get the data from.
	fetchlist := []struct {
		base string        // The AMSL base URL.
		kind string        // Name of the query.
		r    io.ReaderFrom // Struct wrapped in a ReaderFrom.
	}{
		{*live, "metadata_usage", &mur},
		{*live, "holdings_file_concat", &hcr},
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
			log.Fatalf("empty response from %s", k)
		}
	}

	var (
		// This should contain an approximation of the discovery endpoint.
		updates []Discovery
		// Mismatch, where we expected a holding file, but got none, refs #16207.
		mismatched []Discovery
	)

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
				update.EvaluateHoldingsFileForLibrary = "yes"
				// TODO: It might be, that we have a holdings_file_concat
				// mention, but no holding file, later, refs #16207.
				break
			}
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

		// Will be true, if we actually find holding files, refs #16207.
		var resolved bool

		for _, hf := range hfr {
			if hf.ISIL != mu.ISIL {
				continue
			}
			// Create a new item for each holding file.
			ndoc := Discovery{
				DokumentLabel:                  hf.DokumentLabel,
				DokumentURI:                    hf.DokumentURI,
				LinkToHoldingsFile:             hf.LinkToFile,
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
			}
			if hf.DokumentURI != "" {
				ndoc.LinkToHoldingsFile = fmt.Sprintf(
					"%s/OntoWiki/files/get?setResource=%s", *live, hf.DokumentURI)
			}
			updates = append(updates, ndoc)
			resolved = true
		}
		if !resolved {
			log.Printf("mismatch: %s [%s] %s", update.ISIL, update.SourceID, update.MegaCollection)
			mismatched = append(mismatched, update)
		}
	}
	log.Printf("warning: %d mismatched items", len(mismatched))

	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()

	switch {
	case *flatten:
		for _, update := range updates {
			fields := []string{
				update.ShardLabel,
				update.ISIL,
				update.SourceID,
				update.TechnicalCollectionID,
				update.MegaCollection,
				update.HoldingsFileURI,
				update.HoldingsFileLabel,
				update.LinkToHoldingsFile,
				update.EvaluateHoldingsFileForLibrary,
				update.ContentFileURI,
				update.ContentFileLabel,
				update.LinkToContentFile,
				update.ExternalLinkToContentFile,
				update.ProductISIL,
				update.DokumentURI,
				update.DokumentLabel,
			}
			fields = slugifyTabs(fields)
			if _, err := io.WriteString(bw, strings.Join(fields, "\t")+"\n"); err != nil {
				log.Fatal(err)
			}
		}
	case *dbFile != "":
		createTable := `
			create table amsl (
				shard text not null,
				isil text not null,
				sid text not null,
				tcid text not null,
				mc text not null,
				hfuri text,
				hflabel text,
				hflink text,
				hfeval text,
				cfuri text,
				cflabel text,
				cflink text,
				cfelink text,
				pisil text,
				docuri text,
				doclabel text
			);

			create index amsl_isil on amsl(isil);
			create index amsl_isil_sid on amsl(isil, sid);
			create index amsl_isil_sid_mc on amsl(isil, sid, mc);
			create index amsl_mc on amsl(mc);
			create index amsl_sid on amsl(sid);
			create index amsl_sid_mc on amsl(sid, mc);
			create index amsl_sid_tcid on amsl(sid, tcid);
			create index amsl_tcid on amsl(tcid);
		`
		db, err := sql.Open("sqlite3", *dbFile)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		_, err = db.Exec(createTable)
		if err != nil {
			log.Printf("%q: %s\n", err, createTable)
			return
		}
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		stmt, err := tx.Prepare(`insert into amsl (shard, isil, sid, tcid, mc,
			hfuri, hflabel, hflink, hfeval, cfuri, cflabel, cflink, cfelink,
			pisil, docuri, doclabel) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()
		started := time.Now()
		for i, update := range updates {
			if i%100000 == 0 {
				log.Printf("@%d in %s", i, time.Since(started))
			}
			fields := []interface{}{
				update.ShardLabel,
				update.ISIL,
				update.SourceID,
				update.TechnicalCollectionID,
				update.MegaCollection,
				update.HoldingsFileURI,
				update.HoldingsFileLabel,
				update.LinkToHoldingsFile,
				update.EvaluateHoldingsFileForLibrary,
				update.ContentFileURI,
				update.ContentFileLabel,
				update.LinkToContentFile,
				update.ExternalLinkToContentFile,
				update.ProductISIL,
				update.DokumentURI,
				update.DokumentLabel,
			}
			_, err = stmt.Exec(fields...)
			if err != nil {
				log.Fatal(err)
			}
		}
		tx.Commit()
		log.Printf("@%d in %s", len(updates), time.Since(started))
	default:
		if err := json.NewEncoder(bw).Encode(updates); err != nil {
			log.Fatal(err)
		}
	}
}
