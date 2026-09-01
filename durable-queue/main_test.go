package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestQueue_EnqueueDequeue(t *testing.T) {
	q := NewQueue[Job]()
	ctx := context.Background()

	want := Job{ID: "job-1"}

	err := q.Enqueue(want)
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
	q := NewQueue[Job]()
	ctx := context.Background()

	jobs := []Job{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	}

	for _, job := range jobs {
		if err := q.Enqueue(job); err != nil {
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
	q := NewQueue[Job]()

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
