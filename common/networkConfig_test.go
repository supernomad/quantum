package common

import (
	"testing"
)

func TestParseNetworkConfig(t *testing.T) {
	actual, err := ParseNetworkConfig(DefaultNetworkConfig.Bytes())
	if err != nil {
		t.Fatal("ParseNetworkConfig returned an error:", err)
	}
	if actual.Network != DefaultNetworkConfig.Network || actual.LeaseTime != DefaultNetworkConfig.LeaseTime {
		t.Fatalf("ParseNetworkConfig returned the wrong value, got: %v, expected: %v", actual, DefaultNetworkConfig)
	}
}
