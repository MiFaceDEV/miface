//go:build cgo
// +build cgo

package miface

import (
	"testing"
	"time"

	"gocv.io/x/gocv"
)

func TestNewPreviewWindow(t *testing.T) {
	preview := NewPreviewWindow("Test Window")
	if preview == nil {
		t.Fatal("NewPreviewWindow returned nil")
	}
	defer preview.Close()
}

func TestPreviewWindow_Show(t *testing.T) {
	preview := NewPreviewWindow("Test Window")
	defer preview.Close()

	// Create a simple test image
	mat := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer mat.Close()

	// This should not panic
	preview.Show(mat)

	// Give it a moment to process
	time.Sleep(50 * time.Millisecond)
}

func TestPreviewWindow_Close(t *testing.T) {
	preview := NewPreviewWindow("Test Window")

	err := preview.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Second close should be safe (once.Do)
	err = preview.Close()
	if err != nil {
		t.Errorf("Second Close() returned error: %v", err)
	}
}

func TestPreviewWindow_ShowMultiple(t *testing.T) {
	preview := NewPreviewWindow("Test Window")
	defer preview.Close()

	// Show multiple frames
	for i := 0; i < 5; i++ {
		mat := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
		preview.Show(mat)
		mat.Close()
		time.Sleep(10 * time.Millisecond)
	}
}
