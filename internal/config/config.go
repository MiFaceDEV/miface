// Package config provides TOML configuration loading for MiFace.
//
// The configuration file supports the following structure:
//
//	[camera]
//	device_id = 0
//	width = 1280
//	height = 720
//	fps = 30
//
//	[tracking]
//	enable_face = true
//	enable_hands = true
//	enable_pose = true
//	smoothing_factor = 0.5
//
//	[vmc]
//	enabled = true
//	address = "127.0.0.1"
//	port = 39539
//
// Example usage:
//
//	cfg, err := config.Load("config.toml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Camera device: %d\n", cfg.Camera.DeviceID)
package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents the complete configuration for MiFace.
type Config struct {
	Camera   CameraConfig   `toml:"camera"`
	Tracking TrackingConfig `toml:"tracking"`
	VMC      VMCConfig      `toml:"vmc"`
}

// CameraConfig holds webcam capture settings.
type CameraConfig struct {
	// DeviceID is the camera device index (default: 0).
	DeviceID int `toml:"device_id"`
	// Width is the capture width in pixels (default: 1280).
	Width int `toml:"width"`
	// Height is the capture height in pixels (default: 720).
	Height int `toml:"height"`
	// FPS is the target frame rate (default: 30).
	FPS int `toml:"fps"`
}

// TrackingConfig holds face/body tracking settings.
type TrackingConfig struct {
	// EnableFace enables face landmark tracking (default: true).
	EnableFace bool `toml:"enable_face"`
	// EnableHands enables hand landmark tracking (default: true).
	EnableHands bool `toml:"enable_hands"`
	// EnablePose enables pose/body tracking (default: true).
	EnablePose bool `toml:"enable_pose"`
	// SmoothingFactor controls Kalman filter smoothing (0.0-1.0, default: 0.5).
	SmoothingFactor float64 `toml:"smoothing_factor"`
}

// VMCConfig holds VMC (Virtual Motion Capture) protocol sender settings.
// VMC uses the OSC protocol for communication.
type VMCConfig struct {
	// Enabled enables VMC protocol output (default: true).
	Enabled bool `toml:"enabled"`
	// Address is the destination IP address (default: "127.0.0.1").
	Address string `toml:"address"`
	// Port is the destination UDP port (default: 39539).
	Port int `toml:"port"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Camera: CameraConfig{
			DeviceID: 0,
			Width:    1280,
			Height:   720,
			FPS:      30,
		},
		Tracking: TrackingConfig{
			EnableFace:      true,
			EnableHands:     true,
			EnablePose:      true,
			SmoothingFactor: 0.5,
		},
		VMC: VMCConfig{
			Enabled: true,
			Address: "127.0.0.1",
			Port:    39539,
		},
	}
}

// Load reads and parses a TOML configuration file.
// If the file does not exist, it returns the default configuration.
func Load(path string) (*Config, error) {
	cfg := Default()

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return cfg, nil
}

// Validate checks the configuration for invalid values.
func (c *Config) Validate() error {
	if c.Camera.Width <= 0 {
		return fmt.Errorf("camera width must be positive, got %d", c.Camera.Width)
	}
	if c.Camera.Height <= 0 {
		return fmt.Errorf("camera height must be positive, got %d", c.Camera.Height)
	}
	if c.Camera.FPS <= 0 {
		return fmt.Errorf("camera FPS must be positive, got %d", c.Camera.FPS)
	}
	if c.Tracking.SmoothingFactor < 0 || c.Tracking.SmoothingFactor > 1 {
		return fmt.Errorf("smoothing factor must be between 0 and 1, got %f", c.Tracking.SmoothingFactor)
	}
	if c.VMC.Port <= 0 || c.VMC.Port > 65535 {
		return fmt.Errorf("VMC port must be between 1 and 65535, got %d", c.VMC.Port)
	}
	return nil
}
