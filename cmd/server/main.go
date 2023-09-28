package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/kittizz/putes/pkg/cert"
	"github.com/kittizz/putes/pkg/protocol"
)

func main() {
	fmt.Println("pid:", os.Getpid())

	if len(os.Args) == 2 {
		if certFile, keyFile, err := cert.GenCerts(); err == nil {
			cert, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				log.Fatalf("Loadkeys : %s", err)
			}
			config := tls.Config{
				Certificates: []tls.Certificate{cert},
			}

			port, err := strconv.Atoi(os.Args[1])
			if err != nil {
				log.Fatalf("invalid port: %s", os.Args[1])
			}
			p := protocol.NewProtocolProvider()
			p.StartListener(port, &config)
			select {}
		}
	} else {
		fmt.Printf("Usage:\nserver-%s <local-port>\n", runtime.GOOS)
	}
}
