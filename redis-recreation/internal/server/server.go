package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	cmd "github.com/xonoxc/expts/redis-recreation/internal/command"
	"github.com/xonoxc/expts/redis-recreation/internal/resp"
	"github.com/xonoxc/expts/redis-recreation/internal/store"
)

const SERVER_DEFAULT_PORT = ":6379"

type Server struct {
	lnAddr string
	ln     net.Listener
	store  *store.Store
	conns  map[net.Conn]struct{}
	mu     sync.Mutex
}

func NewServer(lnAddr string, store *store.Store) *Server {
	return &Server{
		lnAddr: lnAddr,
		ln:     nil,
		store:  store,
		conns:  make(map[net.Conn]struct{}),
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

	wg.Go(
		func() {
			s.connLoop(wg)
		},
	)

	return nil
}

func (s *Server) connLoop(wg *sync.WaitGroup) {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			log.Printf("failed to accept connection: %v\n", err)
			continue
		}
		log.Println("got new connection => IP:", conn.RemoteAddr())

		s.RegisterConnection(conn)
		wg.Go(func() {
			s.handleConn(conn)
		})

	}
}

func (s *Server) RegisterConnection(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns[conn] = struct{}{}
}

func (s *Server) UnregisterConnection(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.conns, conn)
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	defer s.UnregisterConnection(conn)

	parser := resp.NewParser()
	dsptr := cmd.NewDispatcher(s.store)

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}

			if errors.Is(err, io.EOF) {
				log.Printf("connection closed by client: %s\n", conn.RemoteAddr())
				return
			}

			log.Printf("network error ocurred: %v\n", err)
			return
		}
		log.Printf("received %d bytes from %s\n", n, conn.RemoteAddr())

		parsedBytes, err := parser.Parse(buf[:n])
		if err != nil {
			switch {
			case errors.Is(err, resp.ErrEmptyBuffer):
				rsp := resp.SerializeError("reading empty bytes")
				if err := write(conn, rsp); err != nil {
					return
				}

			case errors.Is(err, resp.ErrMalFormedBytes):
				rsp := resp.SerializeError("invalid payload")
				if err := write(conn, rsp); err != nil {
					return
				}

			case errors.Is(err, resp.ErrIncomplete):
				rsp := resp.SerializeError("incomplete payload")
				if err := write(conn, rsp); err != nil {
					return
				}

			default:
				rsp := resp.SerializeError("unexpected error while reading payload")
				if err := write(conn, rsp); err != nil {
					return
				}
			}
			continue
		}

		comm := cmd.NewCommand(
			parsedBytes[0],
			parsedBytes[1:],
		)

		responseBytes, err := dsptr.Dispatch(comm)
		if err != nil {
			switch {
			case errors.Is(err, cmd.ErrInvalidSyntax):
				rsp := resp.SerializeError("invalid syntax")
				if err := write(conn, rsp); err != nil {
					return
				}

			case errors.Is(err, cmd.ErrrUnkownCommand):
				rsp := resp.SerializeError("invalid unkwon command")
				if err := write(conn, rsp); err != nil {
					return
				}

			case errors.Is(err, cmd.ErrInvalidExpirationTime):
				rsp := resp.SerializeError("invalid expiration time")
				if err := write(conn, rsp); err != nil {
					return
				}

			default:
				rsp := resp.SerializeError("unexpected unkown error while reading payload")
				if err := write(conn, rsp); err != nil {
					return
				}
			}
			continue
		}

		if err := write(conn, responseBytes); err != nil {
			return
		}
	}
}

func (s *Server) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for conn := range s.conns {
		conn.Close()
	}
}

func write(conn net.Conn, data []byte) error {
	_, err := conn.Write(data)
	if err != nil {
		log.Printf("write failed: %v", err)
	}

	return err
}
