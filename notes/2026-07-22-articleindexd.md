# articleindexd

status: open

## Design proposal

A single service to kick off and coordinate article index related task. Could
be connected to a web frontend to manage tasks.

Requirements:

* regular, atomic data acquisision tasks
* plausibility checks
* periodic cleanups
* reporting and measurements

Background:

> Requirements are well-known at this point to bake most of the functionality
> into one or a few binaries. Can keep most of the directory hierarchy, e.g. by
> source, task, etc. - but could get rid of the scheduling (or some basic
> task/dir/filename based matching and dependency management).

File based control, e.g. inject whitelists, blacklists into specific,
documented directories to be picked up, or a config dir, etc.
