#!/usr/bin/env bash

APP=poe2campain

case "$OSTYPE" in
    msys*|cygwin*)
        APP+=.exe
        ;;
esac

go build -ldflags='-s -w' -o dist/$APP ./cmd/poe2campain

cp -r data dist/
