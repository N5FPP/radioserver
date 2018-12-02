package main


func RunCmdHello(state *ClientState) {
	version, name := ParseCmdHelloBody(state.cmdBody)
	state.Info("Received Hello: %s - %s", version.String(), name)
	state.name = name
	state.clientVersion = version

	data := CreateDeviceInfo(state)
	if !state.SendData(data) {
		state.Error("Error sending deviceInfo packet")
	}

	state.SendSync()
}

func RunCmdGetSetting(state *ClientState) {

}

func RunCmdSetSetting(state *ClientState) {
	setting, args := ParseCmdSetSettingBody(state.cmdBody)

	if !IsSettingPossible(setting) {
		state.Error("Invalid Setting %d", setting)
		return
	}

	settingName := SettingNames[setting]
	state.Debug("Set Setting: %s => %d", settingName, args)

	if !state.SetSetting(setting, args) {
		return
	}

	state.SendSync()

	if SettingAffectsGlobal(setting) {
		go serverState.SendSync()
	}
}

func RunCmdPing(state *ClientState) {
	state.Debug("Received PING")
	state.SendPong()
}