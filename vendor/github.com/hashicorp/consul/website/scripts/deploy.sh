#!/bin/sh
set -e

PROJECT="consul"
PROJECT_URL="www.consul.io"
FASTLY_SERVICE_ID="7GrxRJP3PVBuqQbyxYQ0MV"

# Ensure the proper AWS environment variables are set
if [ -z "$AWS_ACCESS_KEY_ID" ]; then
  echo "Missing AWS_ACCESS_KEY_ID!"
  exit 1
fi

if [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
  echo "Missing AWS_SECRET_ACCESS_KEY!"
  exit 1
fi

# Ensure the proper Fastly keys are set
if [ -z "$FASTLY_API_KEY" ]; then
  echo "Missing FASTLY_API_KEY!"
  exit 1
fi

# Ensure we have s3cmd installed
if ! command -v "s3cmd" >/dev/null 2>&1; then
  echo "Missing s3cmd!"
  exit 1
fi

# Get the parent directory of where this script is and cd there
DIR="$(cd "$(dirname "$(readlink -f "$0")")/.." && pwd)"

# Delete any .DS_Store files for our OS X friends.
find "$DIR" -type f -name '.DS_Store' -delete

# Upload the files to S3 - we disable mime-type detection by the python library
# and just guess from the file extension because it's surprisingly more
# accurate, especially for CSS and javascript. We also tag the uploaded files
# with the proper Surrogate-Key, which we will later purge in our API call to
# Fastly.
if [ -z "$NO_UPLOAD" ]; then
  echo "Uploading to S3..."

  # Check that the site has been built
  if [ ! -d "$DIR/build" ]; then
    echo "Missing compiled website! Run 'make build' to compile!"
    exit 1
  fi

  # Set browser-side cache-control to ~4h, but tell Fastly to cache for much
  # longer. We manually purge the Fastly cache, so setting it to a year is more
  # than fine.
  s3cmd \
    --quiet \
    --delete-removed \
    --guess-mime-type \
    --no-mime-magic \
    --acl-public \
    --recursive \
    --add-header="Cache-Control: max-age=14400" \
    --add-header="x-amz-meta-surrogate-control: max-age=31536000, stale-white-revalidate=86400, stale-if-error=604800" \
    --add-header="x-amz-meta-surrogate-key: site-$PROJECT" \
    sync "$DIR/build/" "s3://hc-sites/$PROJECT/latest/"

  # The s3cmd guessed mime type for text files is often wrong. This is
  # problematic for some assets, so force their mime types to be correct.
  echo "Overriding javascript mime-types..."
  s3cmd \
    --mime-type="application/javascript" \
    --add-header="Cache-Control: max-age=31536000" \
    --exclude "*" \
    --include "*.js" \
    --recursive \
    modify "s3://hc-sites/$PROJECT/latest/"

  echo "Overriding css mime-types..."
  s3cmd \
    --mime-type="text/css" \
    --add-header="Cache-Control: max-age=31536000" \
    --exclude "*" \
    --include "*.css" \
    --recursive \
    modify "s3://hc-sites/$PROJECT/latest/"

  echo "Overriding svg mime-types..."
  s3cmd \
    --mime-type="image/svg+xml" \
    --add-header="Cache-Control: max-age=31536000" \
    --exclude "*" \
    --include "*.svg" \
    --recursive \
    modify "s3://hc-sites/$PROJECT/latest/"
fi

# Perform a purge of the surrogate key.
if [ -z "$NO_PURGE" ]; then
  echo "Purging Fastly cache..."
  curl \
    --fail \
    --silent \
    --output /dev/null \
    --request "POST" \
    --header "Accept: application/json" \
    --header "Fastly-Key: $FASTLY_API_KEY" \
    --header "Fastly-Soft-Purge: 1" \
    "https://api.fastly.com/service/$FASTLY_SERVICE_ID/purge/site-$PROJECT"
fi

# Warm the cache with recursive wget.
if [ -z "$NO_WARM" ]; then
  echo "Warming Fastly cache..."
  echo ""
  echo "If this step fails, there are likely missing or broken assets or links"
  echo "on the website. Run the following command manually on your laptop, and"
  echo "search for \"ERROR\" in the output:"
  echo ""
  echo "wget --recursive --delete-after https://$PROJECT_URL/"
  echo ""
  wget \
    --recursive \
    --delete-after \
    --quiet \
    "https://$PROJECT_URL/"
fi
