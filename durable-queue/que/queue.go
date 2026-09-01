package que

import (
	"context"
	"sync"

	"durqueue/job"
	"durqueue/store"
)

type Queue struct {
	mu       sync.Mutex
	enteries []job.Job
	notify   chan struct{}
	store    store.Store
}

func NewQueue(str store.Store) *Queue {
	enteries := make([]job.Job, 0, 10)
	notiChan := make(chan struct{}, 1)

	return &Queue{
		enteries: enteries,
		notify:   notiChan,
		store:    str,
	}
}

func (q *Queue) Enqueue(ctx context.Context, entry job.Job) error {
	if err := q.store.Insert(ctx, entry); err != nil {
		return err
	}

	q.append(entry)

	select {
	case q.notify <- struct{}{}:
	default:
		// do nothing is not able to push
	}

	return nil
}

func (q *Queue) append(entry job.Job) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.enteries = append(q.enteries, entry)
}

func (q *Queue) Dequeue(ctx context.Context) (job.Job, error) {
	var zero job.Job

	for {
		q.mu.Lock()

		if len(q.enteries) > 0 {
			ent := q.enteries[0]
			q.enteries = q.enteries[1:]
			q.mu.Unlock()

			return ent, nil
		}

		q.mu.Unlock()

		select {
		case <-q.notify:
			// check if there is notification

		case <-ctx.Done():
			// context cancelled timed out
			return zero, ctx.Err()

		}
	}
}
