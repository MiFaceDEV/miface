package miface

import (
	"sync"
)

// KalmanFilter implements a simple 1D Kalman filter for landmark smoothing.
// It provides smooth tracking by reducing jitter while maintaining responsiveness.
type KalmanFilter struct {
	mu sync.Mutex

	// State estimate
	x float64
	// Estimate uncertainty
	p float64
	// Process noise (how much we expect the system to change)
	q float64
	// Measurement noise (how noisy the measurements are)
	r float64
	// Initialized flag
	initialized bool
}

// NewKalmanFilter creates a new Kalman filter with the given smoothing factor.
// smoothingFactor controls the trade-off between smoothness and responsiveness:
//   - 0.0 = maximum smoothing (slow response)
//   - 1.0 = no smoothing (instant response)
func NewKalmanFilter(smoothingFactor float64) *KalmanFilter {
	// Map smoothing factor to process/measurement noise ratio
	// Lower smoothing factor = higher R (more trust in prediction)
	// Higher smoothing factor = lower R (more trust in measurement)
	q := 0.1                                    // Process noise
	r := 1.0 - smoothingFactor*0.9 + 0.1        // Measurement noise (0.1 to 1.0)

	return &KalmanFilter{
		p: 1.0, // Initial uncertainty
		q: q,
		r: r,
	}
}

// Update processes a new measurement and returns the filtered value.
func (kf *KalmanFilter) Update(measurement float64) float64 {
	kf.mu.Lock()
	defer kf.mu.Unlock()

	if !kf.initialized {
		kf.x = measurement
		kf.initialized = true
		return measurement
	}

	// Prediction step
	// x_pred = x (assuming constant velocity model with no control input)
	// p_pred = p + q
	pPred := kf.p + kf.q

	// Update step
	// Kalman gain: k = p_pred / (p_pred + r)
	k := pPred / (pPred + kf.r)

	// State update: x = x_pred + k * (measurement - x_pred)
	kf.x = kf.x + k*(measurement-kf.x)

	// Covariance update: p = (1 - k) * p_pred
	kf.p = (1 - k) * pPred

	return kf.x
}

// Reset clears the filter state.
func (kf *KalmanFilter) Reset() {
	kf.mu.Lock()
	defer kf.mu.Unlock()

	kf.x = 0
	kf.p = 1.0
	kf.initialized = false
}

// State returns the current state estimate.
func (kf *KalmanFilter) State() float64 {
	kf.mu.Lock()
	defer kf.mu.Unlock()
	return kf.x
}

// KalmanFilter3D applies Kalman filtering to 3D points.
type KalmanFilter3D struct {
	x, y, z *KalmanFilter
}

// NewKalmanFilter3D creates a new 3D Kalman filter.
func NewKalmanFilter3D(smoothingFactor float64) *KalmanFilter3D {
	return &KalmanFilter3D{
		x: NewKalmanFilter(smoothingFactor),
		y: NewKalmanFilter(smoothingFactor),
		z: NewKalmanFilter(smoothingFactor),
	}
}

// Update processes a new 3D measurement and returns the filtered point.
func (kf *KalmanFilter3D) Update(point Point3D) Point3D {
	return Point3D{
		X: kf.x.Update(point.X),
		Y: kf.y.Update(point.Y),
		Z: kf.z.Update(point.Z),
	}
}

// Reset clears all filter states.
func (kf *KalmanFilter3D) Reset() {
	kf.x.Reset()
	kf.y.Reset()
	kf.z.Reset()
}

// LandmarkSmoother manages Kalman filters for a set of landmarks.
type LandmarkSmoother struct {
	mu      sync.RWMutex
	filters map[int]*KalmanFilter3D
	factor  float64
}

// NewLandmarkSmoother creates a new landmark smoother with the given smoothing factor.
func NewLandmarkSmoother(smoothingFactor float64) *LandmarkSmoother {
	return &LandmarkSmoother{
		filters: make(map[int]*KalmanFilter3D),
		factor:  smoothingFactor,
	}
}

// Smooth applies Kalman filtering to a slice of landmarks.
func (ls *LandmarkSmoother) Smooth(landmarks []Landmark) []Landmark {
	if len(landmarks) == 0 {
		return landmarks
	}

	ls.mu.Lock()
	defer ls.mu.Unlock()

	result := make([]Landmark, len(landmarks))
	for i, lm := range landmarks {
		filter, ok := ls.filters[i]
		if !ok {
			filter = NewKalmanFilter3D(ls.factor)
			ls.filters[i] = filter
		}

		result[i] = Landmark{
			Point:      filter.Update(lm.Point),
			Visibility: lm.Visibility,
		}
	}

	return result
}

// Reset clears all landmark filters.
func (ls *LandmarkSmoother) Reset() {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	for _, f := range ls.filters {
		f.Reset()
	}
}
