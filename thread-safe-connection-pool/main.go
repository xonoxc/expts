package main

import (
	"context"
	"errors"
	"net"
	"sync"
)

var ErrPoolClosed = errors.New("pool is already closed")

type Pool struct {
	mu       sync.Mutex
	cond     *sync.Cond
	conns    []net.Conn
	closed   bool
	currSize int
	maxSize  int
	factory  func() (net.Conn, error)
}

func NewPool(maxSize int, fac func() (net.Conn, error)) *Pool {
	pool := &Pool{
		conns:    make([]net.Conn, 0, maxSize),
		maxSize:  maxSize,
		closed:   false,
		currSize: 0,
		factory:  fac,
	}

	pool.cond = sync.NewCond(&pool.mu)

	return pool
}

func (p *Pool) Acquire(ctx context.Context) (net.Conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	done := make(chan struct{})
	defer close(done)

	p.initContextWatcher(ctx, done)

	for !p.closed {
		if len(p.conns) > 0 {
			conn := p.conns[0]
			p.conns = p.conns[1:]

			return conn, nil
		}

		if p.currSize < p.maxSize {
			p.currSize++
			p.mu.Unlock()

			conn, err := p.factory()

			p.mu.Lock()

			if err != nil {
				p.currSize--
				p.cond.Signal()
				return nil, err
			}

			if p.closed {
				p.currSize--
				conn.Close()
				return nil, ErrPoolClosed
			}

			return conn, nil
		}

		if err := ctx.Err(); err != nil {
			return nil, err
		}

		p.cond.Wait()

	}

	return nil, ErrPoolClosed
}

func (p *Pool) Release(ctx context.Context, conn net.Conn) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	if p.closed {
		conn.Close()
		return ErrPoolClosed
	}

	p.conns = append(p.conns, conn)
	p.cond.Signal()

	return nil
}

func (p *Pool) initContextWatcher(ctx context.Context, done <-chan struct{}) {
	go func() {
		select {
		case <-ctx.Done():
			p.mu.Lock()
			p.cond.Broadcast()
			p.mu.Unlock()

		case <-done:
		}
	}()
}

func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	defer p.cond.Broadcast()

	p.closed = true

	if len(p.conns) == 0 {
		return
	}

	for _, conn := range p.conns {
		conn.Close()
	}
}
