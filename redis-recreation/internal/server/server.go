package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/xonoxc/expts/redis-recreation/internal/store"
)

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

func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.lnAddr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	s.ln = ln

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	s.connLoop(ctx)

	return nil
}

func (s *Server) connLoop(ctx context.Context) {
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

			go s.handleConn(ctx, conn)

		}
	}
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
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
