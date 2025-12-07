// Package mediapipe provides MediaPipe Holistic integration for facial landmark detection.
package mediapipe

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -L${SRCDIR}/../../cpp_core/bazel-bin -lmediapipe_bridge
#cgo LDFLAGS: -Wl,-rpath,${SRCDIR}/../../cpp_core/bazel-bin
#include "../../cpp_core/mediapipe_bridge.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"

	"gocv.io/x/gocv"
)

// ModelComplexity defines the MediaPipe model complexity level.
type ModelComplexity int

const (
	// ComplexityLite is the fastest, least accurate model (0).
	ComplexityLite ModelComplexity = 0
	// ComplexityFull is balanced performance and accuracy (1).
	ComplexityFull ModelComplexity = 1
	// ComplexityHeavy is the most accurate, slowest model (2).
	ComplexityHeavy ModelComplexity = 2
)

// Config holds MediaPipe Holistic configuration.
type Config struct {
	// ModelComplexity controls the trade-off between speed and accuracy.
	ModelComplexity ModelComplexity
	// MinDetectionConfidence is the minimum confidence [0.0, 1.0] for person detection.
	MinDetectionConfidence float32
	// MinTrackingConfidence is the minimum confidence [0.0, 1.0] for landmark tracking.
	MinTrackingConfidence float32
	// StaticImageMode disables tracking between frames (slower but more accurate).
	StaticImageMode bool
	// SmoothLandmarks applies temporal smoothing (only when StaticImageMode=false).
	SmoothLandmarks bool
}

// DefaultConfig returns a recommended configuration for real-time VTubing.
func DefaultConfig() Config {
	return Config{
		ModelComplexity:        ComplexityFull,
		MinDetectionConfidence: 0.5,
		MinTrackingConfidence:  0.5,
		StaticImageMode:        false,
		SmoothLandmarks:        true,
	}
}

// MediaPipeProcessor implements the Processor interface using MediaPipe Holistic.
type MediaPipeProcessor struct {
	config Config
	handle C.MPHandle // Opaque C++ object handle
	mu     sync.Mutex
	closed bool
}

// NewMediaPipeProcessor creates a new MediaPipe processor instance.
func NewMediaPipeProcessor(config Config) (*MediaPipeProcessor, error) {
	p := &MediaPipeProcessor{
		config: config,
	}

	// Initialize the C++ bridge
	cConfig := C.MPConfig{
		model_complexity:         C.int(config.ModelComplexity),
		min_detection_confidence: C.float(config.MinDetectionConfidence),
		min_tracking_confidence:  C.float(config.MinTrackingConfidence),
		static_image_mode:        C.bool(config.StaticImageMode),
		smooth_landmarks:         C.bool(config.SmoothLandmarks),
		refine_face_landmarks:    C.bool(false), // Not exposed in Go config yet
		enable_segmentation:      C.bool(false), // Not exposed in Go config yet
	}

	p.handle = C.MP_Create(&cConfig)
	if p.handle == nil {
		err := C.MP_GetLastError(p.handle)
		return nil, fmt.Errorf("mediapipe init failed: %s", C.GoString(&err.message[0]))
	}

	return p, nil
}

// Process processes a single frame and returns tracking data.
// The input frame must be in RGB format (gocv.MatTypeCV8UC3).
func (p *MediaPipeProcessor) Process(frame gocv.Mat) (*TrackingData, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, fmt.Errorf("processor is closed")
	}

	if frame.Empty() {
		return nil, fmt.Errorf("empty frame")
	}

	// Ensure RGB format
	if frame.Type() != gocv.MatTypeCV8UC3 {
		return nil, fmt.Errorf("frame must be RGB (CV_8UC3), got type %d", frame.Type())
	}

	width := frame.Cols()
	height := frame.Rows()

	// Get raw pixel data pointer
	pixels, _ := frame.DataPtrUint8()

	// Call C++ bridge to process frame
	var result C.MPResults
	success := C.MP_Process(
		p.handle,
		(*C.uint8_t)(unsafe.Pointer(&pixels[0])),
		C.int(width),
		C.int(height),
		&result,
	)

	if !success {
		err := C.MP_GetLastError(p.handle)
		return nil, fmt.Errorf("mediapipe processing failed: %s", C.GoString(&err.message[0]))
	}

	// Convert C result to Go TrackingData
	data := p.convertResult(&result)

	// Free C++ allocated memory
	C.MP_ReleaseResults(&result)

	return data, nil
}

// convertResult converts MediaPipe C++ results to Go TrackingData structure.
func (p *MediaPipeProcessor) convertResult(result *C.MPResults) *TrackingData {
	data := &TrackingData{
		Timestamp: 0, // TODO: Get actual timestamp from MediaPipe
	}

	// Convert face landmarks (468 or 478 points with refinement)
	if result.face_count > 0 {
		data.Face = &FaceData{
			Landmarks:    make([]Landmark, result.face_count),
			BlendShapes:  make(map[string]float32),
			HeadRotation: Quaternion{X: 0, Y: 0, Z: 0, W: 1}, // Identity, will be computed later
			HeadPosition: Point3D{X: 0, Y: 0, Z: 0},          // Will be computed later
		}

		// Copy landmarks from C array
		landmarks := (*[1 << 16]C.MPLandmark)(unsafe.Pointer(result.face_landmarks))[:result.face_count:result.face_count]
		for i, lm := range landmarks {
			data.Face.Landmarks[i] = Landmark{
				Point: Point3D{
					X: float64(lm.x),
					Y: float64(lm.y),
					Z: float64(lm.z),
				},
				Visibility: float32(lm.visibility),
				Presence:   float32(lm.presence),
			}
		}
	}

	// Convert left hand landmarks (21 points)
	if result.left_hand_count > 0 {
		data.LeftHand = &HandData{
			Landmarks: make([]Landmark, result.left_hand_count),
		}

		landmarks := (*[21]C.MPLandmark)(unsafe.Pointer(result.left_hand_landmarks))[:result.left_hand_count:result.left_hand_count]
		for i, lm := range landmarks {
			data.LeftHand.Landmarks[i] = Landmark{
				Point: Point3D{
					X: float64(lm.x),
					Y: float64(lm.y),
					Z: float64(lm.z),
				},
				Visibility: float32(lm.visibility),
				Presence:   float32(lm.presence),
			}
		}
	}

	// Convert right hand landmarks (21 points)
	if result.right_hand_count > 0 {
		data.RightHand = &HandData{
			Landmarks: make([]Landmark, result.right_hand_count),
		}

		landmarks := (*[21]C.MPLandmark)(unsafe.Pointer(result.right_hand_landmarks))[:result.right_hand_count:result.right_hand_count]
		for i, lm := range landmarks {
			data.RightHand.Landmarks[i] = Landmark{
				Point: Point3D{
					X: float64(lm.x),
					Y: float64(lm.y),
					Z: float64(lm.z),
				},
				Visibility: float32(lm.visibility),
				Presence:   float32(lm.presence),
			}
		}
	}

	// Convert pose landmarks (33 points, but we focus on upper body 0-16)
	if result.pose_count > 0 {
		data.Pose = &PoseData{
			Landmarks: make([]Landmark, result.pose_count),
		}

		landmarks := (*[33]C.MPLandmark)(unsafe.Pointer(result.pose_landmarks))[:result.pose_count:result.pose_count]
		for i, lm := range landmarks {
			data.Pose.Landmarks[i] = Landmark{
				Point: Point3D{
					X: float64(lm.x),
					Y: float64(lm.y),
					Z: float64(lm.z),
				},
				Visibility: float32(lm.visibility),
				Presence:   float32(lm.presence),
			}
		}
	}

	return data
}

// Close releases MediaPipe resources.
func (p *MediaPipeProcessor) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	if p.handle != nil {
		C.MP_Destroy(p.handle)
		p.handle = nil
	}

	p.closed = true
	return nil
}

// TrackingData represents the complete tracking output from MediaPipe.
// This is defined here to avoid circular imports with the main miface package.
type TrackingData struct {
	Timestamp int64     // Frame timestamp in milliseconds
	Face      *FaceData // Facial landmarks and expressions
	LeftHand  *HandData // Left hand landmarks
	RightHand *HandData // Right hand landmarks
	Pose      *PoseData // Body pose landmarks
}

// FaceData contains facial tracking information.
type FaceData struct {
	Landmarks    []Landmark         // 468 face mesh landmarks
	BlendShapes  map[string]float32 // ARKit-style blend shapes (to be computed)
	HeadRotation Quaternion         // Head orientation
	HeadPosition Point3D            // Head position in world space
}

// HandData contains hand tracking information.
type HandData struct {
	Landmarks []Landmark // 21 hand landmarks
}

// PoseData contains body pose tracking information.
type PoseData struct {
	Landmarks []Landmark // 33 pose landmarks (focus on upper body)
}

// Landmark represents a single 3D point with confidence scores.
type Landmark struct {
	Point      Point3D // 3D coordinates
	Visibility float32 // Visibility score [0.0, 1.0]
	Presence   float32 // Presence score [0.0, 1.0]
}

// Point3D represents a 3D point in space.
type Point3D struct {
	X, Y, Z float64
}

// Quaternion represents a rotation in 3D space.
type Quaternion struct {
	X, Y, Z, W float64
}
