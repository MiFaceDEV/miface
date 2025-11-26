//go:build cgo
// +build cgo

package miface

import (
	"fmt"
	"sync"

	"gocv.io/x/gocv"
)

const (
	// fourccMJPEG is the FourCC code for Motion JPEG codec.
	// MJPEG is widely supported by USB webcams and provides good compression.
	// FourCC codes are 4-byte identifiers: 'MJPG' = 0x47504A4D
	fourccMJPEG = 0x47504A4D
)

// OpenCVCamera implements CameraSource using OpenCV via GoCV.
//
// Implementation notes:
// - Uses V4L2 backend on Linux to avoid GStreamer "Internal data stream error"
// - Sets MJPEG codec explicitly for maximum USB webcam compatibility
// - Applies BGR→RGB conversion since MediaPipe expects RGB24 format
// - Supports horizontal flip (mirror mode) for natural VTubing experience
// - Thread-safe: mu protects all fields and camera operations
type OpenCVCamera struct {
	mu sync.Mutex // Use Mutex instead of RWMutex - all ops modify state

	deviceID int
	width    int
	height   int
	fps      int

	// Mirror enables horizontal flip for VTubing (user sees themselves mirrored)
	mirror bool

	webcam *gocv.VideoCapture
	opened bool
}

// NewOpenCVCamera creates a new OpenCV-based camera source.
// Set mirror=true to flip the image horizontally (typical for VTubing).
func NewOpenCVCamera(mirror bool) *OpenCVCamera {
	return &OpenCVCamera{
		mirror: mirror,
	}
}

// Open initializes the camera with the given configuration.
func (c *OpenCVCamera) Open(deviceID, width, height, fps int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.opened {
		return fmt.Errorf("camera already opened")
	}

	// Open video capture device with V4L2 backend (Linux)
	// This avoids GStreamer issues and provides better compatibility
	webcam, err := gocv.OpenVideoCaptureWithAPI(deviceID, gocv.VideoCaptureV4L2)
	if err != nil {
		return fmt.Errorf("failed to open camera device %d: %w", deviceID, err)
	}

	if !webcam.IsOpened() {
		webcam.Close()
		return fmt.Errorf("camera device %d not found or unavailable", deviceID)
	}

	// Set MJPEG codec for better compatibility with USB webcams
	webcam.Set(gocv.VideoCaptureFOURCC, fourccMJPEG)

	// Configure camera properties after setting backend and codec
	if width > 0 {
		webcam.Set(gocv.VideoCaptureFrameWidth, float64(width))
	}
	if height > 0 {
		webcam.Set(gocv.VideoCaptureFrameHeight, float64(height))
	}
	if fps > 0 {
		webcam.Set(gocv.VideoCaptureFPS, float64(fps))
	}

	// Verify actual resolution
	actualWidth := webcam.Get(gocv.VideoCaptureFrameWidth)
	actualHeight := webcam.Get(gocv.VideoCaptureFrameHeight)
	actualFPS := webcam.Get(gocv.VideoCaptureFPS)

	c.deviceID = deviceID
	c.width = int(actualWidth)
	c.height = int(actualHeight)
	c.fps = int(actualFPS)
	c.webcam = webcam
	c.opened = true

	// Warm up camera - read and discard first frame
	// Some cameras need a moment to initialize
	warmupMat := gocv.NewMat()
	c.webcam.Read(&warmupMat)
	warmupMat.Close()

	return nil
}

// Read captures a single frame from the camera.
// Returns the frame data as RGB24 bytes, along with width and height.
func (c *OpenCVCamera) Read() ([]byte, int, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.opened {
		return nil, 0, 0, fmt.Errorf("camera not opened")
	}

	// Create fresh Mat for this read (thread-safe)
	mat := gocv.NewMat()
	defer mat.Close()

	// Read frame into Mat
	if ok := c.webcam.Read(&mat); !ok {
		return nil, 0, 0, fmt.Errorf("failed to read frame from camera")
	}

	if mat.Empty() {
		return nil, 0, 0, fmt.Errorf("captured frame is empty")
	}

	// Apply horizontal flip if mirror mode enabled
	if c.mirror {
		gocv.Flip(mat, &mat, 1) //nolint:errcheck // gocv.Flip doesn't return error
	}

	// Convert BGR to RGB (OpenCV uses BGR by default)
	rgbMat := gocv.NewMat()
	defer rgbMat.Close()
	gocv.CvtColor(mat, &rgbMat, gocv.ColorBGRToRGB) //nolint:errcheck // gocv.CvtColor doesn't return error

	// Get frame dimensions
	width := rgbMat.Cols()
	height := rgbMat.Rows()

	// Convert Mat to byte slice
	// MediaPipe expects continuous RGB24 data
	frameData := rgbMat.ToBytes()

	return frameData, width, height, nil
}

// ReadMat captures a frame and returns it as a gocv.Mat for preview.
// The returned Mat should be closed by the caller.
// This is separate from Read() to avoid unnecessary conversions for preview.
func (c *OpenCVCamera) ReadMat() (gocv.Mat, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.opened {
		return gocv.NewMat(), fmt.Errorf("camera not opened")
	}

	// Create fresh Mat for this read (thread-safe)
	mat := gocv.NewMat()
	defer mat.Close()

	// Read frame into Mat
	if ok := c.webcam.Read(&mat); !ok {
		return gocv.NewMat(), fmt.Errorf("failed to read frame from camera")
	}

	if mat.Empty() {
		return gocv.NewMat(), fmt.Errorf("captured frame is empty")
	}

	// Clone for return value
	result := mat.Clone()

	// Apply horizontal flip if mirror mode enabled
	if c.mirror {
		gocv.Flip(result, &result, 1) //nolint:errcheck // gocv.Flip doesn't return error
	}

	return result, nil
}

// Close releases camera resources.
func (c *OpenCVCamera) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.opened {
		return nil
	}

	if c.webcam != nil {
		if err := c.webcam.Close(); err != nil {
			c.opened = false
			return fmt.Errorf("closing webcam: %w", err)
		}
	}

	c.opened = false
	return nil
}

// SetMirror enables or disables horizontal flip.
// Can be called while camera is running.
func (c *OpenCVCamera) SetMirror(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mirror = enabled
}

// IsMirror returns whether horizontal flip is enabled.
func (c *OpenCVCamera) IsMirror() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.mirror
}

// GetActualResolution returns the actual configured resolution.
// This may differ from requested resolution if the camera doesn't support it.
func (c *OpenCVCamera) GetActualResolution() (width, height int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.width, c.height
}

// GetActualFPS returns the actual configured frame rate.
func (c *OpenCVCamera) GetActualFPS() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.fps
}

// EnumerateCameras attempts to detect available camera devices.
// Returns a list of device IDs that can be opened.
// This is a best-effort function and may not work on all systems.
func EnumerateCameras(maxDevices int) []int {
	var devices []int

	if maxDevices <= 0 {
		maxDevices = 10 // Default: try first 10 devices
	}

	for i := 0; i < maxDevices; i++ {
		// Use V4L2 backend for consistency with Open()
		cam, err := gocv.OpenVideoCaptureWithAPI(i, gocv.VideoCaptureV4L2)
		if err != nil {
			continue
		}
		if cam.IsOpened() {
			devices = append(devices, i)
			cam.Close()
		} else {
			cam.Close()
		}
	}

	return devices
}
