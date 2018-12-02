package main

import "github.com/racerxdl/radioserver/SLog"

func main() {
	SLog.Info("RadioServer - %s", ServerVersion.String())

	serverState.deviceInfo = DeviceInfo{
		DeviceType: DeviceRtlsdr,
		DeviceSerial: 0,
		MaximumSampleRate: 10e6,
		MaximumBandwidth: 10e6,
		DecimationStageCount: 4,
		GainStageCount: 4,
		MaximumGainIndex: 16,
		MinimumFrequency: 0,
		MaximumFrequency: 1e9,
	}

	runServer()
}