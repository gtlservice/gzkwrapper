package gzkwrapper

import (
	"testing"
)

var _server_args = &ServerArgs{
	Hosts:     "10.16.75.23:8481,10.16.75.25:8481,10.16.75.26:8481",
	Root:      "/JobConsole",
	Device:    "",
	Location:  "WH7",
	Pulse:     "10s",
	Timeout:   "35s",
	Threshold: 3,
}

var _server *Server = nil

func createServer(t *testing.T) {

	t.Logf("server args: %v", _server_args)
	s, err := NewServer("0f74ff872189646ee487df1f3dfb1c59f97c9af77", _server_args)
	if err != nil {
		t.Error(err)
		return
	}
	_server = s
	t.Logf("server key: %s", _server.Key)
}

func TestCreateServer(t *testing.T) {

	createServer(t)
	if _server != nil {
		t.Logf("server data: %v", _server.Data)
	} else {
		t.Error("server is nil")
	}
}
