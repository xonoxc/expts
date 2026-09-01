package que

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"durqueue/job"
	"durqueue/store"
)

func newTestQueue(t *testing.T) *Queue {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return NewQueue(store.NewSqliteStoreWith(db))
}

func TestQueue_EnqueueDequeue(t *testing.T) {
	q := newTestQueue(t)
	ctx := context.Background()

	want := job.Job{ID: "job-1"}

	err := q.Enqueue(ctx, want)
	if err != nil {
		t.Fatal(err)
	}

	got, err := q.Dequeue(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestQueue_FIFO(t *testing.T) {
	q := newTestQueue(t)
	ctx := context.Background()

	jobs := []job.Job{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	}

	for _, job := range jobs {
		if err := q.Enqueue(ctx, job); err != nil {
			t.Fatal(err)
		}
	}

	for _, want := range jobs {
		got, err := q.Dequeue(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if got.ID != want.ID {
			t.Fatalf("got %s, want %s", got.ID, want.ID)
		}
	}
}

func TestQueue_DequeueEmpty_Blocks(t *testing.T) {
	q := newTestQueue(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		50*time.Millisecond,
	)
	defer cancel()

	_, err := q.Dequeue(ctx)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

func TestQueue_WaitingConsumerReceivesEnqueuedJob(t *testing.T) {
	q := newTestQueue(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan job.Job, 1)
	ready := make(chan struct{})

	go func() {
		close(ready)
		job, err := q.Dequeue(ctx)
		if err != nil {
			return
		}
		done <- job
	}()

	<-ready

	if err := q.Enqueue(ctx, job.Job{ID: "job-1"}); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-done:
		if got.ID != "job-1" {
			t.Fatalf("got %s, want job-1", got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("consumer never received job")
	}
}

func TestQueue_OneJobDeliveredToOneConsumer(t *testing.T) {
	q := newTestQueue(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := make(chan job.Job, 2)

	var started sync.WaitGroup
	started.Add(2)

	for range 2 {
		go func() {
			started.Done()
			job, err := q.Dequeue(ctx)
			if err != nil {
				return
			}

			results <- job
		}()
	}

	started.Wait()

	if err := q.Enqueue(ctx, job.Job{ID: "job-1"}); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-results:
		if got.ID != "job-1" {
			t.Fatalf("got %s, want job-1", got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("no consumer received job")
	}

	select {
	case got := <-results:
		t.Fatalf("job was delivered twice: %+v", got)
	case <-time.After(50 * time.Millisecond):
		// expected: second consumer remains blocked
	}

	cancel()
}
