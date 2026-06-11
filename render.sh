#!/usr/bin/env bash

set -euo pipefail

INPUT_DIR="build/dot"
OUTPUT_DIR="build/svg"

find "$INPUT_DIR" -type f -name "*.dot" | while read -r dotfile; do
    relpath="${dotfile#$INPUT_DIR/}"

    svgfile="${OUTPUT_DIR}/${relpath%.dot}.svg"

    mkdir -p "$(dirname "$svgfile")"

    echo "Converting: $dotfile -> $svgfile"

    dot -Tsvg "$dotfile" -o "$svgfile"
done

echo "Done."