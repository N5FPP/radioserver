package main

import (
	"github.com/racerxdl/radioserver/SLog"
	"github.com/racerxdl/radioserver/frontends"
	"github.com/racerxdl/radioserver/protocol"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	SLog.Info("RadioServer - %s", ServerVersion.String())

	var frontend = frontends.CreateAirspyFrontend(0)
	frontend.Init()

	defer frontend.Destroy()

	SLog.Info("Frontend: %s", frontend.GetName())

	serverState.Frontend = frontend

	serverState.DeviceInfo = protocol.DeviceInfo{
		DeviceType:           frontend.GetDeviceType(),
		DeviceSerial:         frontend.GetUintDeviceSerial(),
		MaximumSampleRate:    frontend.GetMaximumSampleRate(),
		MaximumBandwidth:     frontend.GetMaximumBandwidth(),
		DecimationStageCount: frontend.MaximumDecimationStages(),
		GainStageCount:       0,
		MaximumGainIndex:     frontend.MaximumGainIndex(),
		MinimumFrequency:     frontend.MinimumFrequency(),
		MaximumFrequency:     frontend.MaximumFrequency(),
	}

	stop := make(chan bool, 1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<- c
		SLog.Info("Got SIGTERM! Closing it")
		tcpServerStatus = false
		stop <- true
	}()


	runServer(stop)
	SLog.Info("Closing")
}
