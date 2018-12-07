package main

import (
	"flag"
	"fmt"
	"github.com/racerxdl/radioserver/SLog"
	"github.com/racerxdl/radioserver/frontends"
	"github.com/racerxdl/radioserver/protocol"
	"github.com/racerxdl/segdsp/dsp"
	"os"
	"os/signal"
	"runtime/debug"
	"runtime/pprof"
	"syscall"
)

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			SLog.Fatal(err)
		}
		_ = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Got panic", r)
			debug.PrintStack()
			os.Exit(255)
		}
	}()

	SLog.Info("Protocol Version: %s", ServerVersion.String())
	SLog.Info("Commit Hash: %s", commitHash)
	SLog.Info("SIMD Mode: %s", dsp.GetSIMDMode())

	var frontend = frontends.CreateAirspyFrontend(0)
	//var frontend = frontends.CreateLimeSDRFrontend(0)
	frontend.Init()
	frontend.SetCenterFrequency(106300000)

	defer frontend.Destroy()

	SLog.Info("Frontend: %s", frontend.GetName())

	serverState.Frontend = frontend
	serverState.CanControl = 0

	serverState.DeviceInfo = protocol.DeviceInfo{
		DeviceType:           frontend.GetDeviceType(),
		DeviceSerial:         frontend.GetUintDeviceSerial(),
		MaximumSampleRate:    frontend.GetMaximumSampleRate(),
		MaximumBandwidth:     frontend.GetMaximumBandwidth(),
		DecimationStageCount: frontend.MaximumDecimationStages(),
		GainStageCount:       frontend.MaximumGainIndex(),
		MaximumGainIndex:     0,
		MinimumFrequency:     frontend.MinimumFrequency(),
		MaximumFrequency:     frontend.MaximumFrequency(),
		MinimumIQDecimation:  0,
		Resolution:           0,
		ForcedIQFormat:       protocol.StreamFormatFloat,
	}

	frontend.SetSamplesAvailableCallback(serverState.PushSamples)

	stop := make(chan bool, 1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		SLog.Info("Got SIGTERM! Closing it")
		tcpServerStatus = false
		stop <- true
	}()

	// frontend.Start()
	// defer frontend.Stop()
	runServer(stop)
	SLog.Info("Closing")
}
