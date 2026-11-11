#include <criterion/criterion.h>
#include "queue.h"

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
