// mediapipe_bridge.h
// Pure C interface for MediaPipe Holistic processing
// This header is consumed by Go via CGO

#ifndef MEDIAPIPE_BRIDGE_H
#define MEDIAPIPE_BRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdbool.h>
#include <stdint.h>

// Opaque handle to processor instance
typedef void* MPHandle;

// Configuration options
typedef struct {
    int model_complexity;           // 0=Lite, 1=Full, 2=Heavy
    float min_detection_confidence; // 0.0 to 1.0
    float min_tracking_confidence;  // 0.0 to 1.0
    bool static_image_mode;         // false for video
    bool smooth_landmarks;          // temporal smoothing
    bool refine_face_landmarks;     // enable face mesh refinement
    bool enable_segmentation;       // enable person segmentation
} MPConfig;

// Single 3D landmark point
typedef struct {
    float x;          // Normalized [0, 1]
    float y;          // Normalized [0, 1]
    float z;          // Depth (meters from camera)
    float visibility; // [0, 1] - is landmark visible
    float presence;   // [0, 1] - is landmark present
} MPLandmark;

// Processing results
typedef struct {
    // Face mesh landmarks (468 or 478 with refinement)
    MPLandmark* face_landmarks;
    int face_count;

    // Left hand landmarks (21)
    MPLandmark* left_hand_landmarks;
    int left_hand_count;

    // Right hand landmarks (21)
    MPLandmark* right_hand_landmarks;
    int right_hand_count;

    // Pose landmarks (33)
    MPLandmark* pose_landmarks;
    int pose_count;

    // World landmarks (3D coordinates in meters)
    MPLandmark* pose_world_landmarks;
    int pose_world_count;

    // Processing metadata
    int64_t timestamp_ms;
    float processing_time_ms;
    bool face_detected;
    bool hands_detected;
    bool pose_detected;
} MPResults;

// Error handling
typedef struct {
    int code;              // 0 = success, non-zero = error
    char message[256];     // Error message
} MPError;

// ============================================================================
// API Functions
// ============================================================================

// Initialize MediaPipe processor
// Returns handle on success, NULL on failure
// Call MP_GetLastError() to get error details
MPHandle MP_Create(const MPConfig* config);

// Process RGB image frame
// pixels: RGB24 byte array (width * height * 3)
// width, height: image dimensions
// results: output structure (caller must call MP_ReleaseResults after use)
// Returns true on success, false on failure
bool MP_Process(
    MPHandle handle,
    const uint8_t* pixels,
    int width,
    int height,
    MPResults* results
);

// Release memory allocated for results
void MP_ReleaseResults(MPResults* results);

// Get last error details
MPError MP_GetLastError(MPHandle handle);

// Destroy processor and free resources
void MP_Destroy(MPHandle handle);

// Get library version info
const char* MP_GetVersion(void);

// Check if GPU acceleration is available
bool MP_IsGPUAvailable(void);

#ifdef __cplusplus
}
#endif

#endif // MEDIAPIPE_BRIDGE_H
