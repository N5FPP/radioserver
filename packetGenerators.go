package main

import "time"

func CreateDeviceInfo(state *ClientState) []uint8 {
	var deviceInfo = serverState.deviceInfo
	var bodyData = structToBytes(deviceInfo)

	var header = MessageHeader{
		ProtocolID:     ProtocolVersion,
		MessageType:    MsgTypeDeviceInfo,
		StreamType:     StreamTypeStatus,
		SequenceNumber: uint32(state.packetSent & 0xFFFFFFFF),
		BodySize:       uint32(len(bodyData)),
	}

	return append(structToBytes(header), bodyData...)
}

func CreateClientSync(state *ClientState) []uint8 {
	var syncInfo = state.syncInfo
	var bodyData = structToBytes(syncInfo)

	var header = MessageHeader{
		ProtocolID:     ProtocolVersion,
		MessageType:    MsgTypeClientSync,
		StreamType:     StreamTypeStatus,
		SequenceNumber: uint32(state.packetSent & 0xFFFFFFFF),
		BodySize:       uint32(len(bodyData)),
	}

	return append(structToBytes(header), bodyData...)
}

func CreatePong(state *ClientState) []uint8 {
	var ts = time.Now()
	var pingPacket = PingPacket{
		timestamp: ts.UnixNano(),
	}
	var bodyData = structToBytes(pingPacket)

	var header = MessageHeader{
		ProtocolID:     ProtocolVersion,
		MessageType:    MsgTypePong,
		StreamType:     StreamTypeStatus,
		SequenceNumber: uint32(state.packetSent & 0xFFFFFFFF),
		BodySize:       uint32(len(bodyData)),
	}

	return append(structToBytes(header), bodyData...)
}

func CreateDataPacket(state *ClientState, messageType uint32, samples []interface{}) []uint8 {
	var bodyData = arrayToBytes(samples)

	var header = MessageHeader{
		ProtocolID:     ProtocolVersion,
		MessageType:    messageType,
		StreamType:     StreamTypeStatus,
		SequenceNumber: uint32(state.packetSent & 0xFFFFFFFF),
		BodySize:       uint32(len(bodyData)),
	}

	return append(structToBytes(header), bodyData...)
}
