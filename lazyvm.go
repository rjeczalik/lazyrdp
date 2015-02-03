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
	Port        int    // RDP port of the machine's server, default is 3389

	listener net.Listener
	errch    chan error
	vbox     *VirtualBox
	done     int32
	cnt      uint64
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

func (p *Proxy) loop() {

}

// Run TODO(rjeczalik)
func (p *Proxy) Run() error {
	vbox, err := NewVirtualBox(p.MachineName)
	if err != nil {
		return err
	}
	addr, err := vbox.Addr()
	if err != nil {
		return err
	}
	l, err := InterruptListen("tcp", p.addr())
	if err != nil {
		return err
	}
	p.vbox = vbox
	p.listener = l
	addr = addr + ":" + strconv.Itoa(p.port())
	log.Printf("proxy listening on %s, target is %s . . .", p.listener.Addr(), addr)
AcceptLoop:
	for {
		src, err := p.listener.Accept()
		switch err {
		case nil:
		case ErrInterrupted:
			break AcceptLoop
		default:
			log.Print("accept error:", err)
			continue AcceptLoop
		}
		dst, err := net.Dial("tcp", addr)
		if err != nil {
			src.Close()
			log.Print("dial error:", err)
			continue
		}
		log.Print("accepted new connection")
		go p.serve(src, dst)
	}
	return <-p.err()
}

func (p *Proxy) serve(src, dst net.Conn) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		io.Copy(src, dst)
		log.Print("src -> dst done")
		wg.Done()
	}()
	go func() {
		io.Copy(dst, src)
		log.Print("dst -> src done")
		wg.Done()
	}()
	wg.Wait()
	dst.Close()
	src.Close()
}

// Stop TODO(rjeczalik)
func (p *Proxy) Stop() {
	p.err() <- p.listener.Close()
}
