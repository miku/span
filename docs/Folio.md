# FOLIO

## Questions

* CQL match all
* account locked

## Notes

* multi-tenant
* finc-config, finc-select
* Q4/21 testing

## 404 vs 502

First yields 404, second 502.

```
$ curl -v 'https://zzzz.folio.finc.info/finc-config/metadata-collections?query=(selectedBy=("DE-15"))'
```

```
$ curl -v 'https://zzzz.folio.finc.info/finc-config/metadata-collections?query=(selectedBy=("DIKU-01" or "DE-15"))'
```

Project name: ERM migration.

## Example response

```
{
  "metadataCollections": [
    {
      "id": "6dd325f8-b1d5-4568-a0d7-aecf6b8d6123",
      "label": "21st Century COE Program",
      "description": "This is a test metadata collection",
      "mdSource": {
        "id": "6dd325f8-b1d5-4568-a0d7-aecf6b8d6697",
        "name": "Cambridge University Press Journals"
      },
      "metadataAvailable": "yes",
      "usageRestricted": "no",
      "permittedFor": [
        "DE-15",
        "DE-14"
      ],
      "freeContent": "undetermined",
      "lod": {
        "publication": "permitted (explicit)",
        "note": "Note for test publication"
      },
      "collectionId": "coe-123",
      "facetLabel": "012.1 21st Century COE",
      "solrMegaCollections": [
        "21st Century COE Program"
      ]
    },
    {
      "id": "9a2427cd-4110-4bd9-b6f9-e3475631bbac",
      "label": "21st Century Political Science Association",
      "description": "This is a test metadata collection 2",
      "mdSource": {
        "id": "f6f03fb4-3368-4bc0-bc02-3bf6e19604a5",
        "name": "Early Music Online"
      },
      "metadataAvailable": "yes",
      "usageRestricted": "no",
      "permittedFor": [
        "DE-14"
      ],
      "freeContent": "no",
      "lod": {
        "publication": "permitted (explicit)",
        "note": "Note for test publication"
      },
      "collectionId": "psa-459",
      "facetLabel": "093.8 21st Century Political Science",
      "solrMegaCollections": [
        "21st Century Political Science"
      ]
    }
  ],
  "totalRecords": 2
}
```

Two records.

> HTTP header, auth, login.

* X-OKAPI-HEADER

# Auth

```shell
$ curl --dump-header okapi.txt --request POST \
    --url https://xyz.folio.finc.info/bl-users/login \
    --header 'content-type: application/json' \
    --header 'x-okapi-tenant: de_15' --data '{"sername": "xyz", "password": "xyz"}'
```
