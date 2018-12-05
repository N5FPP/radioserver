package frontends

const SampleTypeFloatIQ = 0
const SampleTypeS16IQ = 1
const SampleTypeS8IQ = 2
const minimumSampleRate = 10e3

type Frontend interface {
	GetDeviceType() uint32
	GetDeviceSerial() string
	GetUintDeviceSerial() uint32
	GetMaximumSampleRate() uint32
	GetMaximumBandwidth() uint32
	SetSampleRate(sampleRate uint32) uint32
	SetCenterFrequency(centerFrequency uint32) uint32
	GetAvailableSampleRates() []uint32
	Start()
	Stop()
	SetAntenna(value string)
	SetAGC(agc bool)
	SetGain(value uint8)
	SetBiasT(value bool)
	GetCenterFrequency() uint32
	GetName() string
	GetShortName() string
	GetSampleRate() uint32
	GetGain() uint8
	SetSamplesAvailableCallback(cb SamplesCallback)
	Init() bool
	Destroy()
	MinimumFrequency() uint32
	MaximumFrequency() uint32
	MaximumGainIndex() uint32
	MaximumDecimationStages() uint32
}

type SamplesCallback func(samples []complex64)
