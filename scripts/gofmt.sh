#!/bin/bash

# Copyright Greg Haskins All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

set -e

# shellcheck source=/dev/null
echo "Checking with gofmt"
sources="./"
OUTPUT=$(gofmt -e -d -s -l $sources)
if [ ! -z "$OUTPUT" ]; then
    # Some files are not gofmt.
    echo >&2 "The following Go files must be formatted with gofmt:"
    for fn in $OUTPUT; do
        echo >&2 "  $fn"
    done
    echo >&2 "Please run 'make format'."
    exit 1
fi

echo "Checking with goimports"
OUTPUT="$(goimports -l "$sources")"
if [ -n "$OUTPUT" ]; then
    echo "The following files contain goimports errors"
    echo "$OUTPUT"
    echo "The goimports command 'goimports -l -w' must be run for these files"
    exit 1
fi

echo "Checking with go vet"
PRINTFUNCS="Debug,Debugf,Print,Printf,Info,Infof,Warning,Warningf,Error,Errorf,Critical,Criticalf,Sprint,Sprintf,Log,Logf,Panic,Panicf,Fatal,Fatalf,Notice,Noticef,Wrap,Wrapf,WithMessage"
OUTPUT="$(go vet -all -printfuncs "$PRINTFUNCS" ./...)"
if [ -n "$OUTPUT" ]; then
    echo "The following files contain go vet errors"
    echo "$OUTPUT"
    #exit 1
fi

exit 0
