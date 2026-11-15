#ifndef QUEUE_H
#define QUEUE_H

#include <stddef.h>
#include <pthread.h>

typedef enum {
	QUEUE_OK,
	QUEUE_FULL,
	QUEUE_EMPTY,
	QUEUE_INVALID,
	QUEUE_SYSTEM_ERROR,
} QueueError;


typedef struct {
	int *buffer;
	size_t capacity;
	size_t size;
	size_t tail;
	size_t head;

	pthread_mutex_t mutex;

	pthread_cond_t not_empty;
	pthread_cond_t not_full;
} Queue;


QueueError queue_init(Queue *q, size_t capacity);
QueueError enqueue(Queue *q , int element);
QueueError dequeue(Queue *q , int *out);
QueueError queue_destroy(Queue *q);

#endif 
