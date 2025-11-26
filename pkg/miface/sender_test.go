package miface

import (
	"bytes"
	"testing"
)

func TestBuildOSCMessage(t *testing.T) {
	tests := []struct {
		name    string
		address string
		args    []interface{}
	}{
		{
			name:    "address only",
			address: "/test",
			args:    nil,
		},
		{
			name:    "with string",
			address: "/test/string",
			args:    []interface{}{"hello"},
		},
		{
			name:    "with int",
			address: "/test/int",
			args:    []interface{}{int32(42)},
		},
		{
			name:    "with float",
			address: "/test/float",
			args:    []interface{}{float32(3.14)},
		},
		{
			name:    "mixed args",
			address: "/test/mixed",
			args:    []interface{}{"bone", float32(1.0), float32(2.0), float32(3.0)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := buildOSCMessage(tt.address, tt.args...)
			if len(msg) == 0 {
				t.Error("expected non-empty message")
			}

			// Verify address is at start
			if !bytes.HasPrefix(msg, []byte(tt.address)) {
				t.Error("message should start with address")
			}
		})
	}
}

func TestAppendOSCString(t *testing.T) {
	tests := []struct {
		input    string
		expected int // Expected length (with null terminator and padding)
	}{
		{"", 4},      // 1 null + 3 padding = 4
		{"a", 4},     // 1 char + 1 null + 2 padding = 4
		{"ab", 4},    // 2 chars + 1 null + 1 padding = 4
		{"abc", 4},   // 3 chars + 1 null + 0 padding = 4
		{"abcd", 8},  // 4 chars + 1 null + 3 padding = 8
	}

	for _, tt := range tests {
		buf := appendOSCString(nil, tt.input)
		if len(buf) != tt.expected {
			t.Errorf("appendOSCString(%q) = len %d, want %d", tt.input, len(buf), tt.expected)
		}
		// Verify null terminator
		if buf[len(tt.input)] != 0 {
			t.Errorf("expected null terminator at position %d", len(tt.input))
		}
	}
}

func TestAppendInt32(t *testing.T) {
	buf := appendInt32(nil, 0x12345678)
	if len(buf) != 4 {
		t.Errorf("expected length 4, got %d", len(buf))
	}
	// OSC uses big-endian
	expected := []byte{0x12, 0x34, 0x56, 0x78}
	if !bytes.Equal(buf, expected) {
		t.Errorf("got %v, want %v", buf, expected)
	}
}

func TestAppendFloat32(t *testing.T) {
	buf := appendFloat32(nil, 1.0)
	if len(buf) != 4 {
		t.Errorf("expected length 4, got %d", len(buf))
	}
	// 1.0 in IEEE 754 = 0x3F800000
	expected := []byte{0x3F, 0x80, 0x00, 0x00}
	if !bytes.Equal(buf, expected) {
		t.Errorf("got %v, want %v", buf, expected)
	}
}

func TestVMCSenderClose(t *testing.T) {
	// Test closing nil sender
	sender := &VMCSender{}
	if err := sender.Close(); err != nil {
		t.Errorf("closing nil conn should not error: %v", err)
	}
}

func TestVMCSenderSendDisabled(t *testing.T) {
	sender := &VMCSender{enabled: false}
	err := sender.Send(&TrackingData{})
	if err != nil {
		t.Errorf("disabled sender should not error: %v", err)
	}
}
