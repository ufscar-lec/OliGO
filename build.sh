#!/bin/sh

set -eu

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
    ZIPFILE="build/${GOOS}-${GOARCH}.zip"

    mkdir -p "$OUTDIR"

    EXT=""
    if [ "$GOOS" = "windows" ]; then
        EXT=".exe"
    fi

    echo "Building OliGO for $GOOS/$GOARCH..."
    GOOS="$GOOS" GOARCH="$GOARCH" \
        go build -o "$OUTDIR/OliGO$EXT" "./cmd/OliGO"

    echo "Creating $ZIPFILE..."
    (
        cd build
        zip -r "${GOOS}-${GOARCH}.zip" "${GOOS}-${GOARCH}"
    )

    rm -rf "$OUTDIR"
done