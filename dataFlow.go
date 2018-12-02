package main

import (
	"bytes"
	"encoding/binary"
)

func parseMessage(state *ClientState, buffer []uint8) {
	state.receivedBytes += uint64(len(buffer))

	consumed := uint32(0)

	for len(buffer) > 0 && tcpServerStatus && state.running {
		if state.currentState == ParserAcquiringHeader {
			for state.currentState == ParserAcquiringHeader && len(buffer) > 0 {
				consumed = parseHeader(state, buffer)
				buffer = buffer[consumed:]
			}

			if state.currentState == ParserReadingData {

				if state.cmd.BodySize > MaxMessageBodySize {
					state.Error("Client sent an BodySize of %d which is higher than max %d", state.cmd.BodySize, MaxMessageBodySize)
					state.running = false
					return
				}

				state.cmdBody = make([]uint8, state.cmd.BodySize)
			}
		}

		if state.currentState == ParserReadingData {
			consumed = parseBody(state, buffer)
			buffer = buffer[consumed:]

			if state.currentState == ParserAcquiringHeader {
				state.cmdReceived++
				runCommand(state)
			}
		}
	}
}

func parseBody(state *ClientState, buffer []uint8) uint32 {
	consumed := uint32(0)

	for len(buffer) > 0 {
		toWrite := min(state.cmd.BodySize - state.parserPosition, uint32(len(buffer)))
		for i := uint32(0); i < toWrite; i++ {
			state.cmdBody[i + state.parserPosition] = buffer[i]
		}
		buffer = buffer[toWrite:]
		consumed += toWrite
		state.parserPosition += toWrite

		if state.parserPosition == state.cmd.BodySize {
			state.parserPosition = 0
			state.currentState = ParserAcquiringHeader
			return consumed
		}
	}

	return consumed
}

func parseHeader(state *ClientState, buffer []uint8) uint32 {
	consumed := uint32(0)

	for len(buffer) > 0 {
		toWrite := min(CommandHeaderSize - state.parserPosition, uint32(len(buffer)))
		for i := uint32(0); i < toWrite; i++ {
			state.headerBuffer[i + state.parserPosition] = buffer[i]
		}
		buffer = buffer[toWrite:]
		consumed += toWrite
		state.parserPosition += toWrite

		if state.parserPosition == CommandHeaderSize {
			state.parserPosition = 0
			buf := bytes.NewReader(state.headerBuffer)
			err := binary.Read(buf, binary.LittleEndian, &state.cmd)
			if err != nil {
				panic(err)
			}

			if state.cmd.BodySize > 0 {
				state.currentState = ParserReadingData
			}

			return consumed
		}
	}

	return consumed
}

func runCommand(state *ClientState) {
	var cmdType = state.cmd.CommandType

	if cmdType == CmdHello {
		RunCmdHello(state)
	} else if cmdType == CmdGetSetting {
		RunCmdGetSetting(state)
	} else if cmdType == CmdSetSetting {
		RunCmdSetSetting(state)
	} else if cmdType == CmdPing {
		RunCmdPing(state)
	}
}