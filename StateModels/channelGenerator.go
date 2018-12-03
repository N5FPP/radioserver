package StateModels

import (
	"github.com/racerxdl/go.fifo"
	"github.com/racerxdl/radioserver/SLog"
	"github.com/racerxdl/radioserver/tools"
	"github.com/racerxdl/segdsp/dsp"
	"sync"
)

var cgLog = SLog.Scope("ChannelGenerator")

const maxFifoSize = 32 * 1024 * 1024    // In Bytes
const maxFifoLength = (maxFifoSize) / 8 // In Complex64 Samples

type ChannelGenerator struct {
	decimator           *dsp.FirFilter
	frequencyTranslator *dsp.FrequencyTranslator

	inputFifo        *fifo.Queue
	running          bool
	settingsMutex    sync.Mutex
	routineNotify    *sync.Cond
	routineNotifyMtx *sync.Mutex
}

func CreateChannelGenerator() *ChannelGenerator {
	var mtx = &sync.Mutex{}
	return &ChannelGenerator{
		inputFifo:        fifo.NewQueue(),
		settingsMutex:    sync.Mutex{},
		routineNotify:    sync.NewCond(mtx),
		routineNotifyMtx: mtx,
	}
}

func (cg *ChannelGenerator) routine() {
	cg.routineNotify.L.Lock()
	defer cg.routineNotify.L.Unlock()
	for cg.running {
		cg.routineNotify.Wait() // Wait for samples, or stop
		if !cg.running {
			break
		}
		cg.doWork()
	}
}

func (cg *ChannelGenerator) doWork() {

}

func (cg *ChannelGenerator) notify() {
	cg.routineNotify.L.Lock()
	cg.routineNotify.Signal()
	cg.routineNotify.L.Unlock()
}

func (cg *ChannelGenerator) Start() {
	if !cg.running {
		if cg.frequencyTranslator == nil || cg.decimator == nil {
			cgLog.Fatal("Trying to start Channel Generator without frequencyTranslator or Decimator")
		}
		cg.running = true
		go cg.routine()
	}
}

func (cg *ChannelGenerator) Stop() {
	if cg.running {
		cg.running = false
		cg.notify()
	}
}

func (cg *ChannelGenerator) UpdateSettings(state *ClientState) {
	cg.settingsMutex.Lock()
	defer cg.settingsMutex.Unlock()
}

func (cg *ChannelGenerator) PutSamples(samples []complex64) {
	cg.inputFifo.UnsafeLock()
	defer cg.inputFifo.UnsafeUnlock()

	var fifoLength = cg.inputFifo.UnsafeLen()

	var samplesToAdd = tools.Min(uint32(maxFifoLength-fifoLength), uint32(len(samples)))

	for i := 0; i < int(samplesToAdd); i++ {
		cg.inputFifo.Add(samples[i])
	}

	cg.notify()
}
