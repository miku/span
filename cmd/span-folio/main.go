// WIP: span-folio talks to FOLIO API to fetch ISIL, collections and other
// information relevant to attachments.  Docs:
// https://s3.amazonaws.com/foliodocs/api/mod-finc-config/p/fincConfigMetadataCollections.html
//
// Get metadata collections per ISIL, each "fincConfigMetadataCollections",
// "FilterToCollections", "Filter".
//
// Tenant specific filter. Whitelist, blacklist filter. EZB holdings is a
// whitelist. Blacklist predatory journals.
//
// Every filter on each collection. Workflow field (testing, approved).
//
// Detour: Regular expressions in RM.
//
// Previously: Technical collection identifier to ISIL (tcid => ISIL).
//
// 1       ShardLabel
// 2       ISIL
// 3       SourceID
// 4       TechnicalCollectionID
// 5       MegaCollection
// 6       HoldingsFileURI
// 7       HoldingsFileLabel
// 8       LinkToHoldingsFile
// 9       EvaluateHoldingsFileForLibrary
// 10      ContentFileURI
// 11      ContentFileLabel
// 12      LinkToContentFile
// 13      ExternalLinkToContentFile
// 14      ProductISIL
// 15      DokumentURI
// 16      DokumentLabel
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span/folio"
	spanfolio "github.com/miku/span/internal/cmd/folio"
	"github.com/miku/span/xflag"
	"github.com/sethgrid/pester"
)

// TODO: Add config file location, also: unify config file handling.
// https://okapi.testing.dev.folio.finc.info

var (
	muFolio   = flag.String("folio", "https://okapi.erm.staging.folio.finc.info", "folio endpoint")
	tenant    = flag.String("tenant", "de_15", "folio tenant")
	limit     = flag.Int("limit", 100000, "limit for lists")
	cqlQuery  = flag.String("cql", `(selectedBy=("*"))`, `cql query, e.g. (selectedBy=("DE-15")`)
	rawOutput = flag.Bool("r", false, "raw output")
	userPass  xflag.UserPassword
)

func main() {
	flag.Var(&userPass, "u", "user:password for api")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-folio [options]\n\n")
		fmt.Fprintf(os.Stderr, "Talks to the FOLIO API to fetch ISIL, metadata collections and related\n")
		fmt.Fprintf(os.Stderr, "attachment information for a tenant. Work in progress.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	api := &folio.API{
		Base:   *muFolio,
		Tenant: *tenant,
		Client: pester.New(),
	}
	cfg := spanfolio.Config{
		User:     userPass.User,
		Password: userPass.Password,
		CQL:      *cqlQuery,
		Limit:    *limit,
		Raw:      *rawOutput,
	}
	if err := spanfolio.Run(cfg, api, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
