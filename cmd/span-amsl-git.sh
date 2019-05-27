#!/bin/bash
#
# Copyright 2019 by Leipzig University Library, http://ub.uni-leipzig.de
#                   The Finc Authors, http://finc.info
#                   Martin Czygan, <martin.czygan@uni-leipzig.de>
#
# This file is part of some open source application.
#
# Some open source application is free software: you can redistribute
# it and/or modify it under the terms of the GNU General Public
# License as published by the Free Software Foundation, either
# version 3 of the License, or (at your option) any later version.
#
# Some open source application is distributed in the hope that it will
# be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
# of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
#
# @license GPL-3.0+ <http://spdx.org/licenses/GPL-3.0+>
#
#
# Fetch AMSL API documents (API, holding files) and commit them into a git
# repo. If run regularly, this can serve as a log of changes.
#
#   $ span-amsl-git.sh AMSL-API-URL WORK-TREE [GIT-DIR]
#
# Example:
#
#   $ span-amsl-git.sh https://example.amsl.technology /var/somerepo
#
set -e -u -o pipefail

if [ "$#" -lt 2 ]; then
    echo "Usage: $0 AMSL-API-URL WORK-TREE [GIT-DIR]"
    exit 1
fi

AMSL_API_URL=$1
WORK_TREE=$2
GIT_DIR=${3:-$WORK_TREE/.git}

echo >&2 "using: $AMSL_API_URL $WORK_TREE $GIT_DIR"

for req in curl jq span-amsl-discovery; do
    command -v $req >/dev/null 2>&1 || {
        echo >&2 "$req required"
        exit 1
    }
done

if [ ! -d "$WORK_TREE" ]; then
    echo "$WORK_TREE is not a directory"
    exit 1
fi

if [ ! -d "$GIT_DIR" ]; then
    echo "$GIT_DIR not found or not a directory"
    exit 1
fi

# Fetch smaller APIs separately.
for api in metadata_usage holdings_file_concat holdingsfiles contentfiles; do
    curl -s --fail "$AMSL_API_URL/outboundservices/list?do=$api" | jq -r --sort-keys . > "$WORK_TREE/$api.json"
done

# Fetch combined API as well.
span-amsl-discovery -live "$AMSL_API_URL" | jq -r --sort-keys . > "$WORK_TREE/discovery.json"

# Fetch holding files, assume that an URI looks like
# http://amsl.technology/discovery/metadata-usage/Dokument/KBART_FREEJOURNALS,
# we utilize the unique base names, e.g. KBART_FREEJOURNALS. Note: Files may or
# may not be compressed, text would be nicer to diff.
if [ -f "$WORK_TREE/holdingsfiles.json" ]; then
    for uri in $(cat "$WORK_TREE/holdingsfiles.json" | jq -r '.[].DokumentURI' | sort -u); do
        if [ -z "$uri" ]; then
            continue
        fi
        name=$(basename $uri)
        if [ -z "$name" ]; then
            continue
        fi
        link="$AMSL_API_URL/OntoWiki/files/get?setResource=$uri"
        mkdir -p "$WORK_TREE/h/"
        curl -s --fail "$link" > "$WORK_TREE/h/$name.tsv"
    done
fi

# Commit, and push to a remote named origin.
if [[ $(git --git-dir "$GIT_DIR" --work-tree "$WORK_TREE" status --porcelain) ]]; then
    date > "$WORK_TREE/.date"
    git --git-dir "$GIT_DIR" --work-tree "$WORK_TREE" add --all
    git --git-dir "$GIT_DIR" --work-tree "$WORK_TREE" commit -m "auto-commit from $(hostname) [$$]"
    git --git-dir "$GIT_DIR" --work-tree "$WORK_TREE" push origin master
else
    exit 0
fi
