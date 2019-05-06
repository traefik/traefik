#!/usr/bin/env bash

set -eu -o pipefail

script_dir="$( cd "$( dirname "${0}" )" && pwd -P)"

if command -v shellcheck
then
    # The list of shell script come from the (grep ...) command, feeding the loop
    while IFS= read -r script_to_check
    do
        # The shellcheck command are run in background, to have an overvie of the linter (instead of a fail at first issue)
        shellcheck "${script_to_check}" &
    done < <(grep '#!' "${script_dir}"/* | cut -d':' -f1)
    wait # Wait for all background command to be completed
else
    echo "== Command shellcheck not found in your PATH. No shell script checked."
fi
