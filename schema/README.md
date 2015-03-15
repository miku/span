Intermediate Schema Specification
=================================

The intermediate schema serves, among others, these purposes:

1. It should assists clients in generating OpenURLs.

2. It should assists clients in generating citation formats.

3. It should provide an intermediary between data sources and export formats.
   Instead of implementing m x n transformations for m data sources and n export formats, it reduces the effort to m + n.

4. It can serve as a catch-all format, leveling out peculiarities of input data formats.

The default serialization is JSON. Documents can be validated against a JSON schema (draft 4). The schema is versioned.

Minor updates shall not break clients. Major updates may break clients.

Notes
-----

* Languages can refer to the language of the article or the abstract.
  Articles and abstracts both might use more than one language. The [ISO639-3](http://www.sil.org/iso639-3/) language codes are used.

----

The `is-<version>.json` file contains the JSON schema.
Versioned examples can be found under the `fixtures` directory.

To run validation against a schema, use one of the many validators available. Here's one [in python](https://pypi.python.org/pypi/jsonschema):

    $ jsonschema -i fixtures/0.9/jats.is is-0.9.json
