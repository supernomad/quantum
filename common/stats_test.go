package common

import (
	"testing"
)

func TestNewStats(t *testing.T) {
	stats := NewStats()
	if stats.Packets != 0 {
		t.Fatalf("NewStats did not return the correct default for Packets, got: %d, expected: %d", stats.Packets, 0)
	}
	if stats.Bandwidth != 0 {
		t.Fatalf("NewStats did not return the correct default for Bandwidth, got: %.6f, expected: %.6f", stats.Bandwidth, 0.0)
	}
	if stats.Bytes != 0 {
		t.Fatalf("NewStats did not return the correct default for Bytes, got: %d, expected: %d", stats.Bytes, 0)
	}
	if stats.PPS != 0 {
		t.Fatalf("NewStats did not return the correct default for PPS, got: %.6f, expected: %.6f", stats.PPS, 0.0)
	}
	if stats.Links == nil {
		t.Fatalf("NewStats did not return the correct default for Links, got: %v, expected: %v", stats.Links, make(map[string]*Stats))
	}
}
