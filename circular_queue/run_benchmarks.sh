

#!/usr/bin/env bash
set -euo pipefail

RUNS=5

BENCHMARKS=(
  bench_1_1
  bench_1_4
  bench_4_1
  bench_4_4
)

for bench in "${BENCHMARKS[@]}"; do
    echo
    echo "===== $bench ====="

    results=()

    for ((i=1; i<=RUNS; i++)); do
        output=$(./"$bench")
        throughput=$(awk '/Items\/sec:/ {print $2}' <<< "$output")

        results+=("$throughput")
        printf "Run %d: %s items/sec\n" "$i" "$throughput"
    done

    median=$(printf '%s\n' "${results[@]}" |
        sort -n |
        awk '{a[NR]=$1} END {print a[(NR+1)/2]}')

    echo "Median: $median items/sec"
done
