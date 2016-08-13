package common

import (
	"os"
	"testing"
	"time"
)

const confFile = "../bin/quantum-test.yml"

func TestNewConfig(t *testing.T) {
	os.Setenv("QUANTUM_INTERFACE_NAME", "different")
	os.Setenv("QUANTUM_STATS_WINDOW", "10s")
	os.Setenv("QUANTUM_LISTEN_PORT", "1")
	os.Setenv("QUANTUM_CONF_FILE", confFile)
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig returned an error, %s", err)
	}
	if cfg == nil {
		t.Fatal("NewConfig returned a blank config")
	}
	if cfg.InterfaceName != "different" {
		t.Fatalf("NewConfig didn't pick up the environment variable replacement for InterfaceName")
	}
	if cfg.StatsWindow != 10*time.Second {
		t.Fatalf("NewConfig didn't pick up the environment variable replacement for StatsWindow")
	}
	if cfg.ListenPort != 1 {
		t.Fatalf("NewConfig didn't pick up the environment variable replacement for ListenPort")
	}
}
