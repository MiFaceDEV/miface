//go:build cgo
// +build cgo

package miface

import (
	"testing"
	"time"
)

func TestOpenCVCamera_Open(t *testing.T) {
	camera := NewOpenCVCamera(false)

	// Try to open camera device 0
	err := camera.Open(0, 640, 480, 30)
	if err != nil {
		// Skip test if no camera available
		t.Skipf("Skipping test: no camera available: %v", err)
	}
	defer camera.Close()

	// Verify camera opened
	width, height := camera.GetActualResolution()
	if width <= 0 || height <= 0 {
		t.Errorf("Invalid resolution: %dx%d", width, height)
	}

	fps := camera.GetActualFPS()
	if fps <= 0 {
		t.Errorf("Invalid FPS: %d", fps)
	}
}

func TestOpenCVCamera_Read(t *testing.T) {
	camera := NewOpenCVCamera(false)

	err := camera.Open(0, 640, 480, 30)
	if err != nil {
		t.Skipf("Skipping test: no camera available: %v", err)
	}
	defer camera.Close()

	// Give camera time to stabilize and retry a few times
	var frameData []byte
	var width, height int
	var readErr error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		time.Sleep(100 * time.Millisecond)
		frameData, width, height, readErr = camera.Read()
		if readErr == nil {
			break
		}
	}
	if readErr != nil {
		t.Fatalf("Failed to read frame after %d attempts: %v", maxRetries, readErr)
	}

	if width <= 0 || height <= 0 {
		t.Errorf("Invalid frame dimensions: %dx%d", width, height)
	}

	// RGB24 should have width * height * 3 bytes
	expectedSize := width * height * 3
	if len(frameData) != expectedSize {
		t.Errorf("Expected frame size %d, got %d", expectedSize, len(frameData))
	}
}

func TestOpenCVCamera_Mirror(t *testing.T) {
	camera := NewOpenCVCamera(true)

	if !camera.IsMirror() {
		t.Error("Expected mirror to be enabled")
	}

	camera.SetMirror(false)
	if camera.IsMirror() {
		t.Error("Expected mirror to be disabled")
	}
}

func TestOpenCVCamera_DoubleOpen(t *testing.T) {
	camera := NewOpenCVCamera(false)

	err := camera.Open(0, 640, 480, 30)
	if err != nil {
		t.Skipf("Skipping test: no camera available: %v", err)
	}
	defer camera.Close()

	// Try to open again
	err = camera.Open(0, 640, 480, 30)
	if err == nil {
		t.Error("Expected error when opening already opened camera")
	}
}

func TestOpenCVCamera_ReadWithoutOpen(t *testing.T) {
	camera := NewOpenCVCamera(false)

	_, _, _, err := camera.Read()
	if err == nil {
		t.Error("Expected error when reading from unopened camera")
	}
}

func TestOpenCVCamera_InvalidDevice(t *testing.T) {
	camera := NewOpenCVCamera(false)

	// Try to open a device that likely doesn't exist
	err := camera.Open(999, 640, 480, 30)
	if err == nil {
		camera.Close()
		t.Skip("Device 999 unexpectedly exists")
	}

	// Error should indicate device not found
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestOpenCVCamera_Close(t *testing.T) {
	camera := NewOpenCVCamera(false)

	err := camera.Open(0, 640, 480, 30)
	if err != nil {
		t.Skipf("Skipping test: no camera available: %v", err)
	}

	// Close should succeed
	err = camera.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Second close should be safe
	err = camera.Close()
	if err != nil {
		t.Errorf("Second close failed: %v", err)
	}
}

func TestEnumerateCameras(t *testing.T) {
	devices := EnumerateCameras(5)

	// We can't guarantee any cameras exist, but the function should not panic
	t.Logf("Found %d camera device(s): %v", len(devices), devices)
}

// Benchmark camera read performance
func BenchmarkOpenCVCamera_Read(b *testing.B) {
	camera := NewOpenCVCamera(false)

	err := camera.Open(0, 640, 480, 30)
	if err != nil {
		b.Skipf("Skipping benchmark: no camera available: %v", err)
	}
	defer camera.Close()

	// Warm up
	camera.Read()
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, err := camera.Read()
		if err != nil {
			b.Fatalf("Read failed: %v", err)
		}
	}
}
