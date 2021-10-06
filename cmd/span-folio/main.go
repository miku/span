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
//
// Notes:
// {
//   "errors": [
//     {
//       "message": "Error verifying user existence: Error looking up user at url http://okapi-app-service-erm-staging:9130/users?query=username==user Expected status code 200, got 400 :function count_estimate(unknown) does not exist",
//       "type": "error",
//       "code": "username.incorrect",
//       "parameters": [
//         {
//           "key": "username",
//           "value": "user"
//         }
//       ]
//     }
//   ]
// }
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/miku/span/folio"
	"github.com/miku/span/strutil"
	"github.com/miku/span/xflag"
	"github.com/sethgrid/pester"
)

// TODO: Add config file location, also: unify config file handling.
// https://okapi.testing.dev.folio.finc.info

var (
	muFolio  = flag.String("folio", "https://okapi.erm.staging.folio.finc.info", "folio endpoint")
	tenant   = flag.String("tenant", "de_15", "folio tenant")
	limit    = flag.Int("limit", 100000, "limit for lists")
	cqlQuery = flag.String("cql", `(selectedBy=("DE-15"))`, "cql query")
	userPass xflag.UserPassword
)

func main() {
	flag.Var(&userPass, "u", "user:password for api")
	flag.Parse()
	api := folio.API{
		Base:   *muFolio,
		Tenant: *tenant,
		Client: pester.New(),
	}
	if userPass.User == "" || userPass.Password == "" {
		log.Fatal("incomplete credentials")
	}
	if err := api.Authenticate(userPass.User, userPass.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("[ok] auth")
	opts := folio.MetadataCollectionsOpts{
		CQL:   *cqlQuery,
		Limit: *limit,
	}
	resp, err := api.MetadataCollections(opts)
	if err != nil {
		log.Fatal(err)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	defer w.Flush()
	for _, entry := range resp.FincConfigMetadataCollections {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			strutil.Truncate(entry.Label, 40),
			strutil.Truncate(strings.Join(entry.SolrMegaCollections, ", "), 40),
			strutil.Truncate(entry.MdSource.Name, 40))
	}
}
