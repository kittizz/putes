package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/kittizz/reverse-shell/pkg/protocol"
)

func main() {
	port := flag.Int("l", 0, "Tunnel provider signaling port")
	providerAddress := flag.String("c", "", "Tunnel provider signaling address")
	targetAddress := flag.String("t", "", "Target address to be tunnelled")

	flag.Parse()

	p := protocol.NewProtocolProvider()

	if *port != 0 {
		p.StartListener(*port, nil)

		// no graceful shutdown yet
		select {}
	} else {
		if len(*providerAddress) == 0 || len(*targetAddress) == 0 {
			fmt.Printf("Usage: tunnel [-l] [[-c] [-t]]\n")
			return
		}

		c, err := p.StartConnector(*providerAddress)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}

		addr := strings.Split(*targetAddress, ":")
		targetPort := 443
		if len(addr) > 1 {
			targetPort, _ = strconv.Atoi(addr[1])
		}

		c.StartTunnelFor(addr[0], targetPort)

		// no graceful shutdown yet
		select {}
	}
}
