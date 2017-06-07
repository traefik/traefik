#!/usr/bin/env bash
#
# git config --global alias.cpr '!sh .github/cpr.sh'

set -e # stop on error

usage="$(basename "$0") pr -- Checkout a Pull Request locally"

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

git remote add $remote git@github.com:$remote/traefik.git
git fetch $remote $branch
git checkout -t -b "$pr--$branch" $remote/$branch