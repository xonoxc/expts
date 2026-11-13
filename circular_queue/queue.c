#include "queue.h"
#include <stdio.h>
#include <stdlib.h>


QueueError queue_init(Queue *q, size_t capacity)
{
	if(!q ||  capacity == 0) 
		return QUEUE_INVALID;

	q->buffer = malloc(sizeof(int) * capacity);

	if (!q-> buffer)
		return QUEUE_SYSTEM_ERROR;


	if (pthread_mutex_init(&q->mutex , NULL)) {
		free(q-> buffer);
		q->buffer = NULL;
		return QUEUE_SYSTEM_ERROR;
	}

	q->capacity = capacity;
	q->head = 0;
	q->size = 0;
	q->tail = 0;

	return QUEUE_OK;
}

QueueError enqueue(Queue *q , int element) 
{
	if (!q) 
		return QUEUE_INVALID;

	int err = pthread_mutex_lock(&q->mutex);
	if(err) 
		return QUEUE_SYSTEM_ERROR;

	if (q->size == q-> capacity) {
		err = pthread_mutex_unlock(&q->mutex);
		if(err) 
			return QUEUE_SYSTEM_ERROR;

		return QUEUE_FULL;
	}

	q-> buffer[q-> tail]= element;

	q->tail = (q->tail+1) % q-> capacity;
	q->size++;


	err = pthread_mutex_unlock(&q->mutex);
	if(err) 
		return QUEUE_SYSTEM_ERROR;


	return QUEUE_OK;
}


QueueError dequeue(Queue *q , int *out) 
{
	if (!q || !out) 
		return QUEUE_INVALID;

	int err = pthread_mutex_lock(&q->mutex);
	if(err) 
		return QUEUE_SYSTEM_ERROR;


	if (q-> size < 1) {
		err = pthread_mutex_unlock(&q->mutex);
		if(err) 
			return QUEUE_SYSTEM_ERROR;

		return QUEUE_EMPTY;
	}

	*out = (q->buffer[q->head]);

	q-> head = (q-> head + 1) % q-> capacity;
	q->size--;


	err = pthread_mutex_unlock(&q->mutex);
	if(err) 
		return QUEUE_SYSTEM_ERROR;

	return QUEUE_OK;
}



QueueError queue_destroy(Queue *q) 
{
	if (!q || !q-> buffer) 
		return QUEUE_INVALID;


	int err = pthread_mutex_destroy(&q->mutex);
	if (err) 
		return QUEUE_SYSTEM_ERROR;


	free(q->buffer);
	q->buffer = NULL;

	q->capacity = 0;
	q->size = 0;
	q->tail= 0;
	q->head = 0;

	return QUEUE_OK;
}
