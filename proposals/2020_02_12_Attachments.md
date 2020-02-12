# Attachments

Currently, `kbart.Holdings` can be loaded (e.g. from a file), which allows
derive some helper indices, such as `h.SerialNumberMap` which allows to find
the relevant entry for an ISSN.

Trying to move to a tabular form, the following steps seem to be required:

* read config table (span-amsl-discovery -f ...) into memory
* (1) determine if record values (sid, collection) is in table
* (2) if so, check if there is a holding file to consider - if yes, find relevant
  entries from holding file and match agains record values (date, volume,
issue)

For (1), we can map (sid, collection) and (sid, tcid) to indices to entries.

