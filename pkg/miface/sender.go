package miface

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"sync"
)

// VMCSender sends tracking data using the VMC (Virtual Motion Capture) protocol.
// VMC is an OSC-based protocol commonly used by VTuber applications.
type VMCSender struct {
	mu      sync.Mutex
	conn    *net.UDPConn
	addr    *net.UDPAddr
	enabled bool
}

// NewVMCSender creates a new VMC protocol sender.
func NewVMCSender(address string, port int) (*VMCSender, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return nil, fmt.Errorf("resolving VMC address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("connecting to VMC endpoint: %w", err)
	}

	return &VMCSender{
		conn:    conn,
		addr:    addr,
		enabled: true,
	}, nil
}

// Send transmits tracking data via VMC protocol.
func (v *VMCSender) Send(data *TrackingData) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.enabled || v.conn == nil {
		return nil
	}

	// Send head bone position/rotation if face data available
	if data.Face != nil {
		// VMC /VMC/Ext/Bone/Pos format: address, bone_name, pos_x, pos_y, pos_z, rot_x, rot_y, rot_z, rot_w
		msg := buildOSCMessage("/VMC/Ext/Bone/Pos",
			"Head",
			float32(data.Face.HeadPosition.X),
			float32(data.Face.HeadPosition.Y),
			float32(data.Face.HeadPosition.Z),
			float32(data.Face.HeadRotation.X),
			float32(data.Face.HeadRotation.Y),
			float32(data.Face.HeadRotation.Z),
			float32(data.Face.HeadRotation.W),
		)
		if _, err := v.conn.Write(msg); err != nil {
			return fmt.Errorf("sending head bone: %w", err)
		}

		// Send blend shapes
		for name, value := range data.Face.BlendShapes {
			msg := buildOSCMessage("/VMC/Ext/Blend/Val", name, float32(value))
			if _, err := v.conn.Write(msg); err != nil {
				return fmt.Errorf("sending blend shape %s: %w", name, err)
			}
		}

		// Send blend shape apply signal
		applyMsg := buildOSCMessage("/VMC/Ext/Blend/Apply")
		if _, err := v.conn.Write(applyMsg); err != nil {
			return fmt.Errorf("sending blend apply: %w", err)
		}
	}

	// Send hand bones if available
	if data.LeftHand != nil && len(data.LeftHand.Landmarks) > 0 {
		v.sendHandBones("Left", data.LeftHand)
	}
	if data.RightHand != nil && len(data.RightHand.Landmarks) > 0 {
		v.sendHandBones("Right", data.RightHand)
	}

	return nil
}

// sendHandBones sends VMC bone data for a hand.
func (v *VMCSender) sendHandBones(side string, hand *HandData) {
	if len(hand.Landmarks) < 21 {
		return
	}

	// Map MediaPipe hand landmarks to VMC bone names
	// MediaPipe indices: 0=Wrist, 1-4=Thumb, 5-8=Index, 9-12=Middle, 13-16=Ring, 17-20=Pinky
	boneNames := []string{
		side + "Hand",         // 0: Wrist
		side + "ThumbProximal", // 1
		side + "ThumbIntermediate", // 2
		side + "ThumbDistal",   // 3
		side + "IndexProximal", // 5
		side + "IndexIntermediate", // 6
		side + "IndexDistal",   // 7
		side + "MiddleProximal", // 9
		side + "MiddleIntermediate", // 10
		side + "MiddleDistal",  // 11
		side + "RingProximal",  // 13
		side + "RingIntermediate", // 14
		side + "RingDistal",    // 15
		side + "LittleProximal", // 17
		side + "LittleIntermediate", // 18
		side + "LittleDistal",  // 19
	}

	landmarkIndices := []int{0, 1, 2, 3, 5, 6, 7, 9, 10, 11, 13, 14, 15, 17, 18, 19}

	for i, boneName := range boneNames {
		idx := landmarkIndices[i]
		if idx >= len(hand.Landmarks) {
			continue
		}
		lm := hand.Landmarks[idx]
		msg := buildOSCMessage("/VMC/Ext/Bone/Pos",
			boneName,
			float32(lm.Point.X),
			float32(lm.Point.Y),
			float32(lm.Point.Z),
			float32(0), // rot_x
			float32(0), // rot_y
			float32(0), // rot_z
			float32(1), // rot_w (identity quaternion)
		)
		_, _ = v.conn.Write(msg)
	}
}

// Close releases VMC sender resources.
func (v *VMCSender) Close() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.enabled = false
	if v.conn != nil {
		return v.conn.Close()
	}
	return nil
}

// buildOSCMessage creates an OSC message with the given address and arguments.
// VMC protocol uses OSC for communication.
func buildOSCMessage(address string, args ...interface{}) []byte {
	// OSC message format:
	// - Address pattern (null-terminated, padded to 4 bytes)
	// - Type tag string (null-terminated, padded to 4 bytes)
	// - Arguments

	buf := make([]byte, 0, 256)

	// Write address pattern
	buf = appendOSCString(buf, address)

	// Build type tag string
	typeTag := ","
	for _, arg := range args {
		switch arg.(type) {
		case int32:
			typeTag += "i"
		case float32:
			typeTag += "f"
		case string:
			typeTag += "s"
		}
	}
	buf = appendOSCString(buf, typeTag)

	// Write arguments
	for _, arg := range args {
		switch v := arg.(type) {
		case int32:
			buf = appendInt32(buf, v)
		case float32:
			buf = appendFloat32(buf, v)
		case string:
			buf = appendOSCString(buf, v)
		}
	}

	return buf
}

// appendOSCString appends a null-terminated, 4-byte aligned string.
func appendOSCString(buf []byte, s string) []byte {
	buf = append(buf, []byte(s)...)
	buf = append(buf, 0) // null terminator

	// Pad to 4-byte boundary
	padding := (4 - (len(s)+1)%4) % 4
	for i := 0; i < padding; i++ {
		buf = append(buf, 0)
	}

	return buf
}

// appendInt32 appends a big-endian 32-bit integer.
func appendInt32(buf []byte, v int32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(v))
	return append(buf, b...)
}

// appendFloat32 appends a big-endian 32-bit float.
func appendFloat32(buf []byte, v float32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, math.Float32bits(v))
	return append(buf, b...)
}
