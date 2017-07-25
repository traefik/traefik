#!/usr/bin/env bash
set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that directory
cd "$DIR"

# Generate the tag
DEPLOY="../pkg/web_ui"

rm -rf $DEPLOY
mkdir -p $DEPLOY

bundle check >/dev/null 2>&1 || bundle install
bundle exec sass styles/base.scss static/base.css

bundle exec ruby scripts/compile.rb

# Copy into deploy
shopt -s dotglob
cp -r $DIR/static $DEPLOY/
cp index.html $DEPLOY/

# Magic scripting
sed -E -e "/ASSETS/,/\/ASSETS/ d" -ibak $DEPLOY/index.html
sed -E -e "s#<\/body>#<script src=\"static/application.min.js\"></script></body>#" -ibak $DEPLOY/index.html

# Remove the backup file from sed
rm $DEPLOY/index.htmlbak

pushd $DEPLOY >/dev/null 2>&1
zip -r ../web_ui.zip ./*
popd >/dev/null 2>&1
