#!/bin/bash

# Fetch AMSL API JSON document and commit documents into a git repo. If run
# regularly, this can serve as a log of changes.
#
#   $ span-amsl-git https://example.amsl.technology /var/amsl-sync
#

set -e -o pipefail

if [ "$#" -lt 2 ]; then
    echo "Usage: $0 AMSL-API-URL WORK-TREE [GIT-DIR]"
    exit 1
fi

AMSL_API_URL=$1
WORK_TREE=$2
GIT_DIR=${3:-$WORK_TREE/.git}

command -v curl >/dev/null 2>&1 || {
    echo >&2 "curl required, https://curl.haxx.se/"
    exit 1
}

command -v span-amsl-discovery >/dev/null 2>&1 || {
    echo >&2 "span-amsl-discovery required, https://github.com/miku/span/releases"
    exit 1
}

if [ ! -d "$WORK_TREE" ]; then
    echo "$WORK_TREE is not a directory"
    exit 1
fi

if [ ! -d "$GIT_DIR" ]; then
    echo "$GIT_DIR not found or not a directory"
    exit 1
fi

# Fetch smaller APIs separately.
curl -s --fail "$AMSL_API_URL/outboundservices/list?do=metadata_usage" | jq -r --sort-keys . > $WORK_TREE/metadata_usage.json
curl -s --fail "$AMSL_API_URL/outboundservices/list?do=holdings_file_concat" | jq -r --sort-keys . > $WORK_TREE/holdings_file_concat.json
curl -s --fail "$AMSL_API_URL/outboundservices/list?do=holdingsfiles" | jq -r --sort-keys . > $WORK_TREE/holdingsfiles.json
curl -s --fail "$AMSL_API_URL/outboundservices/list?do=contentfiles" | jq -r --sort-keys . > $WORK_TREE/contentfiles.json

# Fetch combined API as well.
span-amsl-discovery -live $AMSL_API_URL | jq -r --sort-keys . >$WORK_TREE/discovery.json

# Commit.
if [[ $(git status --porcelain) ]]; then
    git --git-dir $GIT_DIR --work-tree $WORK_TREE add --all
    git --git-dir $GIT_DIR --work-tree $WORK_TREE commit -m "auto-commit from $(hostname) by $(whoami)"
else
    exit 0
fi

