package main

import (
	"github.com/racerxdl/radioserver/SLog"
	"github.com/racerxdl/radioserver/protocol"
)

func main() {
	SLog.Info("RadioServer - %s", ServerVersion.String())

	serverState.DeviceInfo = protocol.DeviceInfo{
		DeviceType:           protocol.DeviceRtlsdr,
		DeviceSerial:         0,
		MaximumSampleRate:    10e6,
		MaximumBandwidth:     10e6,
		DecimationStageCount: 4,
		GainStageCount:       4,
		MaximumGainIndex:     16,
		MinimumFrequency:     0,
		MaximumFrequency:     1e9,
	}

	runServer()
}
