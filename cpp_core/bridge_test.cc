// bridge_test.cc
// Simple test program for MediaPipe bridge

#include "mediapipe_bridge.h"
#include <iostream>
#include <vector>
#include <cstring>

int main() {
    std::cout << "MediaPipe Bridge Test\n";
    std::cout << "=====================\n\n";

    // Display version
    std::cout << "Version: " << MP_GetVersion() << "\n";
    std::cout << "GPU Available: " << (MP_IsGPUAvailable() ? "Yes" : "No") << "\n\n";

    // Create configuration
    MPConfig config;
    config.model_complexity = 1;  // Full model
    config.min_detection_confidence = 0.5f;
    config.min_tracking_confidence = 0.5f;
    config.static_image_mode = false;
    config.smooth_landmarks = true;
    config.refine_face_landmarks = true;
    config.enable_segmentation = false;

    // Initialize processor
    std::cout << "Initializing processor...\n";
    MPHandle handle = MP_Create(&config);
    if (!handle) {
        MPError error = MP_GetLastError(handle);
        std::cerr << "Failed to create processor: " << error.message << "\n";
        return 1;
    }
    std::cout << "Processor created successfully!\n\n";

    // Create a dummy RGB image (640x480)
    const int width = 640;
    const int height = 480;
    std::vector<uint8_t> dummy_frame(width * height * 3);
    
    // Fill with a gradient pattern (for visual debugging if you save it)
    for (int y = 0; y < height; ++y) {
        for (int x = 0; x < width; ++x) {
            int idx = (y * width + x) * 3;
            dummy_frame[idx + 0] = static_cast<uint8_t>((x * 255) / width);  // R
            dummy_frame[idx + 1] = static_cast<uint8_t>((y * 255) / height); // G
            dummy_frame[idx + 2] = 128;  // B
        }
    }

    // Process the frame
    std::cout << "Processing test frame (" << width << "x" << height << ")...\n";
    MPResults results;
    bool success = MP_Process(handle, dummy_frame.data(), width, height, &results);

    if (!success) {
        MPError error = MP_GetLastError(handle);
        std::cerr << "Processing failed: " << error.message << "\n";
        MP_Destroy(handle);
        return 1;
    }

    // Display results
    std::cout << "\nProcessing Results:\n";
    std::cout << "-------------------\n";
    std::cout << "Processing time: " << results.processing_time_ms << " ms\n";
    std::cout << "Timestamp: " << results.timestamp_ms << " ms\n\n";

    std::cout << "Face detected: " << (results.face_detected ? "Yes" : "No") << "\n";
    if (results.face_detected) {
        std::cout << "  - Face landmarks: " << results.face_count << "\n";
        if (results.face_count > 0) {
            std::cout << "    Example landmark[0]: ("
                      << results.face_landmarks[0].x << ", "
                      << results.face_landmarks[0].y << ", "
                      << results.face_landmarks[0].z << ")\n";
        }
    }

    std::cout << "\nHands detected: " << (results.hands_detected ? "Yes" : "No") << "\n";
    if (results.left_hand_count > 0) {
        std::cout << "  - Left hand landmarks: " << results.left_hand_count << "\n";
    }
    if (results.right_hand_count > 0) {
        std::cout << "  - Right hand landmarks: " << results.right_hand_count << "\n";
    }

    std::cout << "\nPose detected: " << (results.pose_detected ? "Yes" : "No") << "\n";
    if (results.pose_detected) {
        std::cout << "  - Pose landmarks: " << results.pose_count << "\n";
        std::cout << "  - World landmarks: " << results.pose_world_count << "\n";
    }

    // Clean up
    MP_ReleaseResults(&results);
    MP_Destroy(handle);

    std::cout << "\n✅ Test completed successfully!\n";
    return 0;
}
