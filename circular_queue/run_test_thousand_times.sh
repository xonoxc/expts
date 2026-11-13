#!/bin/bash


for i in $(seq 1 1000); do
    ./test_queue --filter='queue/concurrent_producers_and_consumers' -q || {
        echo "Failed at iteration $i"
        exit 1
    }
done

echo "All 1000 runs passed"
