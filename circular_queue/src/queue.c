#include "queue.h"
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>


QueueError queue_init(Queue *q, size_t capacity)
{
	if(!q ||  capacity == 0) 
		return QUEUE_INVALID;

	if (capacity > SIZE_MAX / sizeof(int))
		return QUEUE_INVALID;

	q->buffer = malloc(sizeof(int) * capacity);

	if (!q-> buffer)
		return QUEUE_SYSTEM_ERROR;


	if (pthread_mutex_init(&q->mutex , NULL)) {
		free(q-> buffer);
		q->buffer = NULL;
		return QUEUE_SYSTEM_ERROR;
	}

	if (pthread_cond_init(&q->not_empty , NULL)) {
		int err 
			= pthread_mutex_destroy(&q->mutex);
		if (err)  
			return QUEUE_SYSTEM_ERROR;

		free(q-> buffer);
		q->buffer = NULL;
		return QUEUE_SYSTEM_ERROR;
	}

	if (pthread_cond_init(&q->not_full , NULL)) {
		int err
			= pthread_cond_destroy(&q->not_empty);
		if (err) 
			return QUEUE_SYSTEM_ERROR;

		err = pthread_mutex_destroy(&q->mutex);
		if (err) 
			return QUEUE_SYSTEM_ERROR;

		free(q-> buffer);
		q->buffer = NULL;
		return QUEUE_SYSTEM_ERROR;
	}

	q->capacity = capacity;
	q->head = 0;
	q->size = 0;
	q->tail = 0;
	q->shutdown = false;


	return QUEUE_OK;

}

QueueError enqueue(Queue *q , int element) 
{
	if (!q) 
		return QUEUE_INVALID;

	int err = pthread_mutex_lock(&q->mutex);
	if(err) 
		return QUEUE_SYSTEM_ERROR;


	while (q->size == q-> capacity && !q->shutdown) {
		err = pthread_cond_wait(&q->not_full, &q->mutex);
		if (err) {
			pthread_mutex_unlock(&q->mutex);
			return QUEUE_SYSTEM_ERROR;
		}
	}

	if(q->shutdown) {
		err = pthread_mutex_unlock(&q->mutex);
		if(err) 
			return QUEUE_SYSTEM_ERROR;

		return QUEUE_SHUTDOWN;
	}

	q-> buffer[q-> tail]= element;

	q->tail = (q->tail+1) % q-> capacity;
	q->size++;

	err = pthread_cond_signal(&q->not_empty);
	if (err) {
		pthread_mutex_unlock(&q->mutex);
		return QUEUE_SYSTEM_ERROR;
	} 

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


	while (q-> size == 0 && !q->shutdown) {
		err = pthread_cond_wait(&q->not_empty, &q->mutex);
		if (err) {
			pthread_mutex_unlock(&q->mutex);
			return QUEUE_SYSTEM_ERROR;
		}
	}


	if(q->shutdown && q->size == 0) {
		err = pthread_mutex_unlock(&q->mutex);
		if(err) 
			return QUEUE_SYSTEM_ERROR;

		return QUEUE_SHUTDOWN;
	}

	*out = (q->buffer[q->head]);

	q-> head = (q-> head + 1) % q-> capacity;
	q->size--;

	err = pthread_cond_signal(&q->not_full);
	if (err) {
		pthread_mutex_unlock(&q->mutex);
		return QUEUE_SYSTEM_ERROR;
	} 

	err = pthread_mutex_unlock(&q->mutex);
	if(err) 
		return QUEUE_SYSTEM_ERROR;

	return QUEUE_OK;
}

QueueError queue_destroy(Queue *q) 
{
	if (!q || !q-> buffer) 
		return QUEUE_INVALID;


	int err = pthread_cond_destroy(&q->not_full);
	if (err) 
		return QUEUE_SYSTEM_ERROR;

	err = pthread_cond_destroy(&q->not_empty);
	if (err) 
		return QUEUE_SYSTEM_ERROR;

	err = pthread_mutex_destroy(&q->mutex);
	if (err) 
		return QUEUE_SYSTEM_ERROR;

	free(q->buffer);
	q->buffer = NULL;

	q->capacity = 0;
	q->size = 0;
	q->tail= 0;
	q->head = 0;
	q->shutdown = false;

	return QUEUE_OK;
}

QueueError queue_shutdown(Queue *q)
{
	if (!q || !q-> buffer)
		return QUEUE_INVALID;

	int err = pthread_mutex_lock(&q->mutex);
	if (err)
		return QUEUE_SYSTEM_ERROR;

	q->shutdown = true;

	err = pthread_cond_broadcast(&q->not_empty);
	if (err) {
		pthread_mutex_unlock(&q->mutex);
		return QUEUE_SYSTEM_ERROR;
	}

	err = pthread_cond_broadcast(&q->not_full);
	if (err) {
		pthread_mutex_unlock(&q->mutex);
		return QUEUE_SYSTEM_ERROR;
	}

	err = pthread_mutex_unlock(&q->mutex);
	if (err)
		return QUEUE_SYSTEM_ERROR;

	return QUEUE_OK;
}
