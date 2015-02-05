// Package lazyvm provides a proxy for RDP connections, which is capable of
// starting / stopping on idle Windows instances through VirtualBox manager.
//
// Actually it treats underlying connection as a blackbox, so it can be used
// with different protocols / OSes. It was just developed for use with rdesktop
// and Windows.
package lazyvm

import (
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func nonil(err ...error) error {
	for _, err := range err {
		if err != nil {
			return err
		}
	}
	return nil
}

// Proxy TODO(rjeczalik)
type Proxy struct {
	MachineName string // VirtualBox machine name
	Addr        string // network address to listen on, default is ":5000"
	Port        int    // target port of the machine, default is 3389

	lis     net.Listener
	liswg   sync.WaitGroup
	errch   chan error
	vbox    *VirtualBox
	vboxwg  sync.WaitGroup
	running int32
	busy    uint64
}

func (p *Proxy) addr() string {
	if p.Addr != "" {
		return p.Addr
	}
	return ":5000"
}

func (p *Proxy) port() int {
	if p.Port != 0 {
		return p.Port
	}
	return 3389
}

func (p *Proxy) err() chan error {
	if p.errch == nil {
		p.errch = make(chan error, 1)
	}
	return p.errch
}

func (p *Proxy) eventloop() {
	done := make(chan struct{})
	go func() {
		p.liswg.Wait()
		close(done)
	}()
	for atomic.LoadInt32(&p.running) == 1 {
		select {
		case <-done:
			if err := p.vbox.Close(); err != nil {
				if ok, e := p.vbox.Running(); e != nil || ok {
					log.Println("vbox stop error:", err)
					continue
				}
			}
			atomic.StoreInt32(&p.running, 0)
			log.Println("vbox stopped:", p.MachineName)
			return
		default:
			time.Sleep(time.Second)
		}
	}
}

// Stop TODO(rjeczalik): discard vbox saved state
func (p *Proxy) Stop() {
	p.err() <- p.lis.Close()
}

func (p *Proxy) drop(c net.Conn, v ...interface{}) {
	c.Close()
	p.liswg.Done()
	log.Println(v...)
}

// Run TODO(rjeczalik)
func (p *Proxy) Run() error {
	vbox, err := NewVirtualBox(p.MachineName)
	if err != nil {
		return err
	}
	l, err := InterruptibleListen("tcp", p.addr())
	if err != nil {
		return err
	}
	p.vbox = vbox
	p.lis = l
	log.Printf("proxy listening on %s . . .", p.lis.Addr())
AcceptLoop:
	for {
		src, err := p.lis.Accept()
		switch err {
		case nil:
		case ErrInterrupted:
			break AcceptLoop
		default:
			log.Println("accept error:", err)
			continue AcceptLoop
		}
		p.liswg.Add(1) // account connection
		switch ok, err := p.vbox.Running(); {
		case err != nil:
			p.drop(src, "vbox error:", err)
			continue AcceptLoop
		case !ok:
			if err := p.vbox.Start(); err != nil {
				p.drop(src, "vbox start error:", err)
				continue AcceptLoop
			}
			log.Println("vbox started:", p.MachineName)
			if atomic.CompareAndSwapInt32(&p.running, 0, 1) {
				go p.eventloop()
			}
		}
		addr, err := vbox.Addr()
		if err != nil {
			p.drop(src, "addr error:", err)
			continue
		}
		addr = addr + ":" + strconv.Itoa(p.port())
		dst, err := net.Dial("tcp", addr)
		if err != nil {
			p.drop(src, "dial error:", err)
			continue
		}
		log.Printf("proxying %v -> %v . . .", src.RemoteAddr(), addr)
		go p.serve(BusyConn(src, &p.vboxwg, &p.busy), dst)
	}
	return <-p.err()
}

func (p *Proxy) serve(src, dst net.Conn) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		io.Copy(src, dst)
		log.Println("[dbg] src -> dst done")
		wg.Done()
	}()
	go func() {
		io.Copy(dst, src)
		log.Println("[dbg] dst -> src done")
		wg.Done()
	}()
	wg.Wait()
	dst.Close()
	src.Close()
	p.liswg.Done()
}
