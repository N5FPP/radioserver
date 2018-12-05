package StateModels

import (
	"github.com/racerxdl/go.fifo"
	"github.com/racerxdl/radioserver/SLog"
	"github.com/racerxdl/radioserver/protocol"
	"github.com/racerxdl/radioserver/tools"
	"github.com/racerxdl/segdsp/dsp"
	"sync"
	"time"
)

var cgLog = SLog.Scope("ChannelGenerator")

const maxFifoSize = 4096

type OnFFTSamples func(samples []float32)
type OnIQSamples func(samples []complex64)

type ChannelGenerator struct {
	iqFrequencyTranslator  *dsp.FrequencyTranslator
	fftFrequencyTranslator *dsp.FrequencyTranslator

	inputFifo     *fifo.Queue
	running       bool
	settingsMutex sync.Mutex

	fftEnabled bool
	iqEnabled  bool

	onIQSamples   OnIQSamples
	onFFTSamples  OnFFTSamples
	updateChannel chan bool
}

func CreateChannelGenerator() *ChannelGenerator {
	var cg = &ChannelGenerator{
		inputFifo:     fifo.NewQueue(),
		settingsMutex: sync.Mutex{},
		updateChannel: make(chan bool),
	}

	return cg
}

func (cg *ChannelGenerator) routine() {
	defer cg.waitAll()
	for cg.running {
		select {
		case <-cg.updateChannel:
			if !cg.running {
				break
			}
			cg.doWork()
		case <-time.After(1 * time.Second):

		}
		if !cg.running {
			break
		}
	}
}

func (cg *ChannelGenerator) waitAll() {
	var pending = true
	cgLog.Debug("Waiting for all pending to process")
	for pending {
		select {
		case <-cg.updateChannel:
			time.Sleep(time.Millisecond * 10)
		default:
			pending = false
		}
	}
	cgLog.Debug("Routine closed")
}

func (cg *ChannelGenerator) doWork() {
	cg.inputFifo.UnsafeLock()
	defer cg.inputFifo.UnsafeUnlock()
	cg.settingsMutex.Lock()
	defer cg.settingsMutex.Unlock()

	var samples = cg.inputFifo.UnsafeNext().([]complex64)
	//if cg.fftEnabled {
	//	cg.processFFT(samples)
	//}

	if cg.iqEnabled {
		cg.processIQ(samples)
	}
}

func (cg *ChannelGenerator) processIQ(samples []complex64) {
	if cg.iqFrequencyTranslator.GetDecimation() != 1 || cg.iqFrequencyTranslator.GetFrequency() != 0 {
		samples = cg.iqFrequencyTranslator.Work(samples)
	}
	if cg.onIQSamples != nil {
		cg.onIQSamples(samples)
	}
}

func (cg *ChannelGenerator) processFFT(samples []complex64) {
	if cg.fftFrequencyTranslator.GetDecimation() != 1 || cg.fftFrequencyTranslator.GetFrequency() != 0 {
		samples = cg.fftFrequencyTranslator.Work(samples)
	}
	// TODO: FFT
	//if cg.onFFTSamples != nil {
	//	go cg.onFFTSamples(fftSamples)
	//}
}

func (cg *ChannelGenerator) notify() {
	cg.updateChannel <- true
}

func (cg *ChannelGenerator) Start() {
	if !cg.running {
		cgLog.Info("Starting Channel Generator")
		if cg.iqFrequencyTranslator == nil && cg.fftFrequencyTranslator == nil {
			cgLog.Fatal("Trying to start Channel Generator without frequencyTranslator for either IQ or FFT")
		}
		cg.running = true
		go cg.routine()
	}
}

func (cg *ChannelGenerator) Stop() {
	if cg.running {
		cgLog.Info("Stopping")
		cg.running = false
		cg.notify()
	}
}

func (cg *ChannelGenerator) UpdateSettings(state *ClientState) {
	cg.settingsMutex.Lock()
	cgLog.Info("Updating settings")

	var deviceFrequency = state.ServerState.Frontend.GetCenterFrequency()
	var deviceSampleRate = state.ServerState.Frontend.GetSampleRate()

	cg.iqEnabled = (state.CGS.StreamingMode & protocol.StreamTypeIQ) > 0
	cg.fftEnabled = (state.CGS.StreamingMode & protocol.StreamTypeFFT) > 0

	// region IQ Channel
	if cg.iqEnabled {
		var iqDecimationNumber = tools.StageToNumber(state.CGS.IQDecimation)
		var iqFtTaps = tools.GenerateTranslatorTaps(iqDecimationNumber, deviceSampleRate)
		var iqDeltaFrequency = state.CGS.IQCenterFrequency - deviceFrequency
		cg.iqFrequencyTranslator = dsp.MakeFrequencyTranslator(int(iqDecimationNumber), float32(-iqDeltaFrequency), float32(deviceSampleRate), iqFtTaps)
	}
	// endregion
	// region FFT Channel
	if cg.fftEnabled {
		var fftDecimationNumber = tools.StageToNumber(state.CGS.FFTDecimation)
		var fftFtTaps = tools.GenerateTranslatorTaps(fftDecimationNumber, deviceSampleRate)
		var fftDeltaFrequency = state.CGS.FFTCenterFrequency - deviceFrequency
		cg.fftFrequencyTranslator = dsp.MakeFrequencyTranslator(int(fftDecimationNumber), float32(fftDeltaFrequency), float32(deviceSampleRate), fftFtTaps)
	}
	// endregion
	cg.settingsMutex.Unlock()
	if state.CGS.Streaming && !cg.running {
		cg.Start()
	}

	if !state.CGS.Streaming && cg.running {
		cg.Stop()
	}
	cgLog.Info("Settings updated.")
}

func (cg *ChannelGenerator) PushSamples(samples []complex64) {
	cg.inputFifo.UnsafeLock()
	var fifoLength = cg.inputFifo.UnsafeLen()

	if maxFifoSize <= fifoLength {
		cgLog.Debug("Fifo Overflowing!")
		cg.inputFifo.UnsafeUnlock()
		return
	}

	cg.inputFifo.UnsafeAdd(samples)
	cg.inputFifo.UnsafeUnlock()

	cg.notify()
}

func (cg *ChannelGenerator) SetOnIQ(cb OnIQSamples) {
	cg.onIQSamples = cb
}

func (cg *ChannelGenerator) SetOnFFT(cb OnFFTSamples) {
	cg.onFFTSamples = cb
}
