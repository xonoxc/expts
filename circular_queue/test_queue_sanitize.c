#include "queue.h"

#include <pthread.h>
#include <stdatomic.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>

#define PRODUCERS 4
#define CONSUMERS 4
#define ITEMS_PER_PRODUCER 10000
#define TOTAL_ITEMS (PRODUCERS * ITEMS_PER_PRODUCER)
#define CAPACITY 64

static Queue queue;
static atomic_int consumed;
static atomic_int failed;
static atomic_int *seen;

void *producer(void *arg)
{
    int id = *(int *)arg;
    int start = id * ITEMS_PER_PRODUCER + 1;

    for (int i = start; i < start + ITEMS_PER_PRODUCER; i++) {
        QueueError err;

        while ((err = enqueue(&queue, i)) == QUEUE_FULL)
            sched_yield();

        if (err != QUEUE_OK) {
            atomic_store(&failed, 1);
            return NULL;
        }
    }

    return NULL;
}

void *consumer(void *arg)
{
    (void)arg;

    while (atomic_load(&consumed) < TOTAL_ITEMS) {
        int value;
        QueueError err = dequeue(&queue, &value);

        if (err == QUEUE_EMPTY) {
            sched_yield();
            continue;
        }

        if (err != QUEUE_OK) {
            atomic_store(&failed, 1);
            return NULL;
        }

        if (value < 1 || value > TOTAL_ITEMS) {
            atomic_store(&failed, 1);
            continue;
        }

        atomic_fetch_add(&seen[value], 1);
        atomic_fetch_add(&consumed, 1);
    }

    return NULL;
}

int main(void)
{
    pthread_t producers[PRODUCERS];
    pthread_t consumers[CONSUMERS];
    int producer_ids[PRODUCERS];

    atomic_init(&consumed, 0);
    atomic_init(&failed, 0);

    seen = calloc(TOTAL_ITEMS + 1, sizeof(*seen));
    if (!seen)
        return EXIT_FAILURE;

    for (int i = 0; i <= TOTAL_ITEMS; i++)
        atomic_init(&seen[i], 0);

    if (queue_init(&queue, CAPACITY) != QUEUE_OK)
        return EXIT_FAILURE;

    for (int i = 0; i < CONSUMERS; i++) {
        if (pthread_create(&consumers[i], NULL, consumer, NULL) != 0)
            return EXIT_FAILURE;
    }

    for (int i = 0; i < PRODUCERS; i++) {
        producer_ids[i] = i;

        if (pthread_create(&producers[i], NULL, producer, &producer_ids[i]) != 0)
            return EXIT_FAILURE;
    }

    for (int i = 0; i < PRODUCERS; i++)
        pthread_join(producers[i], NULL);

    for (int i = 0; i < CONSUMERS; i++)
        pthread_join(consumers[i], NULL);

    int errors = 0;

    for (int i = 1; i <= TOTAL_ITEMS; i++) {
        if (atomic_load(&seen[i]) != 1)
            errors++;
    }

    printf("Consumed: %d/%d\n", atomic_load(&consumed), TOTAL_ITEMS);
    printf("Missing/duplicate: %d\n", errors);
    printf("Thread errors: %d\n", atomic_load(&failed));

    queue_destroy(&queue);
    free(seen);

    return errors || atomic_load(&failed)
        ? EXIT_FAILURE
        : EXIT_SUCCESS;
}
