package main

import (
	"context"
	"errors"
	"net"
	"sync"
)

var ErrPoolClosed = errors.New("pool is already closed")

type Pool struct {
	mu      sync.Mutex
	cond    *sync.Cond
	conns   []net.Conn
	closed  bool
	maxSize int
}

func NewPool(maxSize int) *Pool {
	pool := &Pool{
		conns:   make([]net.Conn, 0, maxSize),
		maxSize: maxSize,
		closed:  false,
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

	for len(p.conns) == 0 && !p.closed {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		p.cond.Wait()
	}

	if p.closed {
		return nil, ErrPoolClosed
	}

	conn := p.conns[0]
	p.conns = p.conns[1:]

	return conn, nil
}

func (p *Pool) Release(ctx context.Context, conn net.Conn) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	done := make(chan struct{})
	defer close(done)

	p.initContextWatcher(ctx, done)

	for len(p.conns) == p.maxSize && !p.closed {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		p.cond.Wait()
	}

	if p.closed {
		return ErrPoolClosed
	}

	p.conns = append(p.conns, conn)
	p.cond.Broadcast()

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
