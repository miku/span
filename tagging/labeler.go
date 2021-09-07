// Package tagging is a rewrite of span-tag for applying licensing information
// of intermediate schema data. While span-tag uses a declarative approach (a
// JSON configuration), this package tries to express things in code; maybe
// uglier, less declarative, but more flexible, in the best case.
package tagging

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrg/xdg"
	"github.com/jmoiron/sqlx"
	"github.com/miku/span/container"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/licensing"
	"github.com/miku/span/misc"
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

	cache          map[string][]ConfigRow          // cache for prefiltering rows from database
	hfcache        *HFCache                        // holding file cache
	whitelistCache map[string]*container.StringSet // Name (e.g. DE15FIDISSNWHITELIST) -> Set (e.g. a set of ISSN)

	// Returns a unique key for a given document usable as cache key.
	cacheKeyFunc func(doc *finc.IntermediateSchema) string
}

// New returns a initialized labeler using a relational AMSL representation in
// an sqlite3 file. Will fail here, if database cannot be opened.
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
	var (
		key   = l.cacheKeyFunc(doc)
		v, ok = l.cache[key]
	)
	if ok {
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

// Labels returns a list of ISIL that are interested in this document.
func (l *Labeler) Labels(doc *finc.IntermediateSchema) ([]string, error) {
	var (
		labels    = container.NewStringSet()
		rows, err = l.matchingRows(doc)
	)
	if err != nil {
		return nil, err
	}
	var (
		subjectsMusic = []string{"Music", "Music education"}
		subjectsFilm  = []string{"Film studies", "Information science", "Mass communication"}
		isilSet10495  = container.NewStringSet("DE-L152", "DE-1156", "DE-1972", "DE-Kn38")
	)
	// TODO: Distinguish and simplify cases, e.g. with or w/o HF,
	// https://git.io/JvdmC, also log some stats.
	// INFO[12576] lthf => 531,701,113
	// INFO[12576] plain => 112,694,196
	// INFO[12576] 34-music => 3692
	for _, row := range rows {
		// Fields, where KBART links might be, empty strings are just skipped.
		kbarts := []string{
			row.LinkToHoldingsFile,
			row.LinkToContentFile,
			row.ExternalLinkToContentFile,
		}
		// DE-14 uses a KBART (probably) across all sources, so we hard code
		// their link here. Use `-f` to force download all external files.
		if row.ISIL == "DE-14" {
			kbarts = append(kbarts, SLUBEZBKBART)
		}
		switch {
		case row.ISIL == "DE-15" && row.TechnicalCollectionID == "sid-48-col-wisoubl":
			for _, name := range doc.Packages {
				if _, ok := UBLWISOPROFILE[name]; ok {
					labels.Add(row.ISIL)
				}
			}
		case doc.SourceID == "34":
			switch {
			case isilSet10495.Contains(row.ISIL):
				// refs #10495, a subject filter for a few hard-coded ISIL;
				// https://git.io/JvFjE
				if misc.Overlap(doc.Subjects, subjectsMusic) {
					labels.Add(row.ISIL)
				}
			case row.ISIL == "FID-MEDIEN-DE-15":
				// refs #10495, maybe use a TSV with custom column name to use
				// a subject list? https://git.io/JvFjd
				if misc.Overlap(doc.Subjects, subjectsFilm) {
					labels.Add(row.ISIL)
				}
			}
		case row.ISIL == "FID-MEDIEN-DE-15":
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
						ss, err := container.NewStringSetReader(resp.Body)
						if err != nil {
							return nil, err
						}
						l.whitelistCache[DE15FIDISSNWHITELIST] = ss
					}
					log.Printf("loaded whitelist of %d items from %s",
						l.whitelistCache[DE15FIDISSNWHITELIST].Size(),
						row.LinkToHoldingsFile)
				}
				for _, issn := range doc.ISSNList() {
					if l.whitelistCache[DE15FIDISSNWHITELIST].Contains(issn) {
						labels.Add(row.ISIL)
					}
				}
			}
		case row.EvaluateHoldingsFileForLibrary == "yes" && row.LinkToHoldingsFile != "" && row.LinkToContentFile != "":
			ok, err := l.hfcache.Covered(doc, And, kbarts...)
			if err != nil {
				return nil, err
			}
			if ok {
				labels.Add(row.ISIL)
			}
		case row.EvaluateHoldingsFileForLibrary == "yes" && row.LinkToHoldingsFile != "" && row.ExternalLinkToContentFile != "":
			ok, err := l.hfcache.Covered(doc, And, kbarts...)
			if err != nil {
				return nil, err
			}
			if ok {
				labels.Add(row.ISIL)
			}
		case row.EvaluateHoldingsFileForLibrary == "yes" && row.LinkToHoldingsFile != "" && row.LinkToContentFile == "" && row.ExternalLinkToContentFile == "":
			ok, err := l.hfcache.Covered(doc, Or, kbarts...)
			if err != nil {
				return nil, err
			}
			if ok {
				labels.Add(row.ISIL)
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
				labels.Add(row.ISIL)
			}
		case row.LinkToContentFile != "":
			// https://git.io/JvFjp
			ok, err := l.hfcache.Covered(doc, And, kbarts...)
			if err != nil {
				return nil, err
			}
			if ok {
				labels.Add(row.ISIL)
			}
		case row.EvaluateHoldingsFileForLibrary == "no":
			labels.Add(row.ISIL)
		case row.ContentFileURI != "":
		default:
			return nil, fmt.Errorf("none of the attachment modes match for %v", doc)
		}
	}
	return labels.SortedValues(), nil
}
