#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUNS="${RUNS:-1000}"
TEST_BIN="$ROOT/build/bin/test_queue"

make -s -C "$ROOT" "$TEST_BIN"

for ((i=1; i<=RUNS; i++)); do
    if ! "$TEST_BIN" \
        --filter='queue/concurrent_producers_and_consumers' -q; then
        echo "Failed at iteration $i"
        exit 1
    fi
done

echo "All $RUNS runs passed"
