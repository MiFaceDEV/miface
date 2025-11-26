package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Camera.DeviceID != 0 {
		t.Errorf("expected DeviceID 0, got %d", cfg.Camera.DeviceID)
	}
	if cfg.Camera.Width != 1280 {
		t.Errorf("expected Width 1280, got %d", cfg.Camera.Width)
	}
	if cfg.Camera.Height != 720 {
		t.Errorf("expected Height 720, got %d", cfg.Camera.Height)
	}
	if cfg.Camera.FPS != 30 {
		t.Errorf("expected FPS 30, got %d", cfg.Camera.FPS)
	}
	if !cfg.Tracking.EnableFace {
		t.Error("expected EnableFace to be true")
	}
	if !cfg.Tracking.EnableHands {
		t.Error("expected EnableHands to be true")
	}
	if !cfg.Tracking.EnablePose {
		t.Error("expected EnablePose to be true")
	}
	if cfg.Tracking.SmoothingFactor != 0.5 {
		t.Errorf("expected SmoothingFactor 0.5, got %f", cfg.Tracking.SmoothingFactor)
	}
	if !cfg.VMC.Enabled {
		t.Error("expected VMC.Enabled to be true")
	}
	if cfg.VMC.Port != 39539 {
		t.Errorf("expected VMC.Port 39539, got %d", cfg.VMC.Port)
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("unexpected error for non-existent file: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected default config for non-existent file")
	}
}

func TestLoad_ValidFile(t *testing.T) {
	content := `
[camera]
device_id = 1
width = 1920
height = 1080
fps = 60

[tracking]
enable_face = false
enable_hands = true
enable_pose = false
smoothing_factor = 0.8

[vmc]
enabled = false
address = "192.168.1.100"
port = 39540
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Camera.DeviceID != 1 {
		t.Errorf("expected DeviceID 1, got %d", cfg.Camera.DeviceID)
	}
	if cfg.Camera.Width != 1920 {
		t.Errorf("expected Width 1920, got %d", cfg.Camera.Width)
	}
	if cfg.Camera.Height != 1080 {
		t.Errorf("expected Height 1080, got %d", cfg.Camera.Height)
	}
	if cfg.Camera.FPS != 60 {
		t.Errorf("expected FPS 60, got %d", cfg.Camera.FPS)
	}
	if cfg.Tracking.EnableFace {
		t.Error("expected EnableFace to be false")
	}
	if cfg.Tracking.SmoothingFactor != 0.8 {
		t.Errorf("expected SmoothingFactor 0.8, got %f", cfg.Tracking.SmoothingFactor)
	}
	if cfg.VMC.Enabled {
		t.Error("expected VMC.Enabled to be false")
	}
	if cfg.VMC.Address != "192.168.1.100" {
		t.Errorf("expected VMC.Address 192.168.1.100, got %s", cfg.VMC.Address)
	}
	if cfg.VMC.Port != 39540 {
		t.Errorf("expected VMC.Port 39540, got %d", cfg.VMC.Port)
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.toml")
	if err := os.WriteFile(path, []byte("invalid [ toml"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

func TestValidate_InvalidWidth(t *testing.T) {
	cfg := Default()
	cfg.Camera.Width = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid width")
	}
}

func TestValidate_InvalidHeight(t *testing.T) {
	cfg := Default()
	cfg.Camera.Height = -1
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid height")
	}
}

func TestValidate_InvalidFPS(t *testing.T) {
	cfg := Default()
	cfg.Camera.FPS = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid FPS")
	}
}

func TestValidate_InvalidSmoothingFactor(t *testing.T) {
	cfg := Default()
	cfg.Tracking.SmoothingFactor = 1.5
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for smoothing factor > 1")
	}

	cfg.Tracking.SmoothingFactor = -0.1
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for smoothing factor < 0")
	}
}

func TestValidate_InvalidVMCPort(t *testing.T) {
	cfg := Default()
	cfg.VMC.Port = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for VMC port 0")
	}

	cfg.VMC.Port = 70000
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for VMC port > 65535")
	}
}
