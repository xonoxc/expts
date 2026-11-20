#ifndef QUEUE_H
#define QUEUE_H

#include <stddef.h>
#include <pthread.h>
#include <stdbool.h>


typedef enum {
	QUEUE_OK,
	QUEUE_INVALID,
	QUEUE_SYSTEM_ERROR,
	QUEUE_SHUTDOWN,
} QueueError;


typedef struct {
	int *buffer;
	size_t capacity;
	size_t size;
	size_t tail;
	size_t head;
	bool shutdown;

	pthread_mutex_t mutex;

	pthread_cond_t not_empty;
	pthread_cond_t not_full;
} Queue;


QueueError queue_init(Queue *q, size_t capacity);
QueueError enqueue(Queue *q , int element);
QueueError dequeue(Queue *q , int *out);

/*
 * Destroy the queue and release its resources.
 * The caller must ensure no thread is blocked in enqueue()/dequeue():
 * call queue_shutdown() and join those threads first.
 */
QueueError queue_destroy(Queue *q);

/*
 * Wake every blocked producer and consumer. enqueue() then fails with
 * QUEUE_SHUTDOWN; dequeue() drains any remaining items and then fails
 * with QUEUE_SHUTDOWN. Calling it more than once is safe.
 */
QueueError queue_shutdown(Queue *q);

#endif 
