// span-folio talks to FOLIO API to fetch ISIL, collections and other
// information relevant to attachments.
// Docs: https://s3.amazonaws.com/foliodocs/api/mod-finc-config/p/fincConfigMetadataCollections.html
//
// Get metadata collections per ISIL, each "fincConfigMetadataCollections", "FilterToCollections", "Filter".
//
// Tenant specific filter. The "finc-select" is special views. Whitelist,
// blacklist filter. EZB holdings is a whitelist. Blacklist predatory journals.
// Filter. How does a filter rule.
package main

import "flag"

var (
	muFolio = flag.String("folio", "", "folio endpoint")
)

func main() {
	flag.Parse()
}
