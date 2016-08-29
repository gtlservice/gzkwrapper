package gzkwrapper

import (
	"testing"
)

var _worker_args = &WorkerArgs{
	Hosts:    "10.16.75.23:8481,10.16.75.25:8481,10.16.75.26:8481",
	Root:     "/JobConsole",
	Device:   "",
	Location: "WH7",
	Pulse:    "10s",
}

var _worker *Worker = nil

func createWorker(t *testing.T) {

	t.Logf("worker args: %v", _worker_args)
	w, err := NewWorker("0f5ca933189646ee487df1f3dfb1c59f97c9cf82", _worker_args)
	if err != nil {
		t.Error(err)
		return
	}
	_worker = w
	t.Logf("worker key: %s", _worker.Key)
}

func TestCreateWorker(t *testing.T) {

	createWorker(t)
	if _worker != nil {
		t.Logf("worker data: %v", _worker.Data)
	} else {
		t.Error("worker is nil")
	}
}

func TestOpenWorker(t *testing.T) {

	createWorker(t)
	if _worker == nil {
		t.Error("worker is nil")
		return
	}

	defer _worker.Close()
	if err := _worker.Open(); err != nil {
		t.Error(err)
	}
	t.Logf("worker data: %v", _worker.Data)
}

func TestSigninWorker(t *testing.T) {

	createWorker(t)
	if _worker == nil {
		t.Error("worker is nil")
		return
	}

	defer _worker.Close()
	if err := _worker.Open(); err != nil {
		t.Error(err)
	}

	if err := _worker.Signin(nil); err != nil {
		t.Error(err)
	}
	t.Logf("worker singin: %d", _worker.Data.Singin)

	if err := _worker.Signout(); err != nil {
		t.Error(err)
	}
	t.Logf("worker singin: %d", _worker.Data.Singin)
	t.Logf("worker data: %v", _worker.Data)
}
