package main

import "sync"

type ServerState struct {
	deviceInfo    DeviceInfo
	clients       []*ClientState
	clientListMtx sync.Mutex
}

func (s *ServerState) indexOfClient(state *ClientState) int {
	for k, v := range s.clients {
		if v.uuid == state.uuid {
			return k
		}
	}

	return -1
}

func (s *ServerState) PushClient(state *ClientState) {
	s.clientListMtx.Lock()
	defer s.clientListMtx.Unlock()

	s.clients = append(s.clients, state)
}

func (s *ServerState) RemoveClient(state *ClientState) {
	s.clientListMtx.Lock()
	defer s.clientListMtx.Unlock()
	idx := s.indexOfClient(state)
	if idx != -1 {
		s.clients = append(s.clients[:idx], s.clients[idx+1:]...)
	}
}

func (s *ServerState) SendSync() bool {
	s.clientListMtx.Lock()
	defer s.clientListMtx.Unlock()

	for i := 0; i < len(s.clients); i++ {
		go s.clients[i].SendSync()
	}

	return true
}

func (s *ServerState) PushSamples(samples []complex64) {

}
