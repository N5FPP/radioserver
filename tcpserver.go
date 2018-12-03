package main

import (
	"fmt"
	"github.com/racerxdl/radioserver/SLog"
	"github.com/racerxdl/radioserver/StateModels"
	"github.com/racerxdl/radioserver/protocol"
	"math/rand"
	"net"
	"time"
)

var tcpSlog = SLog.Scope("TCP Server")
var tcpServerStatus = false
var listenPort = protocol.DefaultPort
var serverState = StateModels.CreateServerState()

const defaultReadTimeout = 1000

func parseHttpError(err error, state *StateModels.ClientState) {
	if err.Error() == "EOF" {
		state.Running = false
		return
	}

	switch e := err.(type) {
	case net.Error:
		if !e.Timeout() {
			if tcpServerStatus && state.Running {
				state.Error("Error receiving data: %s", e)
			}
			state.Running = false
		}
		break
	default:
		if tcpServerStatus && state.Running {
			state.Error("Error receiving data: %s", e)
		}
		state.Running = false
		break
	}
}

func handleConnection(c net.Conn) {
	var clientState = StateModels.CreateClientState()

	clientState.Addr = c.RemoteAddr()
	clientState.LogInstance = SLog.Scope(fmt.Sprintf("Client %s", c.RemoteAddr()))
	clientState.Conn = c
	clientState.Running = true
	clientState.ServerState = serverState

	serverState.PushClient(clientState)

	tcpSlog.Log("New connection from %s", clientState.Addr)

	for {
		if !tcpServerStatus || !clientState.Running {
			break
		}

		err := c.SetReadDeadline(time.Now().Add(defaultReadTimeout))
		n, err := c.Read(clientState.Buffer)

		if err != nil {
			parseHttpError(err, clientState)
		}

		if !clientState.Running {
			break
		}

		if n > 0 {
			clientState.Debug("Received %d bytes from client!", n)
			var sl = clientState.Buffer[:n]
			parseMessage(clientState, sl)
		}
	}

	serverState.RemoveClient(clientState)
	tcpSlog.Log("Connection closed from %s", clientState.Addr)
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
