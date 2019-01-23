// Package amslutil provides helpers for accessing AMSL API.
//
// For visibility decision we need:
//
// * ISIL, SID, collection (name or id)
// * ISIL and list of serial numbers
//
// AMSL is in transition from v1 (OW) to v2 (FO). Below v1 only.
//
// MU -> tcid
// MU -> isil
// HF[isil] -> LinkToFile (KBART), or DokumentURI
//
//
// MU
//
// {
//   "sourceID": "14",
//   "megaCollection": "RISM",
//   "productISIL": null,
//   "technicalCollectionID": "sid-14-col-rism",
//   "ISIL": "DE-D117",
//   "shardLabel": "UBL-main"
// }
//
// MUC
//
// {
//   "sourceID": "26",
//   "megaCollection": "DOAB Directory of Open Access Books",
//   "productISIL": "ZDB-119-KEO",
//   "technicalCollectionID": "sid-26-col-doab",
//   "ISIL": "DE-14;DE-15;DE-15-FID;DE-540;DE-Ch1;DE-D117;DE-L242",
//   "shardLabel": "UBL-main"
// }
//
// HF
//
// {
//   "ISIL": "DE-15-FID",
//   "DokumentLabel": "FID_ISSN_Filter",
//   "DokumentURI": "http://amsl.technology/discovery/metadata-usage/Dokument/FID_ISSN_Filter",
//   "LinkToFile": "https://live.amsl.technology/OntoWiki/files/get?setResource=http://amsl.technology/discovery/metadata-usage/Dokument/FID_ISSN_Filter"
// }
//
// HFC
//
// {
//   "sourceID": "17",
//   "megaCollection": "Oxford Journals Digital Archive 1849-2010",
//   "productISIL": "ZDB-1-OJD",
//   "technicalCollectionID": null,
//   "ISIL": "DE-105",
//   "shardLabel": "UBL-main"
// }
//
// CF
//
// {
//   "technicalCollectionID": "zdb-39-joa",
//   "megaCollection": "JSTOR Film and Performing Arts",
//   "contentFileLabel": "JSTOR Film and Performing Arts KBART",
//   "contentFileURI": "http://amsl.technology/discovery/Dokument/JSTOR_Film_and_Performing_Arts_KBART",
//   "linkToContentFile": null
// }
//
package amslutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// MetadataUsage is a single entry from the MetadataUsage endpoint.
type MetadataUsage struct {
	ISIL                  string
	MegaCollection        string `json:"megaCollection"`
	ProductISIL           string `json:"productISIL"`
	ShardLabel            string `json:"shardLabel"`
	SourceID              string `json:"sourceID"`
	TechnicalCollectionID string `json:"technicalCollectionID"`
}

// HoldingsFile is a single entry in the holdingsfiles endpoint.
type HoldingsFile struct {
	DokumentLabel string
	DokumentURI   string
	ISIL          string
	LinkToFile    string
}

// HoldingsFileConcat is a single entry in the holdings_file_concat (testing)
// endpoint.
type HoldingsFileConcat struct {
	ISIL                  string
	MegaCollection        string `json:"megaCollection"`
	ProductISIL           string `json:"productISIL"`
	ShardLabel            string `json:"shardLabel"`
	SourceID              string `json:"sourceID"`
	TechnicalCollectionID string `json:"technicalCollectionID"`
}

// ContentFile is a single entry in the contentfile endpoint.
type ContentFile struct {
	ContentFileLabel      string `json:"contentFileLabel"`
	ContentFileURI        string `json:"contentFileURI"`
	LinkToContentFile     string `json:"linkToContentFile"`
	MegaCollection        string `json:"megaCollection"`
	TechnicalCollectionID string `json:"technicalCollectionID"`
}

// Amsl endpoint.
type Amsl struct {
	Server string
}

// ResourceDownloadLink turns a resource URI into a download link. The URI is
// not escaped.
func (amsl *Amsl) ResourceDownloadLink(uri string) string {
	return fmt.Sprintf("%s/OntoWiki/files/get?setResource=%s",
		strings.TrimRight(amsl.Server, "/"), uri)
}

// MetadataUsage endpoint.
func (amsl *Amsl) MetadataUsage() ([]MetadataUsage, error) {
	link := fmt.Sprintf("%s/outboundservices/list?do=metadata_usage",
		strings.TrimRight(amsl.Server, "/"))
	log.Println(link)
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var payload []MetadataUsage
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// MetadataUsageConcat endpoint.
func (amsl *Amsl) MetadataUsageConcat() ([]MetadataUsage, error) {
	link := fmt.Sprintf("%s/outboundservices/list?do=metadata_usage_concat",
		strings.TrimRight(amsl.Server, "/"))
	log.Println(link)
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var payload []MetadataUsage
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// HoldingsFile endpoint.
func (amsl *Amsl) HoldingsFiles() ([]HoldingsFile, error) {
	link := fmt.Sprintf("%s/outboundservices/list?do=holdingsfiles",
		strings.TrimRight(amsl.Server, "/"))
	log.Println(link)
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var payload []HoldingsFile
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// HoldingsFileConcat endpoint.
func (amsl *Amsl) HoldingsFileConcat() ([]HoldingsFileConcat, error) {
	link := fmt.Sprintf("%s/outboundservices/list?do=holdings_file_concat",
		strings.TrimRight(amsl.Server, "/"))
	log.Println(link)
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var payload []HoldingsFileConcat
	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)
	dec := json.NewDecoder(tee)
	if err := dec.Decode(&payload); err != nil {
		log.Printf(buf.String())
		return nil, err
	}
	return payload, nil
}

// ContentFile endpoint.
func (amsl *Amsl) ContentFiles() ([]ContentFile, error) {
	link := fmt.Sprintf("%s/outboundservices/list?do=contentfiles",
		strings.TrimRight(amsl.Server, "/"))
	log.Println(link)
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var payload []ContentFile
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}
