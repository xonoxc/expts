package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/xonoxc/expts/redis-recreation/internal/command"
	"github.com/xonoxc/expts/redis-recreation/internal/resp"
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

func (s *Server) handleConn(_ context.Context, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			conn.Write([]byte("\nErr:failed reading btyes\n"))
			continue
		}

		parser := resp.NewParser()
		parsedBytes, err := parser.Parse(buf[:n])
		if err != nil {
			switch true {
			case errors.Is(err, resp.ErrEmptyBuffer):
				conn.Write([]byte("\nErr:reading empty btyes\n"))
				continue

			case errors.Is(err, resp.ErrMalFormedBytes):
				conn.Write([]byte("\nErr:invalid payload\n"))
				continue

			case errors.Is(err, resp.ErrIncomplete):
				conn.Write([]byte("\nErr:incomplete payload\n"))
				continue

			default:
				conn.Write([]byte("\nErr:unexpected unkown error while reading payload\n"))
				continue
			}
		}

		cmd := command.NewCommand(parsedBytes[0], parsedBytes[1:])
		dsptr := command.NewDispatcher(s.store)

		responseBytes, err := dsptr.Dispatch(cmd)
		if err != nil {
			switch true {
			case errors.Is(err, command.ErrInvalidSyntax):
				conn.Write([]byte("\nErr:invalid syntax\n"))
				continue

			case errors.Is(err, command.ErrrUnkownCommand):
				conn.Write([]byte("\nErr:invalid unkwon command\n"))
				continue

			case errors.Is(err, command.ErrInvalidExpirationTime):
				conn.Write([]byte("\nErr:invalid expiration time\n"))
				continue

			default:
				conn.Write([]byte("\nErr:unexpected unkown error while reading payload\n"))
				continue
			}
		}

		conn.Write(responseBytes)
		continue
	}
}
