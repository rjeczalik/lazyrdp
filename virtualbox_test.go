package lazyvm

import (
	"net"
	"os"
	"testing"
)

var vbox, vboxerr = NewVirtualBox(os.Getenv("TEST_LAZYRDP"))

func connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	return conn.Close()
}

func TestVirtualBox(t *testing.T) {
	if vboxerr != nil {
		t.Skip("skipping as $TEST_LAZYRDP is not a valid vm:", vboxerr)
	}
	if err := vbox.Start(); err != nil {
		t.Fatalf("vbox.Start()=%v", err)
	}
	defer func() {
		if err := vbox.Close(); err != nil {
			t.Fatalf("vbox.Close()=%v", err)
		}
	}()
	addr, err := vbox.Addr()
	if err != nil {
		t.Fatalf("vbox.Addr()=%v", err)
	}
	if err = connect(addr); err != nil {
		t.Fatalf("connect(%q)=%v", addr, err)
	}
	if err = vbox.Hibernate(); err != nil {
		t.Fatalf("vbox.Hibernate()=%v", err)
	}
	if err = vbox.Start(); err != nil {
		t.Fatalf("vbox.Start()=%v", err)
	}
	if addr, err = vbox.Addr(); err != nil {
		t.Fatalf("vbox.Addr()=%v", err)
	}
	if err = connect(addr); err != nil {
		t.Fatalf("connect(%q)=%v", addr, err)
	}
}
