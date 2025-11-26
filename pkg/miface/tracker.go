// Package miface provides real-time facial and upper body tracking for VTubers.
//
// MiFace integrates MediaPipe Holistic for face, hand, and pose landmark detection,
// applies Kalman filtering for smooth tracking, and sends data via VMC/OSC protocols.
//
// # Quick Start
//
// Create a tracker with default configuration:
//
//	tracker, err := miface.NewTracker(nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer tracker.Close()
//
//	// Start tracking
//	if err := tracker.Start(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Subscribe to tracking data
//	dataCh := tracker.Subscribe()
//	for data := range dataCh {
//	    fmt.Printf("Face landmarks: %d\n", len(data.Face.Landmarks))
//	}
//
// # Custom Configuration
//
// Load configuration from TOML file:
//
//	import "github.com/MiFaceDEV/miface/internal/config"
//
//	cfg, _ := config.Load("config.toml")
//	tracker, err := miface.NewTracker(cfg)
//
// # Architecture
//
// MiFace follows a library-first design for maximum reusability:
//
//   - Tracker: Main coordinator managing capture, tracking, and output
//   - CameraSource: Webcam capture abstraction (pluggable)
//   - MediaPipeProcessor: MediaPipe Holistic integration interface
//   - KalmanFilter: Smoothing filter for landmark stabilization
//   - VMCSender/OSCSender: Protocol senders for VTuber applications
//
// All components are concurrent-safe and designed for real-time performance.
package miface

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/MiFaceDEV/miface/internal/config"
)

// Common errors returned by MiFace.
var (
	ErrTrackerClosed   = errors.New("tracker is closed")
	ErrTrackerRunning  = errors.New("tracker is already running")
	ErrTrackerStopped  = errors.New("tracker is not running")
	ErrCameraNotFound  = errors.New("camera device not found")
	ErrMediaPipeInit   = errors.New("failed to initialize MediaPipe")
)

// Point3D represents a 3D coordinate.
type Point3D struct {
	X, Y, Z float64
}

// Landmark represents a tracked landmark point with visibility confidence.
type Landmark struct {
	Point      Point3D
	Visibility float64 // 0.0 to 1.0 confidence score
}

// Quaternion represents a rotation in 3D space.
type Quaternion struct {
	X, Y, Z, W float64
}

// FaceData contains face tracking results.
type FaceData struct {
	// Landmarks contains 468 face mesh landmarks (MediaPipe standard).
	Landmarks []Landmark
	// BlendShapes contains facial expression blend shape weights.
	BlendShapes map[string]float64
	// HeadRotation is the estimated head rotation.
	HeadRotation Quaternion
	// HeadPosition is the estimated head position.
	HeadPosition Point3D
}

// HandData contains hand tracking results for a single hand.
type HandData struct {
	// IsLeft indicates if this is the left hand.
	IsLeft bool
	// Landmarks contains 21 hand landmarks (MediaPipe standard).
	Landmarks []Landmark
	// Confidence is the hand detection confidence (0.0 to 1.0).
	Confidence float64
}

// PoseData contains body pose tracking results.
type PoseData struct {
	// Landmarks contains 33 pose landmarks (MediaPipe standard).
	Landmarks []Landmark
}

// TrackingData contains all tracking results for a single frame.
type TrackingData struct {
	// Timestamp is when this data was captured.
	Timestamp time.Time
	// FrameNumber is the sequential frame number.
	FrameNumber uint64
	// Face contains face tracking data (nil if face tracking disabled).
	Face *FaceData
	// LeftHand contains left hand tracking data (nil if not detected).
	LeftHand *HandData
	// RightHand contains right hand tracking data (nil if not detected).
	RightHand *HandData
	// Pose contains body pose tracking data (nil if pose tracking disabled).
	Pose *PoseData
}

// TrackerState represents the current state of the tracker.
type TrackerState int

const (
	// StateIdle means the tracker is initialized but not running.
	StateIdle TrackerState = iota
	// StateRunning means the tracker is actively capturing and processing.
	StateRunning
	// StateStopped means the tracker has been stopped.
	StateStopped
	// StateClosed means the tracker has been closed and cannot be reused.
	StateClosed
)

func (s TrackerState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	case StateStopped:
		return "stopped"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// CameraSource is the interface for camera capture backends.
type CameraSource interface {
	// Open initializes the camera with the given configuration.
	Open(deviceID, width, height, fps int) error
	// Read captures a single frame. Returns the frame data or error.
	Read() ([]byte, int, int, error)
	// Close releases camera resources.
	Close() error
}

// Processor is the interface for landmark detection processors.
type Processor interface {
	// Process analyzes a frame and returns tracking data.
	Process(ctx context.Context, frame []byte, width, height int) (*TrackingData, error)
	// Close releases processor resources.
	Close() error
}

// Sender is the interface for protocol output senders.
type Sender interface {
	// Send transmits tracking data.
	Send(data *TrackingData) error
	// Close releases sender resources.
	Close() error
}

// Tracker is the main coordinator for face/body tracking.
type Tracker struct {
	cfg *config.Config

	mu          sync.RWMutex
	state       TrackerState
	camera      CameraSource
	processor   Processor
	vmcSender   Sender
	oscSender   Sender
	subscribers []chan *TrackingData

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	frameCount uint64
}

// NewTracker creates a new tracker with the given configuration.
// If cfg is nil, default configuration is used.
func NewTracker(cfg *config.Config) (*Tracker, error) {
	if cfg == nil {
		cfg = config.Default()
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &Tracker{
		cfg:   cfg,
		state: StateIdle,
	}, nil
}

// Config returns the current configuration.
func (t *Tracker) Config() *config.Config {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cfg
}

// State returns the current tracker state.
func (t *Tracker) State() TrackerState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

// SetCameraSource sets a custom camera source.
// Must be called before Start().
func (t *Tracker) SetCameraSource(camera CameraSource) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state != StateIdle {
		return fmt.Errorf("cannot set camera source: tracker is %s", t.state)
	}
	t.camera = camera
	return nil
}

// SetProcessor sets a custom landmark processor.
// Must be called before Start().
func (t *Tracker) SetProcessor(processor Processor) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state != StateIdle {
		return fmt.Errorf("cannot set processor: tracker is %s", t.state)
	}
	t.processor = processor
	return nil
}

// SetVMCSender sets the VMC protocol sender.
// Must be called before Start().
func (t *Tracker) SetVMCSender(sender Sender) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state != StateIdle {
		return fmt.Errorf("cannot set VMC sender: tracker is %s", t.state)
	}
	t.vmcSender = sender
	return nil
}

// SetOSCSender sets the OSC protocol sender.
// Must be called before Start().
func (t *Tracker) SetOSCSender(sender Sender) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state != StateIdle {
		return fmt.Errorf("cannot set OSC sender: tracker is %s", t.state)
	}
	t.oscSender = sender
	return nil
}

// Subscribe returns a channel that receives tracking data.
// The caller must drain the channel or risk blocking the tracker.
// Close the tracker to close all subscriber channels.
func (t *Tracker) Subscribe() <-chan *TrackingData {
	t.mu.Lock()
	defer t.mu.Unlock()

	ch := make(chan *TrackingData, 10)
	t.subscribers = append(t.subscribers, ch)
	return ch
}

// Start begins the tracking loop.
// Returns immediately; tracking runs in background goroutines.
func (t *Tracker) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch t.state {
	case StateRunning:
		return ErrTrackerRunning
	case StateClosed:
		return ErrTrackerClosed
	}

	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.state = StateRunning
	t.frameCount = 0

	t.wg.Add(1)
	go t.trackingLoop()

	return nil
}

// Stop stops the tracking loop.
func (t *Tracker) Stop() error {
	t.mu.Lock()

	if t.state != StateRunning {
		t.mu.Unlock()
		return ErrTrackerStopped
	}

	t.cancel()
	t.state = StateStopped
	t.mu.Unlock()

	t.wg.Wait()
	return nil
}

// Close stops tracking and releases all resources.
func (t *Tracker) Close() error {
	t.mu.Lock()
	if t.state == StateClosed {
		t.mu.Unlock()
		return ErrTrackerClosed
	}

	if t.state == StateRunning {
		t.cancel()
	}
	t.state = StateClosed
	t.mu.Unlock()

	t.wg.Wait()

	var errs []error

	t.mu.Lock()
	if t.camera != nil {
		if err := t.camera.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing camera: %w", err))
		}
	}
	if t.processor != nil {
		if err := t.processor.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing processor: %w", err))
		}
	}
	if t.vmcSender != nil {
		if err := t.vmcSender.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing VMC sender: %w", err))
		}
	}
	if t.oscSender != nil {
		if err := t.oscSender.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing OSC sender: %w", err))
		}
	}

	// Close subscriber channels
	for _, ch := range t.subscribers {
		close(ch)
	}
	t.subscribers = nil
	t.mu.Unlock()

	if len(errs) > 0 {
		return fmt.Errorf("closing tracker: %v", errs)
	}
	return nil
}

// trackingLoop is the main capture and processing loop.
func (t *Tracker) trackingLoop() {
	defer t.wg.Done()

	ticker := time.NewTicker(time.Second / time.Duration(t.cfg.Camera.FPS))
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			t.processFrame()
		}
	}
}

// processFrame captures and processes a single frame.
func (t *Tracker) processFrame() {
	t.mu.RLock()
	camera := t.camera
	processor := t.processor
	vmcSender := t.vmcSender
	oscSender := t.oscSender
	t.mu.RUnlock()

	// Generate mock data if no camera/processor configured
	var data *TrackingData
	if camera != nil && processor != nil {
		frame, width, height, err := camera.Read()
		if err != nil {
			return
		}

		var pErr error
		data, pErr = processor.Process(t.ctx, frame, width, height)
		if pErr != nil {
			return
		}
	} else {
		// Generate stub tracking data for testing
		data = &TrackingData{
			Timestamp:   time.Now(),
			FrameNumber: t.frameCount,
		}
	}

	t.frameCount++
	data.FrameNumber = t.frameCount
	data.Timestamp = time.Now()

	// Send to protocol senders
	if vmcSender != nil {
		_ = vmcSender.Send(data)
	}
	if oscSender != nil {
		_ = oscSender.Send(data)
	}

	// Broadcast to subscribers
	t.mu.RLock()
	subscribers := t.subscribers
	t.mu.RUnlock()

	for _, ch := range subscribers {
		select {
		case ch <- data:
		default:
			// Drop frame if subscriber is slow
		}
	}
}
