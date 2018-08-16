#!/usr/bin/env python
# coding: utf-8

"""
WIP: Create an Excel report from a span-0.1.253-ish span-report output.

    $ span-report -bs 100 -r faster -server 10.1.1.1:8085/solr/biblio > data.ndj
    $ python reports/r1.py -x data.ndj

The span-report output (data.ndj) for AI in June 2018 contains 236149 entries,
where each line contains one issn:

    {
      "c": "Scientific Online Publishing, Co. Ltd. (CrossRef)",
      "dates": {
        "2014-05-30": 6,
        "2015-02-28": 3
      },
      "issn": "2374-4944",
      "sid": "49",
      "size": 9
    }

Output: One sheet per collection, 10 years back (2018-2008) per ISSN in a
single sheet (120 columns). Collection > ISSN > Dates.
"""

import argparse
import collections
import fileinput
import itertools
import json
import logging
import os
import pickle
import random
import string
import sys

import numpy as np
import pandas as pd
import tqdm

logger = logging.getLogger('r1')
logger.setLevel(logging.DEBUG)
ch = logging.StreamHandler()
formatter = logging.Formatter(
    '[%(asctime)s][%(name)s][%(levelname)s] %(message)s')
ch.setFormatter(formatter)
logger.addHandler(ch)

def slugify(s):
    valid_chars = "-_ %s%s" % (string.ascii_letters, string.digits)
    return ''.join(c for c in s if c in valid_chars)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter)

    parser.add_argument('--output', '-o', default='r1.xlsx',
                        help='Excel output filename.')
    parser.add_argument('--excel', '-x', action='store_true',
                        help='Create Excel file manually.')
    parser.add_argument('--force', '-f', action='store_true',
                        help='Do not use cached version.')
    parser.add_argument('--max-sheets', '-m', default=10000, type=int,
                        help='Maximum number of sheets to write.')
    parser.add_argument('file', metavar='file', type=str,
                        help='Raw report file.')
    args = parser.parse_args()

    if args.force and os.path.exists('r1.pkl'):
        os.remove('r1.pkl')

    entries = collections.defaultdict(dict)

    if not os.path.exists('r1.pkl'):
        with open(args.file) as handle:
            for line in handle:
                doc = json.loads(line)
                entries[doc['c']][doc['issn']] = doc

        with open('r1.pkl.tmp', 'wb') as output:
            pickle.dump(entries, output)
        os.rename('r1.pkl.tmp', 'r1.pkl')

    with open('r1.pkl', 'rb') as handle:
        entries = pickle.load(handle)

    if args.excel:
        writer = pd.ExcelWriter(args.output, engine='xlsxwriter')

        names = set()

        for i, (c, blob) in enumerate(entries.items()):
            logger.debug('%s (%s)', c, len(blob))
            drop = True # Drop sheet, if all entries are zero.

            # blob is a dictionary, issn key, the document as value
            # data will be fed into DataFrame
            data = collections.defaultdict(dict)

            for issn, doc in blob.items():
                for year in range(2018, 2007, -1):
                    for month in ('%02d' % s for s in range(12, 0, -1)):
                        prefix = '%d-%s' % (year, month)
                        s = sum(count for date, count in doc['dates'].items()
                                if date[:7] == prefix)

                        data[prefix][issn] = s if s else ''

                        if s > 0:
                            drop = False
            if drop:
                continue

            # Sheetname 'Bioscientifica CrossRef', with case ignored, is already in use.
            # Invalid Excel character '[]:*?/\' in sheetname 'ID Design 2012/DOOEL Skopje (Cr'
            # Excel worksheet name 'Finnish Zoological and Botanical Publishing Board (CrossRef)' must be <= 31 chars
            sheet_name = slugify(c[:31])

            if sheet_name in names or sheet_name.lower() in names:
                sheet_name = '%s-%s' % (sheet_name[:26], random.randint(0, 9999))
                logger.debug("duplicate slugified collection name: %s => %s" % (c, sheet_name))

            names.add(sheet_name)
            names.add(sheet_name.lower())

            pd.DataFrame(data).to_excel(writer, sheet_name=sheet_name)
            if i == args.max_sheets:
                break

        writer.save()
