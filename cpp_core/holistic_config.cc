// holistic_config.cc
// MediaPipe Holistic graph configuration

// This is the graph configuration for MediaPipe Holistic.
// It defines the processing pipeline: face mesh + hands + pose tracking.
//
// Reference: mediapipe/graphs/holistic_tracking/holistic_tracking_cpu.pbtxt

const char* kHolisticGraphConfig = R"pb(
# MediaPipe Holistic Tracking Graph (CPU)
# Inputs: "input_video" (ImageFrame)
# Outputs: "face_landmarks", "pose_landmarks", "left_hand_landmarks", "right_hand_landmarks"

input_stream: "input_video"

# Outputs
output_stream: "face_landmarks"
output_stream: "pose_landmarks"
output_stream: "pose_world_landmarks"
output_stream: "left_hand_landmarks"
output_stream: "right_hand_landmarks"

# Face detection for initial frame
node {
  calculator: "FlowLimiterCalculator"
  input_stream: "input_video"
  input_stream: "FINISHED:face_landmarks"
  input_stream_info: {
    tag_index: "FINISHED"
    back_edge: true
  }
  output_stream: "throttled_input_video"
}

# Face landmark detection
node {
  calculator: "FaceLandmarkCpu"
  input_stream: "IMAGE:throttled_input_video"
  output_stream: "LANDMARKS:face_landmarks"
}

# Pose detection
node {
  calculator: "PoseLandmarkCpu"
  input_stream: "IMAGE:throttled_input_video"
  output_stream: "LANDMARKS:pose_landmarks"
  output_stream: "WORLD_LANDMARKS:pose_world_landmarks"
}

# Hand tracking from pose wrist landmarks
node {
  calculator: "HandLandmarkTrackingCpu"
  input_stream: "IMAGE:throttled_input_video"
  input_stream: "ROI_FROM_POSE:pose_landmarks"
  output_stream: "LEFT_HAND_LANDMARKS:left_hand_landmarks"
  output_stream: "RIGHT_HAND_LANDMARKS:right_hand_landmarks"
}
)pb";

// Note: The above is a SIMPLIFIED placeholder configuration.
// In production, you need to use the actual MediaPipe graph configuration
// from: mediapipe/graphs/holistic_tracking/holistic_tracking_cpu.pbtxt
//
// The real config is much more complex and includes:
// - Image preprocessing
// - Face mesh with attention
// - Pose detection and tracking
// - Hand ROI cropping from pose
// - Hand landmark tracking
// - Result smoothing and filtering
//
// To use the official config:
// 1. Copy the .pbtxt file content here
// 2. Or load it from a file at runtime
// 3. Make sure all calculator dependencies are linked in BUILD file
