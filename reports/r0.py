#!/usr/bin/env python
# coding: utf-8

"""
WIP: Create an Excel report from a span-0.1.253-ish span-report output.

The span-report output for AI in June 2018 contains 236149 entries, where each
line contains one issn:

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

1. For each year, find the ISSN that actually have issues in that year.
2. For a year, shard it into 12 month and sum up issues per month.
3. Write out Excel sheet for a year.

Option: Create an advanced DataFrame:

* DateTimeIndex for each publication date, there are 49434 different dates.
* Column is a hierarchical index, issn > collection. One issn can belong to
  multiple collections, there will be 200000 columns.

Option: Create one DataFrame per sheet and then write it out.
"""

import argparse
import collections
import fileinput
import itertools
import json
import logging
import numpy as np
import os
import pandas as pd
import pickle
import sys
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
    parser = argparse.ArgumentParser()
    parser.add_argument('--publications-per-year', '-y', action='store_true',
                        help='TSV with publications per year')
    parser.add_argument('--date-list', '-d', action='store_true',
                        help='List all occuring dates')
    parser.add_argument('--data-frame', '-f', action='store_true',
                        help='Create Pandas DataFrame')
    parser.add_argument('--excel', '-x', action='store_true',
                        help='Create Excel file manually')
    args = parser.parse_args()

    entries = {}

    if not os.path.exists('r0.pkl'):
        for line in fileinput.input():
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
        yearsize = publications_per_year(entries)

        for year, size in sorted(yearsize.items()):
            print('%s\t%s' % (year, size))

    if args.date_list:
        dates = set()
        for _, doc in entries.items():
            for date in doc['dates']:
                dates.add(date)

        for date in sorted(dates):
            print(date)

    if args.data_frame:
        # Try to pack everything into a single DataFrame to take advantage of
        # DatetimeIndex.
        columns = pd.MultiIndex.from_tuples(
            [(doc['issn'], doc['sid'], doc['c']) for _, doc in entries.items()])

        dates, invalid = set(), set()

        for _, doc in entries.items():
            for date in doc['dates']:
                if not '1700-00-00' < date < '2200-00-00':
                    logger.debug('skipping date: %s', date)
                    invalid.add(date)
                    continue
                dates.add(date)

        logger.debug('skipped %d possible invalid dates', len(invalid))
        index = pd.DatetimeIndex(sorted(dates))

        # MemoryError, 11,655,842,342
        # Size:        11,656,237,206 ~ 11G
        data = np.zeros((len(index), len(columns)), dtype=np.uint16)
        df = pd.DataFrame(data, index=index, columns=columns)
        # logger.debug(df.memory_usage(index=True).sum())

        for _, doc in tqdm.tqdm(entries.items()):
            ck = (doc['issn'], doc['sid'], doc['c'])
            vs = [(date, count) for date, count in doc['dates'].items()
                  if count < 65536 and date not in invalid]
            dates, counts = [v[0] for v in vs], [v[1] for v in vs]
            # 10x faster than setting single values.
            df.loc[df.index.isin(dates), ck] = np.uint16(counts)

        df.to_hdf('df.h5', 'df')  # 22G

        logger.debug("ok")

    if args.excel:
        # 1. For each year, find the ISSN that actually have issues in that year.
        # 2. For a year, shard it into 12 month and sum up issues per month.
        # 3. Write out Excel sheet for a year.

        #             01  02  03 ...
        # ISSN SID C   0  12  31
        #      SID C   9   2   1
        # ISSN SID C   3   4   1
        # ISSN SID C   4   5   2
        # ...

        #
        # C ISSN       1   3   4
        # ...

        # C            1   2   3

        dates = set(itertools.chain(*[doc['dates'].keys() for _, doc in entries.items()]))
        logger.debug("distinct dates %s", len(dates))

        years = sorted([year for year in set([date[:4] for date in dates]) if '1700' < year < '2019'], reverse=True)
        logger.debug("distinct years %s, %s", len(years), years[:10])

        writer = pd.ExcelWriter('df.xlsx', engine='xlsxwriter')

        for year in years:
            df = pd.DataFrame()

            for month in range(1, 13):
                s = pd.Series()

                for _, doc in entries.items():
                    c, prefix = doc['c'], '%d-%02d' % (year, month)
                    for date, count in doc['dates'].items():
                        if date.startswith(prefix):
                            if c not in s:
                                s[c] = 0
                            s[c] += count

                ms = '%02d' % (month)
                df[ms] = s.sort_index()
                logger.debug("done %d-%02d", year, month)

            df.to_excel(writer, sheet_name='%d' % year)

        writer.save()
        logger.debug("ok")

