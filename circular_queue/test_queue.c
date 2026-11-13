#include <criterion/criterion.h>
#include "queue.h"
#include <pthread.h>
#include <stdatomic.h>
#include <sched.h>
#include <stdlib.h>


#define PRODUCER_COUNT 4
#define CONSUMER_COUNT 4
#define ITEMS_PER_PRODUCER 1000
#define TOTAL_ITEMS (PRODUCER_COUNT * ITEMS_PER_PRODUCER)

typedef struct {
	Queue *queue;
	int start;
	int count;
	atomic_int *failed;
} ProducerArgs;

typedef struct {
	Queue *queue;
	atomic_int *consumed;
	atomic_int *failed;
	atomic_int *seen;
} ConsumerArgs;


Test(queue, initializes_empty)
{
	Queue q;

	cr_assert_eq(queue_init(&q, 5), QUEUE_OK);
	cr_assert_eq(q.size, 0);
	cr_assert_eq(q.capacity, 5);
	cr_assert_eq(q.head, 0);
	cr_assert_eq(q.tail, 0);

	cr_assert_eq(queue_destroy(&q), QUEUE_OK);
}

Test(queue, enqueue_and_dequeue)
{
	Queue q;
	int value;

	cr_assert_eq(queue_init(&q, 5), QUEUE_OK);

	cr_assert_eq(enqueue(&q, 10), QUEUE_OK);
	cr_assert_eq(enqueue(&q, 20), QUEUE_OK);
	cr_assert_eq(enqueue(&q, 30), QUEUE_OK);

	cr_assert_eq(dequeue(&q, &value), QUEUE_OK);
	cr_assert_eq(value, 10);

	cr_assert_eq(dequeue(&q, &value), QUEUE_OK);
	cr_assert_eq(value, 20);

	cr_assert_eq(dequeue(&q, &value), QUEUE_OK);
	cr_assert_eq(value, 30);

	cr_assert_eq(queue_destroy(&q), QUEUE_OK);
}

Test(queue, rejects_enqueue_when_full)
{
	Queue q;

	cr_assert_eq(queue_init(&q, 2), QUEUE_OK);

	cr_assert_eq(enqueue(&q, 10), QUEUE_OK);
	cr_assert_eq(enqueue(&q, 20), QUEUE_OK);
	cr_assert_eq(enqueue(&q, 30), QUEUE_FULL);

	cr_assert_eq(q.size, 2);
	cr_assert_eq(queue_destroy(&q), QUEUE_OK);
}

Test(queue, rejects_dequeue_when_empty)
{
	Queue q;
	int value;

	cr_assert_eq(queue_init(&q, 5), QUEUE_OK);

	cr_assert_eq(dequeue(&q, &value), QUEUE_EMPTY);
	cr_assert_eq(q.size, 0);

	cr_assert_eq(queue_destroy(&q), QUEUE_OK);
}

Test(queue, wraps_around)
{
	Queue q;
	int value;

	cr_assert_eq(queue_init(&q, 3), QUEUE_OK);

	cr_assert_eq(enqueue(&q, 10), QUEUE_OK);
	cr_assert_eq(enqueue(&q, 20), QUEUE_OK);
	cr_assert_eq(enqueue(&q, 30), QUEUE_OK);

	cr_assert_eq(dequeue(&q, &value), QUEUE_OK);
	cr_assert_eq(value, 10);

	cr_assert_eq(enqueue(&q, 40), QUEUE_OK);

	cr_assert_eq(dequeue(&q, &value), QUEUE_OK);
	cr_assert_eq(value, 20);

	cr_assert_eq(dequeue(&q, &value), QUEUE_OK);
	cr_assert_eq(value, 30);

	cr_assert_eq(dequeue(&q, &value), QUEUE_OK);
	cr_assert_eq(value, 40);

	cr_assert_eq(q.size, 0);
	cr_assert_eq(queue_destroy(&q), QUEUE_OK);
}

Test(queue, repeated_wraparound)
{
	Queue q;
	int value;

	cr_assert_eq(queue_init(&q, 3), QUEUE_OK);

	for (int i = 0; i < 100; i++) {
		cr_assert_eq(enqueue(&q, i), QUEUE_OK);
		cr_assert_eq(dequeue(&q, &value), QUEUE_OK);
		cr_assert_eq(value, i);
	}

	cr_assert_eq(q.size, 0);

	cr_assert_eq(pthread_mutex_lock(&q.mutex), 0);
	cr_assert_eq(pthread_mutex_unlock(&q.mutex), 0);

	cr_assert_eq(queue_destroy(&q), QUEUE_OK);
}






static void *producer(void *arg)
{
	ProducerArgs *args = arg;

	for (int i = 0; i < args->count; i++) {
		int value = args->start + i;

		while (!atomic_load(args->failed)) {
			QueueError err = enqueue(args->queue, value);

			if (err == QUEUE_OK)
				break;

			if (err != QUEUE_FULL) {
				atomic_store(args->failed, 1);
				return NULL;
			}

			sched_yield();
		}

		if (atomic_load(args->failed))
			return NULL;
	}

	return NULL;
}

static void *consumer(void *arg)
{
	ConsumerArgs *args = arg;

	while (atomic_load(args->consumed) < TOTAL_ITEMS &&
		!atomic_load(args->failed)) {

		int value;
		QueueError err = dequeue(args->queue, &value);

		if (err == QUEUE_EMPTY) {
			sched_yield();
			continue;
		}

		if (err != QUEUE_OK) {
			atomic_store(args->failed, 1);
			return NULL;
		}

		if (value < 1 || value > TOTAL_ITEMS) {
			atomic_store(args->failed, 1);
			return NULL;
		}

		atomic_fetch_add(&args->seen[value], 1);
		atomic_fetch_add(args->consumed, 1);
	}

	return NULL;
}

Test(queue, concurrent_producers_and_consumers)
{
	Queue q;
	pthread_t producers[PRODUCER_COUNT];
	pthread_t consumers[CONSUMER_COUNT];

	atomic_int consumed;
	atomic_int failed;

	atomic_init(&consumed, 0);
	atomic_init(&failed, 0);

	atomic_int *seen = malloc(sizeof(*seen) * (TOTAL_ITEMS + 1));
	cr_assert_not_null(seen);

	for (int i = 0; i <= TOTAL_ITEMS; i++)
		atomic_init(&seen[i], 0);

	cr_assert_eq(queue_init(&q, 64), QUEUE_OK);

	ProducerArgs producer_args[PRODUCER_COUNT];

	ConsumerArgs consumer_args = {
		.queue = &q,
		.consumed = &consumed,
		.failed = &failed,
		.seen = seen
	};

	for (int i = 0; i < CONSUMER_COUNT; i++) {
		cr_assert_eq(
			pthread_create(&consumers[i], NULL, consumer, &consumer_args),
			0
		);
	}

	for (int i = 0; i < PRODUCER_COUNT; i++) {
		producer_args[i] = (ProducerArgs) {
			.queue = &q,
			.start = i * ITEMS_PER_PRODUCER + 1,
			.count = ITEMS_PER_PRODUCER,
			.failed = &failed
		};

		cr_assert_eq(
			pthread_create(&producers[i], NULL, producer, &producer_args[i]),
			0
		);
	}

	for (int i = 0; i < PRODUCER_COUNT; i++)
		cr_assert_eq(pthread_join(producers[i], NULL), 0);

	for (int i = 0; i < CONSUMER_COUNT; i++)
		cr_assert_eq(pthread_join(consumers[i], NULL), 0);

	cr_assert_eq(atomic_load(&failed), 0);
	cr_assert_eq(atomic_load(&consumed), TOTAL_ITEMS);

	for (int i = 1; i <= TOTAL_ITEMS; i++)
		cr_assert_eq(atomic_load(&seen[i]), 1);

	cr_assert_eq(q.size, 0);

	cr_assert_eq(queue_destroy(&q), QUEUE_OK);
	free(seen);
}
