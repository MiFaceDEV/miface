package miface

import (
	"math"
	"testing"
)

func TestNewKalmanFilter(t *testing.T) {
	kf := NewKalmanFilter(0.5)
	if kf == nil {
		t.Fatal("expected non-nil filter")
	}
}

func TestKalmanFilterUpdate(t *testing.T) {
	kf := NewKalmanFilter(0.5)

	// First measurement initializes the filter
	result := kf.Update(10.0)
	if result != 10.0 {
		t.Errorf("first update should return measurement, got %f", result)
	}

	// Subsequent updates should be smoothed
	result = kf.Update(11.0)
	if result <= 10.0 || result >= 11.0 {
		t.Errorf("expected smoothed value between 10 and 11, got %f", result)
	}
}

func TestKalmanFilterSmoothing(t *testing.T) {
	// Test that the filter actually smooths noisy data
	kf := NewKalmanFilter(0.3) // Lower factor = more smoothing

	// Simulate noisy signal around 50
	measurements := []float64{50, 52, 48, 51, 49, 50, 53, 47, 51, 49}
	
	var results []float64
	for _, m := range measurements {
		results = append(results, kf.Update(m))
	}

	// Calculate variance of results (should be less than input)
	inputVar := variance(measurements)
	outputVar := variance(results)

	if outputVar >= inputVar {
		t.Errorf("expected output variance (%f) < input variance (%f)", outputVar, inputVar)
	}
}

func TestKalmanFilterReset(t *testing.T) {
	kf := NewKalmanFilter(0.5)
	kf.Update(100.0)
	kf.Update(100.0)

	kf.Reset()

	// After reset, first measurement should be returned directly
	result := kf.Update(50.0)
	if result != 50.0 {
		t.Errorf("after reset, expected 50.0, got %f", result)
	}
}

func TestKalmanFilter3D(t *testing.T) {
	kf := NewKalmanFilter3D(0.5)

	point := Point3D{X: 1, Y: 2, Z: 3}
	result := kf.Update(point)

	if result.X != 1 || result.Y != 2 || result.Z != 3 {
		t.Errorf("first update should return input point, got %+v", result)
	}

	// Second update should be smoothed
	point2 := Point3D{X: 2, Y: 3, Z: 4}
	result2 := kf.Update(point2)

	if result2.X <= 1 || result2.X >= 2 {
		t.Errorf("expected X between 1 and 2, got %f", result2.X)
	}
}

func TestLandmarkSmoother(t *testing.T) {
	smoother := NewLandmarkSmoother(0.5)

	landmarks := []Landmark{
		{Point: Point3D{X: 1, Y: 1, Z: 1}, Visibility: 0.9},
		{Point: Point3D{X: 2, Y: 2, Z: 2}, Visibility: 0.8},
	}

	result := smoother.Smooth(landmarks)
	if len(result) != len(landmarks) {
		t.Errorf("expected %d landmarks, got %d", len(landmarks), len(result))
	}

	// First smoothing should return original values
	if result[0].Point.X != 1 {
		t.Errorf("expected X=1, got %f", result[0].Point.X)
	}

	// Visibility should be preserved
	if result[0].Visibility != 0.9 {
		t.Errorf("expected visibility 0.9, got %f", result[0].Visibility)
	}
}

func TestLandmarkSmootherReset(t *testing.T) {
	smoother := NewLandmarkSmoother(0.5)

	landmarks := []Landmark{
		{Point: Point3D{X: 100, Y: 100, Z: 100}, Visibility: 1.0},
	}

	smoother.Smooth(landmarks)
	smoother.Smooth(landmarks)
	smoother.Reset()

	// After reset, should return original value
	newLandmarks := []Landmark{
		{Point: Point3D{X: 50, Y: 50, Z: 50}, Visibility: 1.0},
	}
	result := smoother.Smooth(newLandmarks)

	if result[0].Point.X != 50 {
		t.Errorf("after reset, expected X=50, got %f", result[0].Point.X)
	}
}

func TestLandmarkSmootherEmpty(t *testing.T) {
	smoother := NewLandmarkSmoother(0.5)

	result := smoother.Smooth(nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}

	result = smoother.Smooth([]Landmark{})
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

// variance calculates the variance of a slice of float64.
func variance(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	var sum float64
	for _, v := range data {
		sum += v
	}
	mean := sum / float64(len(data))

	var sumSq float64
	for _, v := range data {
		diff := v - mean
		sumSq += diff * diff
	}

	return sumSq / float64(len(data))
}

func TestKalmanFilterSmoothingFactors(t *testing.T) {
	// Test different smoothing factors
	tests := []struct {
		factor float64
		desc   string
	}{
		{0.0, "maximum smoothing"},
		{0.5, "medium smoothing"},
		{1.0, "no smoothing"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			kf := NewKalmanFilter(tt.factor)

			// Initialize
			kf.Update(0)

			// Apply step change
			var result float64
			for i := 0; i < 10; i++ {
				result = kf.Update(100)
			}

			// Higher smoothing factor should track faster
			if tt.factor >= 0.9 && math.Abs(result-100) > 10 {
				t.Errorf("high smoothing factor should track quickly, got %f", result)
			}
		})
	}
}
