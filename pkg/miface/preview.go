//go:build cgo
// +build cgo

package miface

import (
	"runtime"
	"sync"

	"gocv.io/x/gocv"
)

// PreviewWindow provides a simple debug window for camera preview.
// OpenCV UI functions must be called from the main thread on Linux/X11.
type PreviewWindow struct {
	window   *gocv.Window
	frameCh  chan gocv.Mat
	closeCh  chan struct{}
	doneCh   chan struct{}
	once     sync.Once
	initDone chan struct{}
}

// NewPreviewWindow creates a new preview window with the given title.
// Must be called from the main thread.
func NewPreviewWindow(title string) *PreviewWindow {
	p := &PreviewWindow{
		frameCh:  make(chan gocv.Mat, 1),
		closeCh:  make(chan struct{}),
		doneCh:   make(chan struct{}),
		initDone: make(chan struct{}),
	}

	// Start the preview loop in a goroutine locked to OS thread
	go p.previewLoop(title)

	// Wait for initialization to complete
	<-p.initDone

	return p
}

// previewLoop runs the OpenCV UI loop on a dedicated OS thread.
// This is required on Linux/X11 systems.
func (p *PreviewWindow) previewLoop(title string) {
	// Lock this goroutine to an OS thread for OpenCV UI calls
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Create window on this thread
	p.window = gocv.NewWindow(title)
	close(p.initDone)

	for {
		select {
		case frame := <-p.frameCh:
			p.window.IMShow(frame)
			p.window.WaitKey(1)
			frame.Close() // Close the frame after displaying

		case <-p.closeCh:
			if p.window != nil {
				p.window.Close()
			}
			close(p.doneCh)
			return
		}
	}
}

// Show displays a frame in the preview window.
// The frame is cloned internally, so the caller can close the original.
func (p *PreviewWindow) Show(frame gocv.Mat) {
	if frame.Empty() {
		return
	}

	// Clone the frame to avoid race conditions
	cloned := frame.Clone()

	// Non-blocking send - drop frame if channel is full
	select {
	case p.frameCh <- cloned:
	default:
		cloned.Close() // Drop frame if preview is slow
	}
}

// Close closes the preview window and releases resources.
func (p *PreviewWindow) Close() error {
	p.once.Do(func() {
		close(p.closeCh)
		<-p.doneCh
	})
	return nil
}
