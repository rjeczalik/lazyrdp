// Command lazyrdp provides standalone command line tools for proxying RDP
// connections to VirtualBox's Windows instance. It starts the machine in
// headless mode, pauses on idle and resumes on new connection.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/rjeczalik/lazyrdp"
)

const usage = `lazyrdp - starts Windows machine on incoming RDP connection

OPTIONS:

	-addr ADDR     Network address to listen on. Default is ":5000".
	-port PORT     Port of the machine's RDP server. Default is 3389.
	-v             Turn on verbose logging.

USAGE:

	lazyrdp -help
	lazyrdp [OPTION...] MACHINE_NAME

EXAMPLE:

	lazyrdp -addr localhost:5001 -port 3390 win2012r2`

var signals = []os.Signal{
	os.Kill,
	os.Interrupt,
}

var proxy lazyrdp.Proxy

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}

func init() {
	flag.CommandLine.Usage = func() {
		fmt.Println(usage)
		os.Exit(0)
	}
	help := flag.Bool("help", false, "")
	verbose := flag.Bool("v", false, "")
	flag.StringVar(&proxy.Addr, "addr", ":5000", "")
	flag.IntVar(&proxy.Port, "port", 3389, "")
	flag.Parse()
	if *help {
		flag.CommandLine.Usage()
	}
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	switch flag.NArg() {
	case 0:
		die("lazydrp: machine name is missing")
	case 1:
		proxy.MachineName = flag.Arg(0)
	default:
		die("lazydrp: too many arguments")
	}
}

func sighandler(c <-chan os.Signal) {
	once := sync.Once{}
	for range c {
		go once.Do(proxy.Stop)
	}
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	go sighandler(c)
	if err := proxy.Run(); err != nil {
		die(err)
	}
}
