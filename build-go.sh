#!/bin/bash
set -e 

ABSOLUTE_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_ROOT=${SOURCE_ROOT:-$ABSOLUTE_PATH}
GO=$(command -v go 2>/dev/null)
GG_OUT=${BUILT_PRODUCTS_DIR:-$SOURCE_ROOT}
GG_OBJ=$GG_OUT/go-obj
GG_CGO_OBJ=$GG_OUT/cgo-obj
TUMBLRTVDIR=$SOURCE_ROOT/tumblrtv2

echo "stepping into $SOURCE_ROOT"
cd $SOURCE_ROOT
echo "compiling go code..."
CGO_ENABLED=1 GO111MODULE=on $GO build -buildmode=c-archive -ldflags '-linkmode external' -o $TUMBLRTVDIR/go.a
