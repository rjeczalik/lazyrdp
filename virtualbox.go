package lazyvm

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

// VirtualBox TODO(rjeczalik)
type VirtualBox struct {
	name string
	addr string
}

func NewVirtualBox(name string) (*VirtualBox, error) {
	return nil, nil
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
	return exec.Command("VBoxManage", "modifyvm", vbox.name, "savestate").Run()
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
	return exec.Command("VBoxManage", "modifyvm", vbox.name, "acpipowerbutton").Run()
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
		s = s[i+1:]
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

func (vbox *VirtualBox) running() (bool, error) {
	out, err := exec.Command("VBoxManage", "list", "runningvms").Output()
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
