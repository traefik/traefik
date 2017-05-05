#!/usr/bin/env bash
set -e

pip install --user -r requirements.txt

make pull-images
make validate
