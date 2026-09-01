package que

import (
	"context"
	"sync"
)

type Queue[T any] struct {
	mu       sync.Mutex
	enteries []T
	notify   chan struct{}
}

func NewQueue[T any]() *Queue[T] {
	enteries := make([]T, 0, 10)
	notiChan := make(chan struct{}, 1)

	return &Queue[T]{
		enteries: enteries,
		notify:   notiChan,
	}
}

func (q *Queue[T]) Enqueue(entry T) error {
	q.mu.Lock()
	q.enteries = append(q.enteries, entry)
	q.mu.Unlock()

	select {
	case q.notify <- struct{}{}:
	default:
		// do nothing is not able to push
	}

	return nil
}

func (q *Queue[T]) Dequeue(ctx context.Context) (T, error) {
	var zero T

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
