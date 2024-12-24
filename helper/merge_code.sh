#!/bin/bash

MERGED_DIR=merged/cloudpods

mkdir -p $MERGED_DIR

for dir in $(find . -type d -depth 1); do
    echo copy $dir to $MERGED_DIR
    rsync -avP $dir/ $MERGED_DIR
done

echo "Merged code: $MERGED_DIR"
