package main

import (
	"github.com/google/uuid"
	"github.com/racerxdl/radioserver/SLog"
	"net"
	"sync"
	"time"
)

type ChannelGeneratorState struct {
	streaming     bool
	streamingMode uint32

	// Channel Mode
	iqFormat          uint32
	iqCenterFrequency uint32
	iqDecimation      uint32

	// FFT Settings
	fftFormat          uint32
	fftDecimation      uint32
	fftDBOffset        int32
	fftDisplayPixels   uint32
	fftCenterFrequency uint32
	fftDBRange         uint32
}

type ClientState struct {
	uuid          string
	buffer        []uint8
	headerBuffer  []uint8
	log           *SLog.Instance
	addr          net.Addr
	conn          net.Conn
	running       bool
	name          string
	clientVersion Version
	connMtx       sync.Mutex

	currentState   int
	receivedBytes  uint64
	sendBytes      uint64
	connectedSince time.Time
	cmdReceived    uint64
	packetSent     uint64

	// Command State
	cmd            CommandHeader
	cmdBody        []uint8
	parserPosition uint32
	syncInfo       ClientSync

	lastPingTime int64

	// Channel Generator
	cgs ChannelGeneratorState
}

func CreateClientState() *ClientState {
	return &ClientState{
		uuid:           uuid.New().String(),
		buffer:         make([]uint8, 64*1024),
		currentState:   ParserAcquiringHeader,
		connectedSince: time.Now(),
		receivedBytes:  0,
		sendBytes:      0,
		running:        false,
		packetSent:     0,
		cmdReceived:    0,
		parserPosition: 0,
		headerBuffer:   make([]uint8, MessageHeaderSize),
		connMtx:        sync.Mutex{},
		cgs: ChannelGeneratorState{
			streaming:          false,
			streamingMode:      StreamModeIQOnly,
			iqFormat:           StreamFormatInvalid,
			iqCenterFrequency:  0,
			iqDecimation:       0,
			fftFormat:          StreamFormatInvalid,
			fftDecimation:      0,
			fftDBOffset:        0,
			fftDisplayPixels:   defaultFFTDisplayPixels,
			fftCenterFrequency: 0,
			fftDBRange:         defaultFFTRange,
		},
	}
}

func (state *ClientState) Log(str interface{}, v ...interface{}) *ClientState {
	state.log.Log(str, v...)
	return state
}

func (state *ClientState) Info(str interface{}, v ...interface{}) *ClientState {
	state.log.Info(str, v...)
	return state
}

func (state *ClientState) Debug(str interface{}, v ...interface{}) *ClientState {
	state.log.Debug(str, v...)
	return state
}

func (state *ClientState) Warn(str interface{}, v ...interface{}) *ClientState {
	state.log.Warn(str, v...)
	return state
}

func (state *ClientState) Error(str interface{}, v ...interface{}) *ClientState {
	state.log.Error(str, v...)
	return state
}

func (state *ClientState) Fatal(str interface{}, v ...interface{}) {
	state.log.Fatal(str, v)
}

func (state *ClientState) SendData(buffer []uint8) bool {
	state.connMtx.Lock()
	defer state.connMtx.Unlock()

	_, err := state.conn.Write(buffer)
	if err != nil {
		state.log.Error("Error sending data: %s", err)
		return false
	}
	state.packetSent++

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
	case SettingStreamingMode:
		return state.SetStreamingMode(args[0])
	case SettingStreamingEnabled:
		return state.SetStreamingEnabled(args[0] == 1)
	case SettingGain:
		return state.SetGain(args[0])
	case SettingIqFormat:
		return state.SetIQFormat(args[0])
	case SettingIqFrequency:
		return state.SetIQFrequency(args[0])
	case SettingIqDecimation:
		return state.SetIQDecimation(args[0])
	case SettingFFTFormat:
		return state.SetFFTFormat(args[0])
	case SettingFFTFrequency:
		return state.SetFFTFrequency(args[0])
	case SettingFFTDecimation:
		return state.SetFFTDecimation(args[0])
	case SettingFFTDbOffset:
		return state.SetFFTDBOffset(int32(args[0]))
	case SettingFFTDisplayPixels:
		return state.SetFFTDisplayPixels(args[0])
	}

	return false
}

func (state *ClientState) SetStreamingMode(mode uint32) bool {
	state.cgs.streamingMode = mode
	return true
}
func (state *ClientState) SetStreamingEnabled(enabled bool) bool {
	var enabledString = "Enabled"
	if !enabled {
		enabledString = "Disabled"
	}

	state.Log("Streaming %s", enabledString)
	state.cgs.streaming = true

	return false
}
func (state *ClientState) SetIQFormat(format uint32) bool {
	state.cgs.iqFormat = format
	return true
}
func (state *ClientState) SetGain(gain uint32) bool {
	state.Error("Set Gain Not implemented!")
	return false
}
func (state *ClientState) SetIQFrequency(frequency uint32) bool {
	state.cgs.iqCenterFrequency = frequency
	return true
}
func (state *ClientState) SetIQDecimation(decimation uint32) bool {
	if serverState.deviceInfo.DecimationStageCount >= decimation {
		state.cgs.iqDecimation = decimation
		return true
	}

	return false
}
func (state *ClientState) SetFFTFormat(format uint32) bool {
	state.cgs.fftFormat = format
	return true
}

func (state *ClientState) SetFFTFrequency(frequency uint32) bool {
	state.cgs.fftCenterFrequency = frequency
	return true
}
func (state *ClientState) SetFFTDecimation(decimation uint32) bool {
	if serverState.deviceInfo.DecimationStageCount >= decimation {
		state.cgs.fftDecimation = decimation
		return true
	}

	return false
}
func (state *ClientState) SetFFTDBOffset(offset int32) bool {
	state.cgs.fftDBOffset = offset
	return true
}
func (state *ClientState) SetFFTDBRange(fftRange uint32) bool {
	state.cgs.fftDBRange = fftRange
	return false
}
func (state *ClientState) SetFFTDisplayPixels(pixels uint32) bool {
	if pixels >= FFTMinDisplayPixels && pixels <= FFTMaxDisplayPixels {
		state.cgs.fftDisplayPixels = pixels
		return true
	}
	return false
}
