#!/bin/sh
#
# git config --global alias.rmpr '!sh .github/rmpr.sh'

set -e # stop on error

usage="$(basename "$0") pr -- remove a Pull Request local branch & remote"

if [ "$#" -ne 1 ]; then
    echo "Illegal number of parameters"
    echo "$usage" >&2
    exit 1
fi

command -v jq >/dev/null 2>&1 || { echo "I require jq but it's not installed.  Aborting." >&2; exit 1; }

set -x # echo on

initial=$(git rev-parse --abbrev-ref HEAD)
pr=$1
remote=$(curl -s https://api.github.com/repos/containous/traefik/pulls/$pr | jq -r .head.repo.owner.login)
branch=$(curl -s https://api.github.com/repos/containous/traefik/pulls/$pr | jq -r .head.ref)

# clean
git checkout $initial
git branch -D "$pr--$branch"
git remote remove $remote