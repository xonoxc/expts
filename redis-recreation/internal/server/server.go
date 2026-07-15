package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/xonoxc/expts/redis-recreation/internal/store"
)

const SERVER_DEFAULT_PORT = ":6379"

type Server struct {
	lnAddr string
	ln     net.Listener
	store  *store.Store
}

func NewServer(lnAddr string, store *store.Store) *Server {
	return &Server{
		lnAddr: lnAddr,
		ln:     nil,
		store:  store,
	}
}

func (s *Server) Start(ctx context.Context, wg *sync.WaitGroup) error {
	ln, err := net.Listen("tcp", s.lnAddr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	s.ln = ln

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	s.connLoop(ctx, wg)

	return nil
}

func (s *Server) connLoop(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			conn, err := s.ln.Accept()
			if err != nil {
				continue
			}

			log.Println("got new connection => IP:", conn.RemoteAddr())

			wg.Add(1)
			go s.handleConn(ctx, conn, wg)

		}
	}
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	// start a  in  a seperate goroutine loop here that will accept command and write responses

	// there will be a fixed size buffer here
	// in whici ill read messages into
	for {
		// n , err := conn.Read(&buf)
		// something like that

		// if err write a response to the client

		// byte response to the clinet here
	}
}
