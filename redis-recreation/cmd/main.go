package main

import (
	"context"
	"log"

	"github.com/xonoxc/expts/redis-recreation/internal/server"
	"github.com/xonoxc/expts/redis-recreation/internal/store"
)

func main() {
	store := store.New()
	svr := server.NewServer(
		server.SERVER_DEFAULT_PORT, store,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := svr.Start(ctx); err != nil {
		log.Println("startup error:", err)
	}

	log.Printf("server started at port %s", server.SERVER_DEFAULT_PORT)
}
