package main

import (
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestRunMainLoopExitsOnSignal(t *testing.T) {
	ticks := make(chan time.Time)
	sig := make(chan os.Signal, 1)
	done := make(chan struct{})

	go func() {
		runMainLoop(ticks, sig, nil)
		close(done)
	}()

	sig <- syscall.SIGTERM

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runMainLoop() did not exit after signal")
	}
}

func TestRunMainLoopChecksServersOnTick(t *testing.T) {
	origRunLua := runLuaScriptFunc
	t.Cleanup(func() {
		runLuaScriptFunc = origRunLua
	})

	var checks atomic.Int32
	runLuaScriptFunc = func(name string, params []string) (string, bool, error) {
		checks.Add(1)
		return "success", true, nil
	}

	server := &BasicServer{
		Name:               "server",
		TopicPowerState:    "power/topic",
		TopicLastSeenState: "lastseen/topic",
		cfg: Server{
			Check: Action{Type: "ping", Params: []string{"127.0.0.1"}},
		},
	}

	ticks := make(chan time.Time, 1)
	sig := make(chan os.Signal, 1)
	done := make(chan struct{})

	go func() {
		runMainLoop(ticks, sig, []*BasicServer{server})
		close(done)
	}()

	ticks <- time.Now()

	deadline := time.After(2 * time.Second)
	for checks.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("runMainLoop() did not invoke Check() on tick")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	sig <- syscall.SIGTERM
	<-done
}
