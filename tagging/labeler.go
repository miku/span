package tagging

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrg/xdg"
	"github.com/jmoiron/sqlx"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/licensing"
	"github.com/sethgrid/pester"
)

const (
	// SLUBEZBKBART link to DE-14 KBART, to be included across all sources.
	SLUBEZBKBART         = "https://dbod.de/SLUB-EZB-KBART.zip"
	DE15FIDISSNWHITELIST = "DE15FIDISSNWHITELIST"
)

// ConfigRow describes a single entry (e.g. an attachment request) from AMSL.
type ConfigRow struct {
	ShardLabel                     string
	ISIL                           string
	SourceID                       string
	TechnicalCollectionID          string
	MegaCollection                 string
	HoldingsFileURI                string
	HoldingsFileLabel              string
	LinkToHoldingsFile             string
	EvaluateHoldingsFileForLibrary string
	ContentFileURI                 string
	ContentFileLabel               string
	LinkToContentFile              string
	ExternalLinkToContentFile      string
	ProductISIL                    string
	DokumentURI                    string
	DokumentLabel                  string
}

// Labeler updates an intermediate schema document.  We need mostly: ISIL,
// SourceID, MegaCollection, TechnicalCollectionID, HoldFileURI,
// EvaluateHoldingsFileForLibrary
type Labeler struct {
	dbFile string // sqlite filename
	db     *sqlx.DB

	cache          map[string][]ConfigRow
	hfcache        *HFCache
	whitelistCache map[string]map[string]struct{} // Name (e.g. DE15FIDISSNWHITELIST) -> Set (a set of ISSN)

	cacheKeyFunc func(doc *finc.IntermediateSchema) string
}

// New returns a initialized labeler using a relational AMSL representation
// (sqlite). Will fail here, if database cannot be opened.
func New(dbFile string) (*Labeler, error) {
	db, err := sqlx.Connect("sqlite3", fmt.Sprintf("%s?ro=1", dbFile))
	if err != nil {
		return nil, err
	}
	return &Labeler{
		dbFile: dbFile,
		db:     db,
		hfcache: &HFCache{
			forceDownload: true, // TODO: pass this option.
			cacheHome:     filepath.Join(xdg.CacheHome, "span"),
			entries:       make(map[string]map[string][]licensing.Entry),
		},
		cache: make(map[string][]ConfigRow),
		cacheKeyFunc: func(doc *finc.IntermediateSchema) string {
			v := doc.MegaCollections
			sort.Strings(v)
			return doc.SourceID + "@" + strings.Join(v, "@")
		},
	}, nil
}

// matchingRows returns a list of relevant rows for a given document. This is a
// prefilter (going from 200K+ rows 10s of rows).
func (l *Labeler) matchingRows(doc *finc.IntermediateSchema) (result []ConfigRow, err error) {
	key := l.cacheKeyFunc(doc)
	if v, ok := l.cache[key]; ok {
		return v, nil
	}
	if len(doc.MegaCollections) == 0 {
		// TODO: Why zero? Log this to /var/log/span.log or something.
		return result, nil
	}
	// At a minimum, the sid and tcid or collection name must match.
	q, args, err := sqlx.In(`
		SELECT isil, sid, tcid, mc, hflink, hfeval, cflink, cfelink FROM amsl WHERE sid = ? AND (mc IN (?) OR tcid IN (?))
	`, doc.SourceID, doc.MegaCollections, doc.MegaCollections)
	if err != nil {
		return nil, err
	}
	rows, err := l.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cr ConfigRow
		err = rows.Scan(&cr.ISIL,
			&cr.SourceID,
			&cr.TechnicalCollectionID,
			&cr.MegaCollection,
			&cr.LinkToHoldingsFile,
			&cr.EvaluateHoldingsFileForLibrary,
			&cr.ContentFileURI,
			&cr.ExternalLinkToContentFile)
		if err != nil {
			return nil, err
		}
		result = append(result, cr)
	}
	l.cache[key] = result
	return result, nil
}

// Label updates document in place. This may contain hard-coded values for
// special attachment cases.
func (l *Labeler) Labels(doc *finc.IntermediateSchema) ([]string, error) {
	rows, err := l.matchingRows(doc)
	if err != nil {
		return nil, err
	}
	var labels = make(map[string]struct{}) // ISIL to attach

	// TODO: Distinguish and simplify cases, e.g. with or w/o HF,
	// https://git.io/JvdmC, also log some stats.
	// INFO[12576] lthf => 531,701,113
	// INFO[12576] plain => 112,694,196
	// INFO[12576] 34-music => 3692
	// INFO[12576] 34-DE-15-FID-film => 770
	for _, row := range rows {
		// Fields, where KBART links might be, empty strings are just skipped.
		kbarts := []string{row.LinkToHoldingsFile, row.LinkToContentFile, row.ExternalLinkToContentFile}
		// DE-14 uses a KBART (probably) across all sources, so we hard code
		// their link here. Use `-f` to force download all external files.
		if row.ISIL == "DE-14" {
			kbarts = append(kbarts, SLUBEZBKBART)
		}
		switch {
		case row.ISIL == "DE-15" && row.TechnicalCollectionID == "sid-48-col-wisoubl":
			for _, name := range doc.Packages {
				if _, ok := UBLWISOPROFILE[name]; ok {
					labels[row.ISIL] = struct{}{}
				}
			}
		case doc.SourceID == "34":
			switch {
			case stringsContain([]string{"DE-L152", "DE-1156", "DE-1972", "DE-Kn38"}, row.ISIL):
				// refs #10495, a subject filter for a few hard-coded ISIL; https://git.io/JvFjE
				if stringsOverlap(doc.Subjects, []string{"Music", "Music education"}) {
					labels[row.ISIL] = struct{}{}
				}
			case row.ISIL == "DE-15-FID":
				// refs #10495, maybe use a TSV with custom column name to use a subject list? https://git.io/JvFjd
				if stringsOverlap(doc.Subjects, []string{"Film studies", "Information science", "Mass communication"}) {
					labels[row.ISIL] = struct{}{}
				}
			}
		case row.ISIL == "DE-15-FID":
			if strings.Contains(row.LinkToHoldingsFile, "FID_ISSN_Filter") {
				// Here, the holdingfile URL contains a list of ISSN.  URI like ...
				// discovery/metadata-usage/Dokument/FID_ISSN_Filter - but that
				// might change. Assuming just a single file.
				if _, ok := l.whitelistCache[DE15FIDISSNWHITELIST]; !ok {
					// Load from file, once. One value per line.
					resp, err := pester.Get(row.LinkToHoldingsFile)
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()
					if l.whitelistCache == nil {
						l.whitelistCache = make(map[string]map[string]struct{})
					}
					l.whitelistCache[DE15FIDISSNWHITELIST] = make(map[string]struct{})
					if err := setFromLines(resp.Body, l.whitelistCache[DE15FIDISSNWHITELIST]); err != nil {
						return nil, err
					}
					log.Printf("loaded whitelist of %d items from %s", len(l.whitelistCache[DE15FIDISSNWHITELIST]), row.LinkToHoldingsFile)
				}
				whitelist, ok := l.whitelistCache[DE15FIDISSNWHITELIST]
				if !ok {
					return nil, fmt.Errorf("whitelist cache broken")
				}
				for _, issn := range doc.ISSNList() {
					if _, ok := whitelist[issn]; ok {
						labels[row.ISIL] = struct{}{}
					}
				}
			}
		case row.EvaluateHoldingsFileForLibrary == "yes" && row.LinkToHoldingsFile != "" && row.LinkToContentFile != "":
			ok, err := l.hfcache.Covered(doc, And, kbarts...)
			if err != nil {
				return nil, err
			}
			if ok {
				labels[row.ISIL] = struct{}{}
			}
		case row.EvaluateHoldingsFileForLibrary == "yes" && row.LinkToHoldingsFile != "" && row.ExternalLinkToContentFile != "":
			ok, err := l.hfcache.Covered(doc, And, kbarts...)
			if err != nil {
				return nil, err
			}
			if ok {
				labels[row.ISIL] = struct{}{}
			}
		case row.EvaluateHoldingsFileForLibrary == "yes" && row.LinkToHoldingsFile != "" && row.LinkToContentFile == "" && row.ExternalLinkToContentFile == "":
			ok, err := l.hfcache.Covered(doc, Or, kbarts...)
			if err != nil {
				return nil, err
			}
			if ok {
				labels[row.ISIL] = struct{}{}
			}
		case row.EvaluateHoldingsFileForLibrary == "yes" && row.LinkToHoldingsFile == "":
			return nil, fmt.Errorf("no holding file to evaluate: %v", row)
		case row.EvaluateHoldingsFileForLibrary == "no" && row.LinkToHoldingsFile != "":
			return nil, fmt.Errorf("config provides holding file, but does not want to evaluate it: %v", row)
		case row.ExternalLinkToContentFile != "":
			// https://git.io/JvFjx
			ok, err := l.hfcache.Covered(doc, And, kbarts...)
			if err != nil {
				return nil, err
			}
			if ok {
				labels[row.ISIL] = struct{}{}
			}
		case row.LinkToContentFile != "":
			// https://git.io/JvFjp
			ok, err := l.hfcache.Covered(doc, And, kbarts...)
			if err != nil {
				return nil, err
			}
			if ok {
				labels[row.ISIL] = struct{}{}
			}
		case row.EvaluateHoldingsFileForLibrary == "no":
			labels[row.ISIL] = struct{}{}
		case row.ContentFileURI != "":
		default:
			return nil, fmt.Errorf("none of the attachment modes match for %v", doc)
		}
	}
	var keys []string
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// stringsSliceContains returns true, if value appears in a string slice.
func stringsContain(ss []string, v string) bool {
	for _, w := range ss {
		if v == w {
			return true
		}
	}
	return false
}

// stringsOverlap returns true, if at least one value is in both ss and vv.
// Inefficient.
func stringsOverlap(ss, vv []string) bool {
	for _, s := range ss {
		for _, v := range vv {
			if s == v {
				return true
			}
		}
	}
	return false
}

// setFromLines populates a set from lines in a reader.
func setFromLines(r io.Reader, m map[string]struct{}) error {
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		m[line] = struct{}{}
	}
	return nil
}
