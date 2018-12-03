package StateModels

import (
	"github.com/racerxdl/radioserver/protocol"
	"github.com/racerxdl/radioserver/tools"
	"time"
)

func CreateDeviceInfo(state *ClientState) []uint8 {
	var deviceInfo = state.ServerState.DeviceInfo
	var bodyData = tools.StructToBytes(deviceInfo)

	var header = protocol.MessageHeader{
		ProtocolID:     state.ServerVersion.ToUint32(),
		MessageType:    protocol.MsgTypeDeviceInfo,
		StreamType:     protocol.StreamTypeStatus,
		SequenceNumber: uint32(state.SentPackets & 0xFFFFFFFF),
		BodySize:       uint32(len(bodyData)),
	}

	return append(tools.StructToBytes(header), bodyData...)
}

func CreateClientSync(state *ClientState) []uint8 {
	var syncInfo = state.SyncInfo
	var bodyData = tools.StructToBytes(syncInfo)

	var header = protocol.MessageHeader{
		ProtocolID:     state.ServerVersion.ToUint32(),
		MessageType:    protocol.MsgTypeClientSync,
		StreamType:     protocol.StreamTypeStatus,
		SequenceNumber: uint32(state.SentPackets & 0xFFFFFFFF),
		BodySize:       uint32(len(bodyData)),
	}

	return append(tools.StructToBytes(header), bodyData...)
}

func CreatePong(state *ClientState) []uint8 {
	var ts = time.Now()
	var pingPacket = protocol.PingPacket{
		Timestamp: ts.UnixNano(),
	}
	var bodyData = tools.StructToBytes(pingPacket)

	var header = protocol.MessageHeader{
		ProtocolID:     state.ServerVersion.ToUint32(),
		MessageType:    protocol.MsgTypePong,
		StreamType:     protocol.StreamTypeStatus,
		SequenceNumber: uint32(state.SentPackets & 0xFFFFFFFF),
		BodySize:       uint32(len(bodyData)),
	}

	return append(tools.StructToBytes(header), bodyData...)
}

func CreateDataPacket(state *ClientState, messageType uint32, samples []interface{}) []uint8 {
	var bodyData = tools.ArrayToBytes(samples)

	var header = protocol.MessageHeader{
		ProtocolID:     state.ServerVersion.ToUint32(),
		MessageType:    messageType,
		StreamType:     state.CGS.StreamingMode,
		SequenceNumber: uint32(state.SentPackets & 0xFFFFFFFF),
		BodySize:       uint32(len(bodyData)),
	}

	return append(tools.StructToBytes(header), bodyData...)
}
