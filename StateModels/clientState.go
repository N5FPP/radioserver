package StateModels

import (
	"github.com/google/uuid"
	"github.com/racerxdl/radioserver/SLog"
	"github.com/racerxdl/radioserver/protocol"
	"net"
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

type ClientState struct {
	UUID          string
	Buffer        []uint8
	HeaderBuffer  []uint8
	LogInstance   *SLog.Instance
	Addr          net.Addr
	Conn          net.Conn
	Running       bool
	Name          string
	ClientVersion protocol.Version
	connMtx       sync.Mutex

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
	CG  ChannelGenerator
}

func CreateClientState() *ClientState {
	return &ClientState{
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
		connMtx:        sync.Mutex{},
		CGS: ChannelGeneratorState{
			Streaming:          false,
			StreamingMode:      protocol.StreamModeIQOnly,
			IQFormat:           protocol.StreamFormatInvalid,
			IQCenterFrequency:  0,
			IQDecimation:       0,
			FFTFormat:          protocol.StreamFormatInvalid,
			FFTDecimation:      0,
			FFTDBOffset:        0,
			FFTDisplayPixels:   protocol.DefaultFFTDisplayPixels,
			FFTCenterFrequency: 0,
			FFTDBRange:         protocol.DefaultFFTRange,
		},
	}
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

func (state *ClientState) SendData(buffer []uint8) bool {
	state.connMtx.Lock()
	defer state.connMtx.Unlock()

	_, err := state.Conn.Write(buffer)
	if err != nil {
		state.LogInstance.Error("Error sending data: %s", err)
		return false
	}
	state.SentPackets++

	return true
}

func (state *ClientState) SendSync() {
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
	state.CGS.Streaming = true

	return false
}
func (state *ClientState) SetIQFormat(format uint32) bool {
	state.CGS.IQFormat = format
	return true
}
func (state *ClientState) SetGain(gain uint32) bool {
	state.Error("Set Gain Not implemented!")
	return false
}
func (state *ClientState) SetIQFrequency(frequency uint32) bool {
	state.CGS.IQCenterFrequency = frequency
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
