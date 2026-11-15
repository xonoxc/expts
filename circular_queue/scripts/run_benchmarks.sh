#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUNS="${RUNS:-5}"

CONFIGS=("1 1" "1 4" "4 1" "4 4")

for config in "${CONFIGS[@]}"; do
    read -r producers consumers <<< "$config"

    name="bench_${producers}_${consumers}"

    echo
    echo "===== $name ====="

    make -s -C "$ROOT" bench-build \
        PRODUCERS="$producers" \
        CONSUMERS="$consumers"

    binary="$ROOT/build/bin/$name"
    results=()

    for ((i=1; i<=RUNS; i++)); do
        output=$("$binary")
        throughput=$(awk '/Items\/sec:/ {print $2; exit}' <<< "$output")

        [[ "$throughput" =~ ^[0-9]+([.][0-9]+)?$ ]] || {
            echo "Invalid throughput: $throughput"
            exit 1
        }

        results+=("$throughput")
        printf 'Run %d: %s items/sec\n' "$i" "$throughput"
    done

    median=$(printf '%s\n' "${results[@]}" |
        sort -n |
        awk '{a[NR]=$1} END {
            if (NR % 2) print a[(NR+1)/2]
            else print (a[NR/2]+a[NR/2+1])/2
        }')

    echo "Median: $median items/sec"
done
