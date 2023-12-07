#!/usr/bin/env python

import json

def load_mappings():
    with open("subjects.json") as f:
        old = json.loads(f.read())
        # "crossref" -> ["fincclass[de]"]
    with open("classification.json") as f:
        new = json.loads(f.read())
        # "fincclass[en]" -> {"de": ...}
    return old, new

def make_map_fincclass_de_en(mapping):
    result = {}
    for k, blob in mapping.items():
        result[blob["de"]] = k
    return result


def main():
    old, new = load_mappings()
    map_de_en = make_map_fincclass_de_en(new)

    updated = {}
    for k, vs in old.items():
        updated_vs = [map_de_en[v] for v in vs if v in map_de_en]
        updated[k] = updated_vs

    print(json.dumps(updated))


if __name__ == '__main__':
    main()
