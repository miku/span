# Notes on fast and transparent index updates

> 2019-03-29

Index dims: 100-250M docs, 300-700G, single node or distributed.

Required, currently done offline:

* data updates
* licensing (content + licensing data)
* deduplication (doi only)

Relevant events:

* [ ] amsl, folio entry has changed (single, batch)
* [ ] associated kbart file has changed
* [ ] one or more documents can be updated (auto-update, manual update)
* [ ] one or more documents need to be deleted
* [ ] index schema changes

Concerns:

* transparency (feed of actions)
* how to rollback an action on a live index
* performance hits
* automation, layer tradeoffs
* backup and rebuild

Developer dreams:

* create throwaway indices (with new data) plus some web for end user testing

Options:

* ignore duplicates, handle them in solr at search time
* licensing and deduplication proxy (extra component)
* implement custom request handler in solr (tied to solr)
* ...
