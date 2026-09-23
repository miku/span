# handling crossref full update

Crossref.org occasionally fully reindex their records; this happened in 09/2026
and the daily update process, that usually takes minutes to a few hours now is
struggling to fetch all updated records from crossref;

```
-rw-r--r--. 1 md md 199G Aug 21 08:06 feed-3-index-2026-08-17-2026-08-17.json.zst
-rw-r--r--. 1 md md 6.2G Aug 21 08:08 feed-3-index-2026-08-18-2026-08-18.json.zst
-rw-r--r--. 1 md md 2.7G Aug 21 08:08 feed-3-index-2026-08-19-2026-08-19.json.zst
-rw-r--r--. 1 md md  17G Sep  1 13:17 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-1309140327
-rw-r--r--. 1 md md 9.2G Sep  9 09:06 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-1420458665
-rw-r--r--. 1 md md 3.5G Sep  1 13:17 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-1688174208
-rw-r--r--. 1 md md  48G Sep 17 09:40 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-1695358617
-rw-r--r--. 1 md md 5.2G Sep 15 13:21 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-1832001529
-rw-r--r--. 1 md md  42G Sep 17 09:46 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-1894024874
-rw-r--r--. 1 md md  85G Sep 15 02:02 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-2086004777
-rw-r--r--. 1 md md  19G Sep 16 09:38 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-2122072648
-rw-r--r--. 1 md md  16G Sep  6 21:00 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-2318313064
-rw-r--r--. 1 md md 2.2G Sep 17 14:00 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-2435838979
-rw-r--r--. 1 md md 503M Sep 21 16:09 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-2461136986
-rw-r--r--. 1 md md  33G Sep 17 19:49 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-3184279228
-rw-r--r--. 1 md md  12G Sep 20 11:03 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-3429449822
-rw-r--r--. 1 md md 1.9G Sep 21 16:09 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-3437638962
-rw-r--r--. 1 md md  33G Sep  1 13:17 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-3496035955
-rw-r--r--. 1 md md  16G Sep 16 09:12 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-4022669824
-rw-r--r--. 1 md md  26G Sep 20 09:52 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-4143514661
-rw-r--r--. 1 md md  18G Sep 20 11:06 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-4158028194
-rw-r--r--. 1 md md  41G Sep  1 13:17 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-469964947
-rw-r--r--. 1 md md  25G Sep  1 13:17 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-629605696
-rw-r--r--. 1 md md  11G Sep  1 13:17 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-726174889
-rw-r--r--. 1 md md  21G Sep 15 08:59 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-866617010
-rw-r--r--. 1 md md  11G Sep 17 19:45 feed-3-index-2026-08-20-2026-08-20.json.zst-tmp-962607514
```

The command was:

```
span-crossref-sync -q -p zstd -c /data/finc/crossref -t 45m -mode s -verbose -s 2026-08-17 -P feed-3- -i d
```

If fetches data (likely in increments of 20) and at some point receives a HTTP
429. We already send an email address in the header/URL to identify ourselves (as suggested in the crossref API docs).

We would need a way to make the synchronisation process more resilient; it is
ok, if it takes a week, but it would be good to finish. Maybe we need a large
window (100 records per request); or somehow make the window more finegrained;
e.g. write every record we find and on an incomplete fetch try to fetch only
those records, which have been indexed at the "second" of the last update or
later; that we maybe circumventing the need to do a full fetch (we are
currently bound by daily "slices" because that was usually ok as for window
size).

Style: Keep the code compact and readable, and easy to follow.

## findings

The root cause was not rate limiting. After the first page, `WriteWindow` sent
only `cursor` and `mailto`, but the API does not keep the query with the
cursor:

```
cursor only:           {"n": 20,   "total": 186789181}   # all of crossref
cursor+filter+rows:    {"n": 1000, "total": 349025}      # 2026-08-20
```

So every harvest paged through *all* of crossref, 20 records at a time, and
`seen >= total` compared against 186M. That explains the 199G file for
2026-08-17 and the endless 2026-08-20 runs. The millions of requests then
caused the HTTP 429s. It probably worked before because the API used to keep the
query for a cursor. The large reindex only made the problem visible.

Contributing:

* pester ignores Retry-After and backs off up to 2^10s (17min), but cursors
  expire after ~5min; after one long backoff the cursor was dead
* daily cron started a new run while the old one was still going (the many
  `-tmp-` files with overlapping mtimes); the polite pool allows 3 concurrent
  requests (`x-concurrency-limit: 3`), and all runs competed for the same
  window; SIGTERM/kill left the temporary files behind
* 1000 rows is already the API maximum; the "increments of 20" were the
  default page size of the query-less cursor requests

## changes

* every request sends filter, rows, sort, order and cursor
* results are sorted by `indexed` ascending; the date filters accept
  `yyyy-MM-ddThh:mm:ss` (UTC, inclusive), so we can restart a query at the
  second of the last record written
* each page is appended to `FILE.partial` as its own gzip/zstd stream
  (concatenated streams are valid files), then `FILE.checkpoint` stores offset,
  last timestamp (ms) and the DOIs at that timestamp; on restart the partial
  file is truncated to the offset, and records at the boundary are skipped
* on a failed request (after pester retries, now capped at 30s per wait) the
  harvest waits 1m, 2m, ... 30m and resumes from the checkpoint with a fresh
  cursor; `--max-resumes` (default 20) consecutive failures without progress
  end the run, and the next cron run continues from the checkpoint
* if the cursor ends before `total-results`, a fresh query from the checkpoint
  decides: new records, continue; nothing new, the window is complete (the
  old code had a 10% tolerance heuristic instead)
* a per prefix flock (`CACHE/PREFIX-sync.lock`) prevents overlapping runs; with
  the lock held, stale `PREFIX*.json.*-tmp-*` files of earlier versions are
  removed (~700G in the listing above)
* `-mode t` now only counts (`rows=0`) and prints date and total, without
  touching the cache

Tested live against the API: killed a run of 2026-08-20 after 7 pages, it
resumed at `2026-08-20T02:29:56`, skipped the 3 records of the boundary
second, no duplicate DOI; a concurrent run is refused. The resumed run then
completed the window in 2h45m: 1.0G zst, 348809 records, no duplicate DOI
(the API reports 347890 for that day now, records keep moving to later index
dates).

## rollout

On the server, the existing `feed-3-...-2026-08-20...-tmp-*` files are removed
on the first run of the new version. The 2026-08-17 file (199G) is likely
"all of crossref" and should be deleted and fetched again; 2026-08-18 and
2026-08-19 look plausible, but can be checked with `-mode t` against the
line counts.
