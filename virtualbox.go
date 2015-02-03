package lazyvm

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

// ErrNotExist TODO(rjeczalik)
var ErrNotExist = errors.New("lazyvm: vm does not exist")

// VirtualBox TODO(rjeczalik)
type VirtualBox struct {
	name string
}

// NewVirtualBox TODO(rjeczalik)
func NewVirtualBox(name string) (*VirtualBox, error) {
	vbox := &VirtualBox{name: name}
	switch ok, err := vbox.exists(); {
	case err != nil:
		return nil, err
	case !ok:
		return nil, ErrNotExist
	}
	return vbox, nil
}

// Start TODO(rjeczalik)
func (vbox *VirtualBox) Start() error {
	ok, err := vbox.running()
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return exec.Command("VBoxManage", "startvm", vbox.name, "-type", "headless").Run()
}

// Hibernate TODO(rjeczalik)
func (vbox *VirtualBox) Hibernate() error {
	ok, err := vbox.running()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return exec.Command("VBoxManage", "controlvm", vbox.name, "savestate").Run()
}

// Close TODO(rjeczalik)
func (vbox *VirtualBox) Close() error {
	ok, err := vbox.running()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return exec.Command("VBoxManage", "controlvm", vbox.name, "acpipowerbutton").Run()
}

var errParseAddr = errors.New("lazyvm: unable to parse network address")

// Addr TODO(rjeczalik)
func (vbox *VirtualBox) Addr() (string, error) {
	ok, err := vbox.running()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("machine is not started")
	}
	out, err := exec.Command("VBoxManage", "guestproperty", "enumerate", vbox.name).Output()
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		s := scanner.Text()
		i := strings.Index(s, "/V4/IP")
		if i == -1 {
			continue
		}
		s = s[i+1:]
		if i = strings.Index(s, "value:"); i == -1 {
			continue
		}
		s = s[i+len("value:")+1:]
		if i = strings.IndexByte(s, ','); i == -1 {
			continue
		}
		if s = strings.TrimSpace(s[:i]); s == "" {
			continue
		}
		return s, nil
	}
	return "", nonil(scanner.Err(), errParseAddr)
}

func (vbox *VirtualBox) list(cmd string) (bool, error) {
	out, err := exec.Command("VBoxManage", "list", cmd).Output()
	if err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	name := `"` + vbox.name + `" `
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), name) {
			return true, nil
		}
	}
	return false, scanner.Err()
}

func (vbox *VirtualBox) exists() (bool, error) {
	return vbox.list("vms")
}

func (vbox *VirtualBox) running() (bool, error) {
	return vbox.list("runningvms")
}
