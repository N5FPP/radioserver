package main

func CreateDeviceInfo(state *ClientState) []uint8 {
	var deviceInfo = serverState.deviceInfo
	var bodyData = structToBytes(deviceInfo)

	var header = MessageHeader{
		ProtocolID: ProtocolVersion,
		MessageType: MsgTypeDeviceInfo,
		StreamType: StreamTypeStatus,
		SequenceNumber: uint32(state.packetSent & 0xFFFFFFFF),
		BodySize: uint32(len(bodyData)),
	}

	return append(structToBytes(header), bodyData...)
}

func CreateClientSync(state *ClientState) []uint8 {
	var syncInfo = state.syncInfo
	var bodyData = structToBytes(syncInfo)

	var header = MessageHeader{
		ProtocolID: ProtocolVersion,
		MessageType: MsgTypeClientSync,
		StreamType: StreamTypeStatus,
		SequenceNumber: uint32(state.packetSent & 0xFFFFFFFF),
		BodySize: uint32(len(bodyData)),
	}

	return append(structToBytes(header), bodyData...)
}
