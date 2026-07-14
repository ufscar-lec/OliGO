#!/bin/sh

set -eu

BINARIES="blockParse filterProbes"

TARGETS="
linux amd64
linux arm64
windows amd64
windows arm64
darwin amd64
darwin arm64
"

echo "$TARGETS" | while read GOOS GOARCH; do
    [ -n "$GOOS" ] || continue

    OUTDIR="build/${GOOS}-${GOARCH}"
    mkdir -p "$OUTDIR"

    EXT=""
    if [ "$GOOS" = "windows" ]; then
        EXT=".exe"
    fi

    for BIN in $BINARIES; do
        echo "Building $BIN for $GOOS/$GOARCH..."
        GOOS="$GOOS" GOARCH="$GOARCH" \
            go build -o "$OUTDIR/$BIN$EXT" "./cmd/$BIN"
    done
done
