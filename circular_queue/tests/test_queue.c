#include <errno.h>
#include <stdbool.h>
#include <time.h>
#include <criterion/criterion.h>
#include "queue.h"
#include <pthread.h>
#include <stdatomic.h>
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
	cr_assert_not(q.shutdown);

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
		QueueError err = enqueue(args->queue, value);
		if(err == QUEUE_SHUTDOWN)
			break;

		if (err != QUEUE_OK) {
			atomic_store(args->failed, 1);
			return NULL;
		}
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
		if(err == QUEUE_SHUTDOWN)
			break;


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

	for (int i = 0; i < PRODUCER_COUNT; i++) {
		cr_assert_eq(pthread_join(producers[i], NULL), 0);
	}

	cr_assert_eq(queue_shutdown(&q), QUEUE_OK);

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



typedef struct {
	Queue *queue;

	pthread_mutex_t mutex;
	pthread_cond_t cond;

	bool completed;
	int val;

	QueueError error;
}BlockingConsumerArgs;


static void *blocking_consumer(void *arg)
{
	BlockingConsumerArgs *args = arg;

	args-> error = dequeue(args->queue, &args->val);


	pthread_mutex_lock(&args->mutex);
	args->completed = true;
	pthread_cond_signal(&args->cond);
	pthread_mutex_unlock(&args->mutex);

	return NULL;
}

typedef struct {
	Queue *queue;

	pthread_mutex_t mutex;
	pthread_cond_t cond;

	bool completed;
	QueueError error;
}BlockingProducerArgs;


static void *blocking_producer(void *arg)
{
	BlockingProducerArgs *args  = arg;

	int element = 43;

	args->error = enqueue(args->queue , element);

	pthread_mutex_lock(&args->mutex);
	args->completed = true;
	pthread_cond_signal(&args->cond);
	pthread_mutex_unlock(&args->mutex);

	return NULL;
}


static int wait_for_producer_completion(BlockingProducerArgs *args)
{
	struct timespec deadline;

	clock_gettime(CLOCK_REALTIME , &deadline);
	deadline.tv_sec += 2;

	pthread_mutex_lock(&args->mutex);

	int rc = 0;

	while (!args->completed  && rc == 0) {
		rc = pthread_cond_timedwait(
			&args->cond,
			&args->mutex,
			&deadline
		);
	}

	pthread_mutex_unlock(&args->mutex);

	return rc;
}


static int wait_for_completion(BlockingConsumerArgs *args)
{
	struct timespec deadline;

	clock_gettime(CLOCK_REALTIME , &deadline);
	deadline.tv_sec += 2;

	pthread_mutex_lock(&args->mutex);

	int rc = 0;

	while (!args->completed  && rc == 0) {
		rc = pthread_cond_timedwait(
			&args->cond,
			&args->mutex,
			&deadline
		);
	}

	pthread_mutex_unlock(&args->mutex);

	return rc;
}


Test(queue, consumer_blocks_when_empty)
{
	Queue q;
	pthread_t thread;

	BlockingConsumerArgs args = {
		.queue = &q,
		.completed = false
	};

	cr_assert_eq(queue_init(&q, 1), QUEUE_OK);

	pthread_mutex_init(&args.mutex, NULL);
	pthread_cond_init(&args.cond, NULL);

	cr_assert_eq(
		pthread_create(&thread, NULL, blocking_consumer, &args),
		0
	);

	cr_assert_eq(wait_for_completion(&args), ETIMEDOUT);

	cr_assert_eq(enqueue(&q, 42), QUEUE_OK);

	cr_assert_eq(wait_for_completion(&args), 0);

	cr_assert_eq(pthread_join(thread, NULL), 0);

	cr_assert_eq(args.error, QUEUE_OK);
	cr_assert_eq(args.val, 42);

	pthread_cond_destroy(&args.cond);
	pthread_mutex_destroy(&args.mutex);
	cr_assert_eq(queue_destroy(&q), QUEUE_OK);
}

Test(queue, producer_blocks_when_full)
{
	Queue q;
	pthread_t thread;

	BlockingProducerArgs args = {
		.queue = &q,
		.completed = false
	};


	cr_assert_eq(queue_init(&q, 1), QUEUE_OK);


	pthread_mutex_init(&args.mutex, NULL);
	pthread_cond_init(&args.cond, NULL);


	cr_assert_eq(
		pthread_create(
			&thread, 
			NULL, 
			blocking_producer, 
			&args
		),
		0
	);
	cr_assert_eq(enqueue(&q, 22) , QUEUE_OK);


	cr_assert_eq(wait_for_producer_completion(&args) , ETIMEDOUT);

	int out;
	cr_assert_eq(dequeue(&q , &out), QUEUE_OK);

	cr_assert_eq(wait_for_producer_completion(&args) , 0);

	cr_assert_eq(args.error, QUEUE_OK);

	cr_assert_eq(pthread_join(thread, NULL), 0);

	pthread_cond_destroy(&args.cond);
	pthread_mutex_destroy(&args.mutex);
	cr_assert_eq(queue_destroy(&q), QUEUE_OK);
}


Test(queue, shutdown_wakes_blocked_consumer)
{
	Queue q;
	pthread_t thread;

	BlockingConsumerArgs args = {
		.queue = &q,
		.completed = false
	};

	cr_assert_eq(queue_init(&q, 1), QUEUE_OK);

	pthread_mutex_init(&args.mutex, NULL);
	pthread_cond_init(&args.cond, NULL);

	cr_assert_eq(
		pthread_create(&thread, NULL, blocking_consumer, &args),
		0
	);

	cr_assert_eq(wait_for_completion(&args), ETIMEDOUT);

	cr_assert_eq(queue_shutdown(&q), QUEUE_OK);

	cr_assert_eq(wait_for_completion(&args), 0);

	cr_assert_eq(pthread_join(thread, NULL), 0);

	cr_assert_eq(args.error, QUEUE_SHUTDOWN);

	pthread_cond_destroy(&args.cond);
	pthread_mutex_destroy(&args.mutex);
	cr_assert_eq(queue_destroy(&q), QUEUE_OK);
}


Test(queue, shutdown_wakes_blocked_producer)
{
	Queue q;
	pthread_t thread;

	BlockingProducerArgs args = {
		.queue = &q,
		.completed = false
	};

	cr_assert_eq(queue_init(&q, 1), QUEUE_OK);

	pthread_mutex_init(&args.mutex, NULL);
	pthread_cond_init(&args.cond, NULL);

	cr_assert_eq(enqueue(&q, 22), QUEUE_OK);

	cr_assert_eq(
		pthread_create(&thread, NULL, blocking_producer, &args),
		0
	);

	cr_assert_eq(wait_for_producer_completion(&args), ETIMEDOUT);

	cr_assert_eq(queue_shutdown(&q), QUEUE_OK);

	cr_assert_eq(wait_for_producer_completion(&args), 0);

	cr_assert_eq(pthread_join(thread, NULL), 0);

	cr_assert_eq(args.error, QUEUE_SHUTDOWN);

	pthread_cond_destroy(&args.cond);
	pthread_mutex_destroy(&args.mutex);
	cr_assert_eq(queue_destroy(&q), QUEUE_OK);
}


Test(queue, shutdown_drains_remaining_items)
{
	Queue q;
	int value;

	cr_assert_eq(queue_init(&q, 4), QUEUE_OK);

	cr_assert_eq(enqueue(&q, 10), QUEUE_OK);
	cr_assert_eq(enqueue(&q, 20), QUEUE_OK);

	cr_assert_eq(queue_shutdown(&q), QUEUE_OK);

	cr_assert_eq(dequeue(&q, &value), QUEUE_OK);
	cr_assert_eq(value, 10);

	cr_assert_eq(dequeue(&q, &value), QUEUE_OK);
	cr_assert_eq(value, 20);

	cr_assert_eq(dequeue(&q, &value), QUEUE_SHUTDOWN);

	cr_assert_eq(queue_destroy(&q), QUEUE_OK);
}


Test(queue, enqueue_after_shutdown_is_rejected)
{
	Queue q;

	cr_assert_eq(queue_init(&q, 2), QUEUE_OK);
	cr_assert_eq(queue_shutdown(&q), QUEUE_OK);

	cr_assert_eq(enqueue(&q, 10), QUEUE_SHUTDOWN);
	cr_assert_eq(q.size, 0);

	cr_assert_eq(queue_destroy(&q), QUEUE_OK);
}


Test(queue, shutdown_is_idempotent_and_rejects_destroyed_queue)
{
	Queue q;

	cr_assert_eq(queue_shutdown(NULL), QUEUE_INVALID);

	cr_assert_eq(queue_init(&q, 2), QUEUE_OK);
	cr_assert_eq(queue_shutdown(&q), QUEUE_OK);
	cr_assert_eq(queue_shutdown(&q), QUEUE_OK);
	cr_assert_eq(queue_destroy(&q), QUEUE_OK);

	cr_assert_eq(queue_shutdown(&q), QUEUE_INVALID);
}
