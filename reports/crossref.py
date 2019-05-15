#!/usr/bin/env python
# coding: utf-8

"""
WIP.

There are a few stats on collection name and publisher name usage in Crossref.

Currently, only the publisher field is used, which seems to contain a lot of
garbage. We can use the members API to fetch DOI prefixes, a canonical name and
name variants.

Inputs:

* a list of publisher names in Crossref as of May 2019 (jq, CrossrefUniqItems)
* a list of DOI in Crossref (jq, CrossreqUniqItems)
* a list of DOI prefixes in the members API responses, canonical names and variants
* a list of collection names in AMSL
"""
