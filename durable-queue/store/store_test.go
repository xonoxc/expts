package store

import (
	"context"
	"errors"
	"sync"
	"testing"

	"durqueue/job"
)

func TestStore_InsertGet(t *testing.T) {
	store := NewSqliteStore[job.Job]()
	ctx := context.Background()

	want := job.Job{ID: "job-1"}

	if err := store.Insert(ctx, want); err != nil {
		t.Fatal(err)
	}

	got, err := store.Get(ctx, "job-1")
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestStore_GetMissing(t *testing.T) {
	store := NewSqliteStore[job.Job]()

	_, err := store.Get(context.Background(), "missing")

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStore_Update(t *testing.T) {
	store := NewSqliteStore[job.Job]()
	ctx := context.Background()

	if err := store.Insert(ctx, job.Job{ID: "1"}); err != nil {
		t.Fatal(err)
	}

	updated := job.Job{ID: "1"}

	if err := store.Update(ctx, updated); err != nil {
		t.Fatal(err)
	}

	got, err := store.Get(ctx, "1")
	if err != nil {
		t.Fatal(err)
	}

	if got != updated {
		t.Fatalf("got %+v, want %+v", got, updated)
	}
}

func TestStore_UpdateMissing(t *testing.T) {
	store := NewSqliteStore[job.Job]()

	err := store.Update(
		context.Background(),
		job.Job{ID: "missing"},
	)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStore_Delete(t *testing.T) {
	store := NewSqliteStore[job.Job]()
	ctx := context.Background()

	if err := store.Insert(ctx, job.Job{ID: "1"}); err != nil {
		t.Fatal(err)
	}

	if err := store.Delete(ctx, "1"); err != nil {
		t.Fatal(err)
	}

	_, err := store.Get(ctx, "1")
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestStore_DeleteMissing(t *testing.T) {
	store := NewSqliteStore[job.Job]()

	err := store.Delete(context.Background(), "missing")

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStore_InsertDuplicate(t *testing.T) {
	store := NewSqliteStore[job.Job]()
	ctx := context.Background()

	if err := store.Insert(ctx, job.Job{ID: "1"}); err != nil {
		t.Fatal(err)
	}

	err := store.Insert(ctx, job.Job{ID: "1"})

	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestStore_OperationsAreIndependent(t *testing.T) {
	store := NewSqliteStore[job.Job]()
	ctx := context.Background()

	jobs := []job.Job{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	}

	for _, job := range jobs {
		if err := store.Insert(ctx, job); err != nil {
			t.Fatal(err)
		}
	}

	if err := store.Update(ctx, job.Job{ID: "2"}); err != nil {
		t.Fatal(err)
	}

	if err := store.Delete(ctx, "2"); err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"1", "3"} {
		got, err := store.Get(ctx, id)
		if err != nil {
			t.Fatal(err)
		}

		if got.ID != id {
			t.Fatalf("got %+v, want ID %s", got, id)
		}
	}
}

func TestStore_ConcurrentAccess(t *testing.T) {
	store := NewSqliteStore[job.Job]()
	ctx := context.Background()

	const workers = 100

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()

			id := string(rune(i + 1000))
			job := job.Job{ID: id}

			if err := store.Insert(ctx, job); err != nil {
				return
			}

			_, _ = store.Get(ctx, id)
			_ = store.Update(ctx, job)
		}(i)
	}

	wg.Wait()
}

func TestStore_CancelledContext(t *testing.T) {
	store := NewSqliteStore[job.Job]()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := store.Insert(ctx, job.Job{ID: "1"})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}
