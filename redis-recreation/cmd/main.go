package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/xonoxc/expts/redis-recreation/internal/server"
	"github.com/xonoxc/expts/redis-recreation/internal/store"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	var wg sync.WaitGroup

	store := store.New()
	store.StartStoreGC(ctx, &wg)

	svr := server.NewServer(
		server.SERVER_DEFAULT_PORT, store,
	)

	log.Printf("server started at port %s", server.SERVER_DEFAULT_PORT)

	err := svr.Start(ctx, &wg)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("graceful shutdown complete")
	case <-time.After(30 * time.Second):

		log.Println("shutdown timed out")
	}
}
