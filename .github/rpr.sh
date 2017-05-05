#!/usr/bin/env bash
#
# git config --global alias.rpr '!sh .github/rpr.sh'

set -e # stop on error

usage="$(basename "$0") pr remote/branch -- rebase a Pull Request against a remote branch"

if [ "$#" -ne 2 ]; then
    echo "Illegal number of parameters"
    echo "$usage" >&2
    exit 1
fi

command -v jq >/dev/null 2>&1 || { echo "I require jq but it's not installed.  Aborting." >&2; exit 1; }

set -x # echo on

initial=$(git rev-parse --abbrev-ref HEAD)
pr=$1
base=$2
remote=$(curl -s https://api.github.com/repos/containous/traefik/pulls/$pr | jq -r .head.repo.owner.login)
branch=$(curl -s https://api.github.com/repos/containous/traefik/pulls/$pr | jq -r .head.ref)

clean ()
{
    git checkout $initial
    .github/rmpr.sh $pr
}

trap clean EXIT

.github/cpr.sh $pr

git rebase $base
git push --force-with-lease $remote $pr--$branch:$branch
