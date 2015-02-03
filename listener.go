package lazyvm

import (
	"errors"
	"net"
	"sync/atomic"
)

// ErrInterrupted TODO(rjeczalik)
var ErrInterrupted = errors.New("interrupted")

type listener struct {
	net.Listener
	done int32
	err  chan error
	conn chan net.Conn
}

// InterruptListen TODO(rjeczalik)
func InterruptListen(network, addr string) (net.Listener, error) {
	lis, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	l := &listener{
		Listener: lis,
		err:      make(chan error, 1),
		conn:     make(chan net.Conn, 1),
	}
	go func() {
		for atomic.LoadInt32(&l.done) == 0 {
			conn, err := l.Listener.Accept()
			switch err {
			case nil:
				l.conn <- conn
			default:
				l.err <- err
			}
		}
	}()
	return l, nil
}

// Accept TODO(rjeczalik)
func (l *listener) Accept() (net.Conn, error) {
	select {
	case err := <-l.err:
		return nil, err
	case conn := <-l.conn:
		return conn, nil
	}
}

// Close TODO(rjeczalik)
func (l *listener) Close() error {
	atomic.StoreInt32(&l.done, 1)
	l.err <- ErrInterrupted
	return l.Listener.Close()
}
