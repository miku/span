#!/usr/bin/env python

"""
Create a mapping from any name for a publisher to one or more canonical names.
Currently, about 67 names have more than one primary name.

Input is the names.ndj file generated via Makefile, e.g.

    # Primary and other names.
    assets/crossref/names.ndj:
            span-crossref-members | jq -rc '.message.items[]| {"primary": .["primary-name"], "names": .["names"]}' > assets/crossref/names.ndj

Output is a single JSON object. This mapping can then be used to unify
publisher names for crossref.

  ...
  "Kozminski University": [
    "Kozminski University"
  ],
  "Faculty of Medicine, Chulalongkorn University (Asian Biomedicine)": [
    "Faculty of Medicine, Chulalongkorn University (Asian Biomedicine)"
  ],
  "Springer - Real Academia de Ciencias Exactas, Fisicas y Naturales": [
    "Springer Nature"
  ],
  ...

"""

import collections
import json
import os

name_to_primary = collections.defaultdict(set)
names_file = os.path.join(os.path.dirname(__file__), '../assets/crossref/names.ndj')

class SetEncoder(json.JSONEncoder):
    """
    Helper to encode python sets into JSON lists.  So you can write something
    like this:

        json.dumps({"things": set([1, 2, 3], cls=SetEncoder)}
    """

    def default(self, obj):
        """
        Decorate call to standard implementation.
        """
        if isinstance(obj, set):
            return list(obj)
        return json.JSONEncoder.default(self, obj)


if __name__ == '__main__':
    with open(names_file) as handle:
        for line in handle:
            doc = json.loads(line)
            for name in doc["names"]:
                if not name.strip():
                    continue
                name_to_primary[name].add(doc["primary"])

    print(json.dumps(name_to_primary, cls=SetEncoder))

