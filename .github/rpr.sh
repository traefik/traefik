#!/bin/sh
#
# git config --global alias.rpr '!sh .github/rpr.sh'

set -e # stop on error

usage="$(basename "$0") pr -- rebase a Pull Request against current branch"

if [ "$#" -ne 1 ]; then
    echo "Illegal number of parameters"
    echo "$usage" >&2
    exit 1
fi

command -v jq >/dev/null 2>&1 || { echo "I require jq but it's not installed.  Aborting." >&2; exit 1; }

set -x # echo on

base=$(git rev-parse --abbrev-ref HEAD)
pr=$1
remote=$(curl -s https://api.github.com/repos/containous/traefik/pulls/$pr | jq -r .head.repo.owner.login)
branch=$(curl -s https://api.github.com/repos/containous/traefik/pulls/$pr | jq -r .head.ref)

git checkout $base

git remote add $remote git@github.com:$remote/traefik.git
git fetch $remote $branch
git checkout -t $remote/$branch
git rebase origin/$base
git push -f $remote $branch

# clean
git checkout $base
git branch -D $branch
git remote remove $remote