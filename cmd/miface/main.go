// Package main provides the CLI wrapper for MiFace.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MiFaceDEV/miface/internal/config"
	"github.com/MiFaceDEV/miface/pkg/miface"
)

var (
	version = "0.1.0"
)

func main() {
	// Command line flags
	configPath := flag.String("config", "", "Path to TOML configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	vmcAddr := flag.String("vmc-addr", "", "VMC target address (overrides config)")
	vmcPort := flag.Int("vmc-port", 0, "VMC target port (overrides config)")
	cameraID := flag.Int("camera", -1, "Camera device ID (overrides config)")
	vrmPath := flag.String("vrm", "", "Path to VRM file for calibration")
	noMirror := flag.Bool("no-mirror", false, "Disable horizontal flip (mirror mode)")
	preview := flag.Bool("preview", false, "Show camera preview window (debug mode)")
	verbose := flag.Bool("verbose", false, "Enable verbose output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "MiFace - Real-time facial and upper body tracking for VTubers\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s                          # Run with default settings\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -config config.toml      # Run with custom config\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -preview                 # Show camera preview window\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -vmc-port 39540          # Override VMC port\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -vrm model.vrm           # Calibrate with VRM model\n", os.Args[0])
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("MiFace version %s\n", version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Apply command line overrides
	if *vmcAddr != "" {
		cfg.VMC.Address = *vmcAddr
	}
	if *vmcPort > 0 {
		cfg.VMC.Port = *vmcPort
	}
	if *cameraID >= 0 {
		cfg.Camera.DeviceID = *cameraID
	}

	if *verbose {
		log.Printf("Configuration:")
		log.Printf("  Camera: device=%d, %dx%d@%dfps",
			cfg.Camera.DeviceID, cfg.Camera.Width, cfg.Camera.Height, cfg.Camera.FPS)
		log.Printf("  Tracking: face=%v, hands=%v, pose=%v, smoothing=%.2f",
			cfg.Tracking.EnableFace, cfg.Tracking.EnableHands,
			cfg.Tracking.EnablePose, cfg.Tracking.SmoothingFactor)
		log.Printf("  VMC: enabled=%v, %s:%d",
			cfg.VMC.Enabled, cfg.VMC.Address, cfg.VMC.Port)
	}

	// Load VRM for calibration if provided
	if *vrmPath != "" {
		skeleton, err := miface.LoadVRMSkeleton(*vrmPath)
		if err != nil {
			log.Fatalf("Failed to load VRM file: %v", err)
		}

		props := skeleton.GetProportions()
		if *verbose {
			log.Printf("VRM skeleton loaded:")
			log.Printf("  Bones: %d", len(skeleton.Bones))
			log.Printf("  Human bones: %d", len(skeleton.HumanBones))
			log.Printf("  Arm span: %.3f", skeleton.ArmSpan)
			log.Printf("  Height: %.3f", skeleton.Height)
			log.Printf("  Head size: %.3f", skeleton.HeadSize)
			log.Printf("  Proportions:")
			log.Printf("    Arm length: %.3f", props.ArmLength)
			log.Printf("    Upper arm: %.3f", props.UpperArmLength)
			log.Printf("    Lower arm: %.3f", props.LowerArmLength)
			log.Printf("    Shoulder width: %.3f", props.ShoulderWidth)
		} else {
			log.Printf("VRM calibration loaded: %d bones, height=%.2f",
				len(skeleton.HumanBones), skeleton.Height)
		}
	}

	// Create tracker
	tracker, err := miface.NewTracker(cfg)
	if err != nil {
		log.Fatalf("Failed to create tracker: %v", err)
	}
	defer tracker.Close()

	// Set up OpenCV camera
	mirror := !*noMirror // Mirror enabled by default for VTubing
	camera := miface.NewOpenCVCamera(mirror)
	if err := camera.Open(cfg.Camera.DeviceID, cfg.Camera.Width, cfg.Camera.Height, cfg.Camera.FPS); err != nil {
		log.Fatalf("Failed to open camera: %v", err)
	}
	if err := tracker.SetCameraSource(camera); err != nil {
		log.Fatalf("Failed to set camera source: %v", err)
	}

	// Log actual camera settings
	actualWidth, actualHeight := camera.GetActualResolution()
	actualFPS := camera.GetActualFPS()
	if *verbose {
		log.Printf("Camera opened: device=%d, resolution=%dx%d, fps=%d, mirror=%v",
			cfg.Camera.DeviceID, actualWidth, actualHeight, actualFPS, mirror)
	} else {
		log.Printf("Camera opened: %dx%d@%dfps", actualWidth, actualHeight, actualFPS)
	}

	// Set up preview window if enabled
	if *preview {
		previewWindow := miface.NewPreviewWindow("MiFace Preview")
		if err := tracker.SetPreviewWindow(previewWindow); err != nil {
			log.Fatalf("Failed to set preview window: %v", err)
		}
		log.Println("Preview window enabled")
	}

	// Set up VMC sender if enabled
	if cfg.VMC.Enabled {
		vmcSender, err := miface.NewVMCSender(cfg.VMC.Address, cfg.VMC.Port)
		if err != nil {
			log.Fatalf("Failed to create VMC sender: %v", err)
		}
		if err := tracker.SetVMCSender(vmcSender); err != nil {
			log.Fatalf("Failed to set VMC sender: %v", err)
		}
		log.Printf("VMC sender configured: %s:%d", cfg.VMC.Address, cfg.VMC.Port)
	}

	// Subscribe to tracking data for verbose output
	var dataCh <-chan *miface.TrackingData
	if *verbose {
		dataCh = tracker.Subscribe()
	}

	// Start tracking
	if err := tracker.Start(); err != nil {
		log.Fatalf("Failed to start tracker: %v", err)
	}
	log.Println("Tracking started. Press Ctrl+C to stop.")

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Main loop
	if *verbose && dataCh != nil {
		// Verbose mode: log tracking data
		frameCount := uint64(0)
		for {
			select {
			case sig := <-sigCh:
				log.Printf("Received signal %v, shutting down...", sig)
				return

			case data, ok := <-dataCh:
				if !ok {
					return
				}
				frameCount++
				if frameCount%30 == 0 { // Log every 30 frames (~1 second at 30fps)
					log.Printf("Frame %d: face=%v, leftHand=%v, rightHand=%v",
						data.FrameNumber,
						data.Face != nil,
						data.LeftHand != nil,
						data.RightHand != nil)
				}
			}
		}
	} else {
		// Non-verbose mode: just wait for shutdown signal
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down...", sig)
	}
}
