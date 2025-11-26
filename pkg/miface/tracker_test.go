package miface

import (
	"testing"
	"time"
)

func TestNewTracker(t *testing.T) {
	tracker, err := NewTracker(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tracker == nil {
		t.Fatal("expected non-nil tracker")
	}
	defer tracker.Close()

	if tracker.State() != StateIdle {
		t.Errorf("expected state Idle, got %s", tracker.State())
	}
}

func TestTrackerStartStop(t *testing.T) {
	tracker, err := NewTracker(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tracker.Close()

	// Start tracker
	if err := tracker.Start(); err != nil {
		t.Fatalf("failed to start: %v", err)
	}
	if tracker.State() != StateRunning {
		t.Errorf("expected state Running, got %s", tracker.State())
	}

	// Double start should fail
	if err := tracker.Start(); err != ErrTrackerRunning {
		t.Errorf("expected ErrTrackerRunning, got %v", err)
	}

	// Stop tracker
	if err := tracker.Stop(); err != nil {
		t.Fatalf("failed to stop: %v", err)
	}
	if tracker.State() != StateStopped {
		t.Errorf("expected state Stopped, got %s", tracker.State())
	}

	// Double stop should fail
	if err := tracker.Stop(); err != ErrTrackerStopped {
		t.Errorf("expected ErrTrackerStopped, got %v", err)
	}
}

func TestTrackerClose(t *testing.T) {
	tracker, err := NewTracker(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := tracker.Start(); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	// Close should stop and cleanup
	if err := tracker.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}
	if tracker.State() != StateClosed {
		t.Errorf("expected state Closed, got %s", tracker.State())
	}

	// Double close should fail
	if err := tracker.Close(); err != ErrTrackerClosed {
		t.Errorf("expected ErrTrackerClosed, got %v", err)
	}

	// Start after close should fail
	if err := tracker.Start(); err != ErrTrackerClosed {
		t.Errorf("expected ErrTrackerClosed, got %v", err)
	}
}

func TestTrackerSubscribe(t *testing.T) {
	tracker, err := NewTracker(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tracker.Close()

	// Subscribe before start
	ch := tracker.Subscribe()
	if ch == nil {
		t.Fatal("expected non-nil channel")
	}

	// Start and wait for data
	if err := tracker.Start(); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	// Should receive data within reasonable time
	select {
	case data := <-ch:
		if data == nil {
			t.Fatalf("received nil data")
		}
		if data.FrameNumber == 0 {
			t.Error("expected non-zero frame number")
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("timeout waiting for tracking data")
	}
}

func TestTrackerState(t *testing.T) {
	tests := []struct {
		state TrackerState
		str   string
	}{
		{StateIdle, "idle"},
		{StateRunning, "running"},
		{StateStopped, "stopped"},
		{StateClosed, "closed"},
		{TrackerState(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.str {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.str)
		}
	}
}

// MockCameraSource implements CameraSource for testing.
type MockCameraSource struct {
	opened bool
	closed bool
}

func (m *MockCameraSource) Open(deviceID, width, height, fps int) error {
	m.opened = true
	return nil
}

func (m *MockCameraSource) Read() ([]byte, int, int, error) {
	return make([]byte, 640*480*3), 640, 480, nil
}

func (m *MockCameraSource) Close() error {
	m.closed = true
	return nil
}

// MockProcessor implements Processor for testing.
type MockProcessor struct {
	closed bool
}

func (m *MockProcessor) Process(ctx interface{}, frame []byte, width, height int) (*TrackingData, error) {
	return &TrackingData{
		Timestamp: time.Now(),
		Face: &FaceData{
			Landmarks:    make([]Landmark, 468),
			BlendShapes:  map[string]float64{"smile": 0.5},
			HeadRotation: Quaternion{W: 1},
		},
	}, nil
}

func (m *MockProcessor) Close() error {
	m.closed = true
	return nil
}

func TestTrackerWithMockComponents(t *testing.T) {
	tracker, err := NewTracker(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	camera := &MockCameraSource{}
	if err := tracker.SetCameraSource(camera); err != nil {
		t.Fatalf("failed to set camera: %v", err)
	}

	if err := tracker.Start(); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	// Cannot set camera while running
	if err := tracker.SetCameraSource(&MockCameraSource{}); err == nil {
		t.Error("expected error setting camera while running")
	}

	if err := tracker.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	if !camera.closed {
		t.Error("expected camera to be closed")
	}
}
