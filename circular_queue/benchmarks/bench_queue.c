#include "queue.h"
#include <pthread.h>
#include <stdatomic.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

#ifndef PRODUCERS
#define PRODUCERS 4
#endif

#ifndef CONSUMERS
#define CONSUMERS 4
#endif

#ifndef ITEMS_PER_PRODUCER
#define ITEMS_PER_PRODUCER 1000000
#endif

#ifndef CAPACITY
#define CAPACITY 64
#endif

#define TOTAL_ITEMS (PRODUCERS * ITEMS_PER_PRODUCER)

static Queue queue;

static atomic_int started = 0;
static atomic_int producers_done = 0;
static atomic_int errors = 0;

static atomic_ulong full_retries = 0;
static atomic_ulong empty_retries = 0;

typedef struct {
	int id;
	long consumed;
} ConsumerArgs;


static void wait_for_start(void)
{
	while (!atomic_load(&started))
		sched_yield();
}

static void *producer(void *arg)
{
	int id = *(int *)arg;
	int first = id * ITEMS_PER_PRODUCER + 1;

	wait_for_start();

	for (int i = first; i < first + ITEMS_PER_PRODUCER; i++) {
		QueueError err;

		while ((err = enqueue(&queue, i)) == QUEUE_FULL) {
			atomic_fetch_add(&full_retries , 1);
			sched_yield();
		}

		if (err != QUEUE_OK) {
			atomic_store(&errors, 1);
			break;
		}
	}

	atomic_fetch_add(&producers_done, 1);
	return NULL;
}

static void *consumer(void *arg)
{
	ConsumerArgs *args = arg;

	wait_for_start();

	for (;;) {
		int value;
		QueueError err = dequeue(&queue, &value);

		if (err == QUEUE_OK) {
			args->consumed++;
			continue;
		}

		if (err == QUEUE_EMPTY) {
			atomic_fetch_add(&empty_retries, 1);

			if (atomic_load(&producers_done) == PRODUCERS)
				break;

			sched_yield();
			continue;
		}

		atomic_store(&errors, 1);
		break;
	}

	return NULL;
}

static double elapsed_seconds(
	struct timespec start,
	struct timespec end)
{
	return (end.tv_sec - start.tv_sec) +
	(end.tv_nsec - start.tv_nsec) / 1e9;
}

int main(void)
{
	pthread_t producer_threads[PRODUCERS];
	pthread_t consumer_threads[CONSUMERS];

	int producer_ids[PRODUCERS];
	ConsumerArgs consumer_args[CONSUMERS] = {0};

	if (queue_init(&queue, CAPACITY) != QUEUE_OK) {
		fprintf(stderr, "Queue initialization failed\n");
		return EXIT_FAILURE;
	}

	for (int i = 0; i < CONSUMERS; i++) {
		consumer_args[i].id = i;

		if (pthread_create(&consumer_threads[i], NULL,
					 consumer, &consumer_args[i]) != 0) {
			fprintf(stderr, "Consumer creation failed\n");
			return EXIT_FAILURE;
		}
	}

	for (int i = 0; i < PRODUCERS; i++) {
		producer_ids[i] = i;

		if (pthread_create(&producer_threads[i], NULL,
					 producer, &producer_ids[i]) != 0) {
			fprintf(stderr, "Producer creation failed\n");
			return EXIT_FAILURE;
		}
	}

	struct timespec start, end;

	clock_gettime(CLOCK_MONOTONIC, &start);
	atomic_store(&started, 1);

	for (int i = 0; i < PRODUCERS; i++)
		pthread_join(producer_threads[i], NULL);

	for (int i = 0; i < CONSUMERS; i++)
		pthread_join(consumer_threads[i], NULL);

	clock_gettime(CLOCK_MONOTONIC, &end);

	double elapsed = elapsed_seconds(start, end);

	long total_consumed = 0;

	for (int i = 0; i < CONSUMERS; i++)
		total_consumed += consumer_args[i].consumed;

	printf("\n--- Queue Benchmark ---\n");
	printf("Producers:       %d\n", PRODUCERS);
	printf("Consumers:       %d\n", CONSUMERS);
	printf("Capacity:        %d\n", CAPACITY);
	printf("Items produced:  %d\n", TOTAL_ITEMS);
	printf("Items consumed:  %ld\n", total_consumed);
	printf("Elapsed:         %.4f seconds\n", elapsed);
	printf("Items/sec:       %.0f\n", total_consumed / elapsed);
	printf("Queue ops/sec:   %.0f\n",
		(2.0 * total_consumed) / elapsed);
	printf("Errors:          %d\n", atomic_load(&errors));
	printf("Full retries:  %lu\n", atomic_load(&full_retries));
	printf("Empty retries: %lu\n", atomic_load(&empty_retries));

	queue_destroy(&queue);

	return atomic_load(&errors) ||
	total_consumed != TOTAL_ITEMS
	? EXIT_FAILURE : EXIT_SUCCESS;
}
