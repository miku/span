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
# Example (the git repo must exist, but can be empty):
#
#   $ span-amsl-git.sh https://example.amsl.technology /var/some-git-repo
#
set -u -o pipefail

if [ "$#" -lt 2 ]; then
    echo "Usage: $0 AMSL-API-URL WORK-TREE [GIT-DIR]"
    exit 1
fi

AMSL_API_URL=$1
WORK_TREE=$2
GIT_DIR=${3:-$WORK_TREE/.git}

# A separate dir for holdings files.
HOLDINGSFILES_DIR="$WORK_TREE/h"
mkdir -p "$HOLDINGSFILES_DIR"

echo >&2 "using: $AMSL_API_URL $WORK_TREE $GIT_DIR"

for req in curl jq unzip zipinfo span-amsl-discovery; do
    command -v $req >/dev/null 2>&1 || {
        echo >&2 "$req required"
        exit 1
    }
done

for dir in $WORK_TREE $GIT_DIR $HOLDINGSFILES_DIR; do
    if [ ! -d "$dir" ]; then
        echo "$dir is not a directory"
        exit 1
    fi
done

# Fetch smaller APIs separately.
for api in metadata_usage holdings_file_concat holdingsfiles contentfiles; do
    curl -s --fail "$AMSL_API_URL/outboundservices/list?do=$api" | jq -r --sort-keys . >"$WORK_TREE/$api.json"
done

# Fetch combined API as well.
span-amsl-discovery -live "$AMSL_API_URL" | jq -r --sort-keys . >"$WORK_TREE/discovery.json"

# Fetch holding files, assume that an URI looks like
# http://amsl.technology/discovery/metadata-usage/Dokument/KBART_FREEJOURNALS,
# we utilize the unique base names, e.g. KBART_FREEJOURNALS. Note: Files may or
# may not be compressed, text would be nicer to diff.
if [ -f "$WORK_TREE/holdingsfiles.json" ]; then
    for uri in $(jq -r '.[].DokumentURI' <"$WORK_TREE/holdingsfiles.json" | sort -u); do
        if [ -z "$uri" ]; then
            continue
        fi

        name=$(basename "$uri")
        if [ -z "$name" ]; then
            continue
        fi

        link="$AMSL_API_URL/OntoWiki/files/get?setResource=$uri"
        target="$HOLDINGSFILES_DIR/$name.tsv"
        tmp="$target.tmp"

        curl -s --fail "$link" >"$tmp"

        # Test if zip, non-zip might fail with 9.
        if unzip -z "$tmp" >/dev/null 2>&1; then
            filecount=$(zipinfo -t "$tmp" | awk '{print $1}')
            if [[ "$filecount" -ne 1 ]]; then
                echo "expected single file in zip $tmp, got $filecount"
                exit 1
            else
                unzip -p "$tmp" >"$target" && rm -f "$tmp"
            fi
        else
            # Assume already plain text.
            mv "$tmp" "$target"
        fi
    done
fi

# Commit, and push to a remote named origin.
if [[ $(git --git-dir "$GIT_DIR" --work-tree "$WORK_TREE" status --porcelain) ]]; then
    date >"$WORK_TREE/.date"
    git --git-dir "$GIT_DIR" --work-tree "$WORK_TREE" add --all
    git --git-dir "$GIT_DIR" --work-tree "$WORK_TREE" commit -m "auto-commit from $(hostname) [$$]"
    git --git-dir "$GIT_DIR" --work-tree "$WORK_TREE" push origin master
else
    exit 0
fi
