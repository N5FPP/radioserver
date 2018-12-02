package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/racerxdl/radioserver/SLog"
	"math/rand"
	"net"
	"sync"
	"time"
)

var tcpSlog = SLog.Scope("TCP Server")
var tcpServerStatus = false
var listenPort = defaultPort
var serverState = ServerState{
	clientListMtx: sync.Mutex{},
	clients: make([]*ClientState, 0),
}

const defaultReadTimeout = 1000

func parseHttpError(err error, state *ClientState) {
	if err.Error() == "EOF" {
		state.running = false
		return
	}

	switch e := err.(type) {
	case net.Error:
		if !e.Timeout() {
			if tcpServerStatus && state.running {
				state.log.Error("Error receiving data: %s", e)
			}
			state.running = false
		}
		break
	default:
		if tcpServerStatus && state.running {
			state.log.Error("Error receiving data: %s", e)
		}
		state.running = false
		break
	}
}

func handleConnection(c net.Conn) {
	var clientState = &ClientState{
		uuid: uuid.New().String(),
		buffer: make([]uint8, 64 * 1024),
		addr: c.RemoteAddr(),
		log: SLog.Scope(fmt.Sprintf("Client %s", c.RemoteAddr())),
		currentState: ParserAcquiringHeader,
		connectedSince: time.Now(),
		receivedBytes: 0,
		sendBytes: 0,
		conn: c,
		running: true,
		packetSent: 0,
		cmdReceived: 0,
		parserPosition: 0,
		headerBuffer: make([]uint8, MessageHeaderSize),
		connMtx: sync.Mutex{},
	}

	serverState.PushClient(clientState)

	tcpSlog.Log("New connection from %s", clientState.addr)

	for {
		if !tcpServerStatus || !clientState.running {
			break
		}

		err := c.SetReadDeadline(time.Now().Add(defaultReadTimeout))
		n, err := c.Read(clientState.buffer)

		if err != nil {
			parseHttpError(err, clientState)
		}

		if !clientState.running {
			break
		}

		if n > 0 {
			clientState.log.Debug("Received %d bytes from client!", n)
			var sl = clientState.buffer[:n]
			parseMessage(clientState, sl)
		}
	}

	serverState.RemoveClient(clientState)
	tcpSlog.Log("Connection closed from %s", clientState.addr)
	c.Close()

}

func runServer() {
	tcpSlog.Info("Starting TCP Server")
	l, err := net.Listen("tcp4", fmt.Sprintf(":%d", listenPort))

	if err != nil {
		tcpSlog.Error("Error listening: %s", err)
		return
	}

	defer l.Close()

	tcpSlog.Info("Listening at port %d", listenPort)

	rand.Seed(time.Now().Unix() + rand.Int63() + rand.Int63())

	tcpServerStatus = true

	for tcpServerStatus {
		c, err := l.Accept()
		if err != nil {
			tcpSlog.Error("Error accepting client: %s", err)
			tcpServerStatus = false
			break
		}
		go handleConnection(c)
	}
}