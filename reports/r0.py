#!/usr/bin/env python
# coding: utf-8

"""
WIP: Create an Excel report from a span-0.1.253-ish span-report output.

    $ span-report -bs 100 -r faster -server 10.1.1.1:8085/solr/biblio > data.json
    $ python reports/r0.py -x data.json

The span-report output (data.json) for AI in June 2018 contains 236149 entries,
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

Now we try to fit 40 dates per issn (on average, 10M data points in total) into
a single Excel file, so that it's readable. The raw data is about 150MB.

A first version (https://git.io/fNSTT) was slow, 90min for an Excel file with
315 sheets, years 2018 to 1700.
"""

import argparse
import collections
import fileinput
import itertools
import json
import logging
import os
import pickle
import sys

import numpy as np
import pandas as pd
import tqdm

logger = logging.getLogger('r0')
logger.setLevel(logging.DEBUG)
ch = logging.StreamHandler()
formatter = logging.Formatter(
    '[%(asctime)s][%(name)s][%(levelname)s] %(message)s')
ch.setFormatter(formatter)
logger.addHandler(ch)


def contains_year_entries(doc, year=2018):
    """
    Returns true, if an entry contains publications in a given year.
    """
    for date in doc['dates']:
        if date.startswith(year):
            return True
    return False


def publications_per_year(entries):
    """
    Group by publication, size per year.
    """
    yearsize = collections.defaultdict(int)

    for _, doc in tqdm.tqdm(entries.items(), total=len(entries)):
        for year in [str(y) for y in range(2018, 1900, -1)]:
            if contains_year_entries(doc, year):
                yearsize[year] += 1

    return yearsize


if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter)

    parser.add_argument('--publications-per-year', '-y', action='store_true',
                        help='TSV with publications per year.')
    parser.add_argument('--date-list', '-d', action='store_true',
                        help='List all occuring dates.')
    parser.add_argument('--output', '-o', default='df.xlsx',
                        help='Excel output filename.')
    parser.add_argument('--excel', '-x', action='store_true',
                        help='Create Excel file manually.')
    parser.add_argument('--force', '-f', action='store_true',
                        help='Do not use cached version.')
    parser.add_argument('--max-sheets', '-m', default=10000, type=int,
                        help='Maximum number of sheets to write.')
    parser.add_argument('file', metavar='file', type=str, nargs=1,
                        help='Raw report file.')
    args = parser.parse_args()

    if args.force and os.path.exists('r0.pkl'):
        os.remove('r0.pkl')

    entries = {}

    if not os.path.exists('r0.pkl'):
        with open(args.file) as handle:
            for line in handle:
                doc = json.loads(line)
                key = (doc['issn'], doc['c'])
                if key in entries:
                    raise ValueError('duplicate issn and collection: %s' % key)
                entries[key] = doc

        with open('r0.pkl.tmp', 'wb') as output:
            pickle.dump(entries, output)
        os.rename('r0.pkl.tmp', 'r0.pkl')

    with open('r0.pkl', 'rb') as handle:
        entries = pickle.load(handle)

    if args.publications_per_year:
        for year, size in sorted(publications_per_year(entries).items()):
            print('%s\t%s' % (year, size))
        sys.exit(0)

    if args.date_list:
        dates = set()

        for _, doc in entries.items():
            for date in doc['dates']:
                dates.add(date)

        for date in sorted(dates):
            print(date)
        sys.exit(0)

    if args.excel:
        # First version, per sheet: collections x month.
        dates = set(itertools.chain(*[doc['dates'].keys()
                                      for _, doc in entries.items()]))
        logger.debug("distinct dates %s", len(dates))

        years = sorted([year for year in set(
            [date[:4] for date in dates]) if '1700' < year < '2019'], reverse=True)
        logger.debug("distinct years %s, %s ...", len(years), years[:10])

        writer = pd.ExcelWriter(args.output, engine='xlsxwriter')

        for i, year in enumerate(tqdm.tqdm(years, total=min(len(years), args.max_sheets)), start=1):
            data = collections.defaultdict(dict)

            for month in ('%02d' % s for s in range(1, 13)):
                prefix = '%s-%s' % (year, month)
                for _, doc in entries.items():
                    s = sum(count for date,
                            count in doc['dates'].items() if date[:7] == prefix)
                    if s > 0:
                        data[month][doc['c']] = s

            pd.DataFrame(data).to_excel(writer, sheet_name=year)
            if i == args.max_sheets:
                break

        writer.save()
