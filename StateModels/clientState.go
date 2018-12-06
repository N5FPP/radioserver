package StateModels

import (
	"github.com/google/uuid"
	"github.com/racerxdl/radioserver/SLog"
	"github.com/racerxdl/radioserver/protocol"
	"github.com/racerxdl/radioserver/tools"
	"net"
	"strings"
	"sync"
	"time"
)

type ChannelGeneratorState struct {
	Streaming     bool
	StreamingMode uint32

	// Channel Mode
	IQFormat          uint32
	IQCenterFrequency uint32
	IQDecimation      uint32

	// FFT Settings
	FFTFormat          uint32
	FFTDecimation      uint32
	FFTDBOffset        int32
	FFTDisplayPixels   uint32
	FFTCenterFrequency uint32
	FFTDBRange         uint32
}

// region ClientState

type ClientState struct {
	sync.Mutex
	UUID           string
	Buffer         []uint8
	HeaderBuffer   []uint8
	LogInstance    *SLog.Instance
	Addr           net.Addr
	Conn           net.Conn
	Running        bool
	Name           string
	ClientVersion  protocol.Version
	CurrentState   int
	ReceivedBytes  uint64
	SentBytes      uint64
	ConnectedSince time.Time
	CmdReceived    uint64
	SentPackets    uint64

	ServerVersion protocol.Version
	ServerState   *ServerState

	// Command State
	Cmd            protocol.CommandHeader
	CmdBody        []uint8
	ParserPosition uint32
	SyncInfo       protocol.ClientSync

	LastPingTime int64

	// Channel Generator
	CGS ChannelGeneratorState
	CG  *ChannelGenerator
}

func CreateClientState(centerFrequency uint32) *ClientState {
	var cs = &ClientState{
		UUID:           uuid.New().String(),
		Buffer:         make([]uint8, 64*1024),
		CurrentState:   protocol.ParserAcquiringHeader,
		ConnectedSince: time.Now(),
		ReceivedBytes:  0,
		SentBytes:      0,
		Running:        false,
		SentPackets:    0,
		CmdReceived:    0,
		ParserPosition: 0,
		LogInstance:    SLog.Scope("ClientState"),
		HeaderBuffer:   make([]uint8, protocol.MessageHeaderSize),
		CGS: ChannelGeneratorState{
			Streaming:          false,
			StreamingMode:      protocol.StreamModeIQOnly,
			IQFormat:           protocol.StreamFormatInvalid,
			IQCenterFrequency:  centerFrequency,
			IQDecimation:       0,
			FFTFormat:          protocol.StreamFormatInvalid,
			FFTDecimation:      0,
			FFTDBOffset:        0,
			FFTDisplayPixels:   protocol.DefaultFFTDisplayPixels,
			FFTCenterFrequency: centerFrequency,
			FFTDBRange:         protocol.DefaultFFTRange,
		},
		CG: CreateChannelGenerator(),
	}

	cs.CG.SetOnFFT(cs.onFFT)
	cs.CG.SetOnIQ(cs.onIQ)

	return cs
}

func (state *ClientState) Log(str interface{}, v ...interface{}) *ClientState {
	state.LogInstance.Log(str, v...)
	return state
}

func (state *ClientState) Info(str interface{}, v ...interface{}) *ClientState {
	state.LogInstance.Info(str, v...)
	return state
}

func (state *ClientState) Debug(str interface{}, v ...interface{}) *ClientState {
	state.LogInstance.Debug(str, v...)
	return state
}

func (state *ClientState) Warn(str interface{}, v ...interface{}) *ClientState {
	state.LogInstance.Warn(str, v...)
	return state
}

func (state *ClientState) Error(str interface{}, v ...interface{}) *ClientState {
	state.LogInstance.Error(str, v...)
	return state
}

func (state *ClientState) Fatal(str interface{}, v ...interface{}) {
	state.LogInstance.Fatal(str, v)
}

func (state *ClientState) FullStop() {
	state.Info("Fully stopping Client")
	state.CG.Stop()
	state.Info("Client stopped")
}

func (state *ClientState) SendData(buffer []uint8) bool {
	state.Lock()
	defer state.Unlock()

	n, err := state.Conn.Write(buffer)
	if err != nil {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "closed") && !strings.Contains(errMsg, "broken pipe") {
			state.LogInstance.Error("Error sending data: %s", err)
		}
		return false
	}

	state.SentPackets++

	if n > 0 {
		state.SentBytes += uint64(n)
	}

	return true
}

func (state *ClientState) onFFT(samples []float32) {
	var samplesToSend interface{}
	var msgType uint32

	switch state.CGS.IQFormat {
	// TODO: DInt4
	case protocol.StreamFormatUint8:
		samplesToSend = tools.Float32ToUInt8(samples)
		msgType = protocol.MsgTypeUint8FFT
	default:
		samplesToSend = nil
	}

	if samplesToSend != nil {
		var data = CreateDataPacket(state, msgType, samplesToSend)
		state.SendData(data)
	}
}

func (state *ClientState) onIQ(samples []complex64) {
	var samplesToSend interface{}
	var msgType uint32

	switch state.CGS.IQFormat {
	case protocol.StreamFormatInt16:
		samplesToSend = tools.Complex64ToInt16(samples)
		msgType = protocol.MsgTypeInt16IQ
	case protocol.StreamFormatUint8:
		samplesToSend = tools.Complex64ToUInt8(samples)
		msgType = protocol.MsgTypeUint8IQ
	case protocol.StreamFormatFloat:
		samplesToSend = samples
		msgType = protocol.MsgTypeFloatIQ
	default:
		samplesToSend = nil
	}

	if samplesToSend != nil {
		state.SendIQ(samplesToSend, msgType)
	}
}

func (state *ClientState) SendIQ(samples interface{}, messageType uint32) {
	var bodyData = tools.ArrayToBytes(samples)

	var header = protocol.MessageHeader{
		ProtocolID:     state.ServerVersion.ToUint32(),
		MessageType:    messageType,
		StreamType:     state.CGS.StreamingMode,
		SequenceNumber: uint32(state.SentPackets & 0xFFFFFFFF),
		BodySize:       uint32(len(bodyData)),
	}

	if len(bodyData) > protocol.MaxMessageBodySize {
		// Segmentation
		for len(bodyData) > 0 {
			chunkSize := tools.Min(protocol.MaxMessageBodySize, uint32(len(bodyData)))
			segment := bodyData[:chunkSize]
			bodyData = bodyData[chunkSize:]
			header.BodySize = uint32(len(segment))
			header.SequenceNumber = uint32(state.SentPackets & 0xFFFFFFFF)
			state.SendData(CreateRawPacket(header, segment))
		}
		return
	}

	state.SendData(CreateRawPacket(header, bodyData))
}

func (state *ClientState) updateSync() {
	state.SyncInfo.FFTCenterFrequency = state.CGS.FFTCenterFrequency
	state.SyncInfo.IQCenterFrequency = state.CGS.IQCenterFrequency
	state.SyncInfo.CanControl = state.ServerState.CanControl
	state.SyncInfo.Gain = uint32(state.ServerState.Frontend.GetGain())
	state.SyncInfo.DeviceCenterFrequency = state.ServerState.Frontend.GetCenterFrequency()

	var halfSampleRate = state.ServerState.Frontend.GetSampleRate() / 2
	var centerFreq = state.CGS.IQCenterFrequency

	state.SyncInfo.MaximumIQCenterFrequency = centerFreq + halfSampleRate
	state.SyncInfo.MinimumIQCenterFrequency = centerFreq - halfSampleRate
	state.SyncInfo.MaximumFFTCenterFrequency = centerFreq + halfSampleRate
	state.SyncInfo.MinimumFFTCenterFrequency = centerFreq - halfSampleRate
}

func (state *ClientState) SendSync() {
	state.updateSync()
	data := CreateClientSync(state)
	if !state.SendData(data) {
		state.Error("Error sending syncInfo packet")
	}
}

func (state *ClientState) SendPong() {
	data := CreatePong(state)
	if !state.SendData(data) {
		state.Error("Error sending pong packet")
	}
}

func (state *ClientState) SetSetting(setting uint32, args []uint32) bool {
	switch setting {
	case protocol.SettingStreamingMode:
		return state.SetStreamingMode(args[0])
	case protocol.SettingStreamingEnabled:
		return state.SetStreamingEnabled(args[0] == 1)
	case protocol.SettingGain:
		return state.SetGain(args[0])
	case protocol.SettingIqFormat:
		return state.SetIQFormat(args[0])
	case protocol.SettingIqFrequency:
		return state.SetIQFrequency(args[0])
	case protocol.SettingIqDecimation:
		return state.SetIQDecimation(args[0])
	case protocol.SettingFFTFormat:
		return state.SetFFTFormat(args[0])
	case protocol.SettingFFTFrequency:
		return state.SetFFTFrequency(args[0])
	case protocol.SettingFFTDecimation:
		return state.SetFFTDecimation(args[0])
	case protocol.SettingFFTDbOffset:
		return state.SetFFTDBOffset(int32(args[0]))
	case protocol.SettingFFTDisplayPixels:
		return state.SetFFTDisplayPixels(args[0])
	}

	return false
}

func (state *ClientState) SetStreamingMode(mode uint32) bool {
	state.CGS.StreamingMode = mode
	return true
}
func (state *ClientState) SetStreamingEnabled(enabled bool) bool {
	var enabledString = "Enabled"
	if !enabled {
		enabledString = "Disabled"
	}

	state.Log("Streaming %s", enabledString)
	state.CGS.Streaming = enabled

	return true
}
func (state *ClientState) SetIQFormat(format uint32) bool {
	state.CGS.IQFormat = format
	return true
}
func (state *ClientState) SetGain(gain uint32) bool {
	state.ServerState.Frontend.SetGain(uint8(gain))
	return true
}
func (state *ClientState) SetIQFrequency(frequency uint32) bool {
	state.CGS.IQCenterFrequency = frequency
	state.updateSync()
	return true
}
func (state *ClientState) SetIQDecimation(decimation uint32) bool {
	if state.ServerState.DeviceInfo.DecimationStageCount >= decimation {
		state.CGS.IQDecimation = decimation
		return true
	}

	return false
}
func (state *ClientState) SetFFTFormat(format uint32) bool {
	state.CGS.FFTFormat = format
	return true
}

func (state *ClientState) SetFFTFrequency(frequency uint32) bool {
	state.CGS.FFTCenterFrequency = frequency
	state.updateSync()
	return true
}

func (state *ClientState) SetFFTDecimation(decimation uint32) bool {
	if state.ServerState.DeviceInfo.DecimationStageCount >= decimation {
		state.CGS.FFTDecimation = decimation
		return true
	}

	return false
}

func (state *ClientState) SetFFTDBOffset(offset int32) bool {
	state.CGS.FFTDBOffset = offset
	return true
}
func (state *ClientState) SetFFTDBRange(fftRange uint32) bool {
	state.CGS.FFTDBRange = fftRange
	return false
}
func (state *ClientState) SetFFTDisplayPixels(pixels uint32) bool {
	if pixels >= protocol.FFTMinDisplayPixels && pixels <= protocol.FFTMaxDisplayPixels {
		state.CGS.FFTDisplayPixels = pixels
		return true
	}
	return false
}

// endregion
