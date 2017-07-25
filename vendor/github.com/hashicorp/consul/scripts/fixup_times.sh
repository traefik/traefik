#!/usr/bin/env bash
set -e
regex='bindataFileInfo.*name: \"(.+)\".*time.Unix.(.+),'
while read line; do
    if [[ $line =~ $regex ]]; then
        file=${BASH_REMATCH[1]}
        ts=${BASH_REMATCH[2]}
        touch --date @$ts $file
    fi
done
