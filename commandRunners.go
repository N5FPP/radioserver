package main

import (
	"github.com/racerxdl/radioserver/StateModels"
	"github.com/racerxdl/radioserver/protocol"
	"time"
)

func RunCmdHello(state *StateModels.ClientState) {
	version, name := protocol.ParseCmdHelloBody(state.CmdBody)
	state.Info("Received Hello: %s - %s", version.String(), name)
	state.Name = name
	state.ClientVersion = version

	data := StateModels.CreateDeviceInfo(state)
	if !state.SendData(data) {
		state.Error("Error sending deviceInfo packet")
	}

	state.SendSync()
}

func RunCmdGetSetting(state *StateModels.ClientState) {
	// TODO
	state.Warn("!!!! RunCmdGetSetting not implemented !!!!")
}

func RunCmdSetSetting(state *StateModels.ClientState) {
	setting, args := protocol.ParseCmdSetSettingBody(state.CmdBody)

	if !protocol.IsSettingPossible(setting) {
		state.Error("Invalid Setting %d", setting)
		return
	}

	settingName := protocol.SettingNames[setting]
	state.Debug("Set Setting: %s => %d", settingName, args)

	if !state.SetSetting(setting, args) {
		return
	}

	state.SendSync()

	if protocol.SettingAffectsGlobal(setting) {
		go serverState.SendSync()
	}
}

func RunCmdPing(state *StateModels.ClientState) {
	timestamp := protocol.ParseCmdPingBody(state.CmdBody)
	delta := float64(time.Now().UnixNano()-timestamp) / 1e6
	state.Debug("Received PING %.2f ms", delta)

	state.LastPingTime = timestamp
	state.SendPong()
}
