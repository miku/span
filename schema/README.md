Schema
======

The `is-<version>.json` file contains the JSON schema.
Versioned examples can be found under the `fixtures` directory.

To run validation against a schema, use one of the many validators available. Here's one [in python](https://pypi.python.org/pypi/jsonschema):

    $ jsonschema -i fixtures/0.9/jats.is is-0.9.json
