#!/usr/bin/env bash
set -euo pipefail

APP=poe2campain
OUT_DIR=.build

case "$OSTYPE" in
    msys*|cygwin*)
        APP+=.exe
        ;;
esac

mkdir -p "$OUT_DIR"
go build -ldflags='-s -w' -o "$OUT_DIR/$APP" ./cmd/poe2campain

rm -rf "$OUT_DIR/data"
cp -r data "$OUT_DIR/"

echo "Built $OUT_DIR/$APP"
