// mediapipe_bridge.cc
// Implementation of MediaPipe Holistic C wrapper

#include "mediapipe_bridge.h"

#include <cstring>
#include <memory>
#include <string>
#include <vector>

// MediaPipe includes
#include "mediapipe/framework/calculator_framework.h"
#include "mediapipe/framework/formats/image_frame.h"
#include "mediapipe/framework/formats/image_frame_opencv.h"
#include "mediapipe/framework/formats/landmark.pb.h"
#include "mediapipe/framework/port/status.h"
#include "mediapipe/framework/port/parse_text_proto.h"

// OpenCV for image handling
#include <opencv2/core.hpp>
#include <opencv2/imgproc.hpp>

// Graph configuration (defined in holistic_config.cc)
extern const char* kHolisticGraphConfig;

namespace {

// Thread-local error storage
thread_local MPError g_last_error = {0, ""};

void SetError(int code, const std::string& message) {
    g_last_error.code = code;
    strncpy(g_last_error.message, message.c_str(), sizeof(g_last_error.message) - 1);
    g_last_error.message[sizeof(g_last_error.message) - 1] = '\0';
}

void ClearError() {
    g_last_error.code = 0;
    g_last_error.message[0] = '\0';
}

// Convert MediaPipe NormalizedLandmarkList to MPLandmark array
MPLandmark* ConvertLandmarks(
    const mediapipe::NormalizedLandmarkList& landmarks,
    int* out_count
) {
    *out_count = landmarks.landmark_size();
    if (*out_count == 0) {
        return nullptr;
    }

    auto* result = new MPLandmark[*out_count];
    for (int i = 0; i < *out_count; ++i) {
        const auto& lm = landmarks.landmark(i);
        result[i].x = lm.x();
        result[i].y = lm.y();
        result[i].z = lm.z();
        result[i].visibility = lm.has_visibility() ? lm.visibility() : 1.0f;
        result[i].presence = lm.has_presence() ? lm.presence() : 1.0f;
    }
    return result;
}

// Convert MediaPipe LandmarkList (world coordinates) to MPLandmark array
MPLandmark* ConvertWorldLandmarks(
    const mediapipe::LandmarkList& landmarks,
    int* out_count
) {
    *out_count = landmarks.landmark_size();
    if (*out_count == 0) {
        return nullptr;
    }

    auto* result = new MPLandmark[*out_count];
    for (int i = 0; i < *out_count; ++i) {
        const auto& lm = landmarks.landmark(i);
        result[i].x = lm.x();
        result[i].y = lm.y();
        result[i].z = lm.z();
        result[i].visibility = lm.has_visibility() ? lm.visibility() : 1.0f;
        result[i].presence = lm.has_presence() ? lm.presence() : 1.0f;
    }
    return result;
}

} // anonymous namespace

// ============================================================================
// Processor Implementation
// ============================================================================

class MediaPipeProcessor {
public:
    explicit MediaPipeProcessor(const MPConfig* config)
        : config_(*config), frame_count_(0) {
        
        // Parse graph configuration
        mediapipe::CalculatorGraphConfig graph_config;
        if (!mediapipe::ParseTextProto<mediapipe::CalculatorGraphConfig>(
                kHolisticGraphConfig, &graph_config)) {
            throw std::runtime_error("Failed to parse graph config");
        }

        // Initialize calculator graph
        graph_ = std::make_unique<mediapipe::CalculatorGraph>();
        auto status = graph_->Initialize(graph_config);
        if (!status.ok()) {
            throw std::runtime_error("Graph initialization failed: " + status.message());
        }

        // Start the graph
        status = graph_->StartRun({});
        if (!status.ok()) {
            throw std::runtime_error("Failed to start graph: " + status.message());
        }
    }

    ~MediaPipeProcessor() {
        if (graph_) {
            auto status = graph_->CloseAllInputStreams();
            if (status.ok()) {
                graph_->WaitUntilDone();
            }
        }
    }

    bool Process(const uint8_t* pixels, int width, int height, MPResults* results) {
        if (!pixels || !results) {
            SetError(1, "Invalid arguments");
            return false;
        }

        try {
            // Clear previous results
            memset(results, 0, sizeof(MPResults));
            auto start = std::chrono::high_resolution_clock::now();

            // Create OpenCV Mat from raw RGB data (no copy)
            cv::Mat rgb_mat(height, width, CV_8UC3, const_cast<uint8_t*>(pixels));

            // Convert to MediaPipe ImageFrame
            auto image_frame = std::make_unique<mediapipe::ImageFrame>(
                mediapipe::ImageFormat::SRGB,
                width,
                height,
                width * 3,
                const_cast<uint8_t*>(pixels),
                [](uint8_t*) {} // No-op deleter (we don't own the memory)
            );

            // Create packet with timestamp
            auto timestamp = mediapipe::Timestamp(frame_count_++);
            auto packet = mediapipe::MakePacket<mediapipe::ImageFrame>(
                std::move(*image_frame)).At(timestamp);

            // Send to graph
            auto status = graph_->AddPacketToInputStream("input_video", packet);
            if (!status.ok()) {
                SetError(2, "Failed to add packet: " + status.message());
                return false;
            }

            // Fetch results from output streams
            FetchResults(results);

            // Calculate processing time
            auto end = std::chrono::high_resolution_clock::now();
            results->processing_time_ms = 
                std::chrono::duration<float, std::milli>(end - start).count();
            results->timestamp_ms = timestamp.Value() / 1000;

            ClearError();
            return true;

        } catch (const std::exception& e) {
            SetError(3, std::string("Processing error: ") + e.what());
            return false;
        }
    }

private:
    void FetchResults(MPResults* results) {
        // Try to get face landmarks
        mediapipe::Packet face_packet;
        if (graph_->GetOutputStream("face_landmarks")->GetPacket(&face_packet)) {
            const auto& face_landmarks = 
                face_packet.Get<mediapipe::NormalizedLandmarkList>();
            results->face_landmarks = ConvertLandmarks(
                face_landmarks, &results->face_count);
            results->face_detected = results->face_count > 0;
        }

        // Try to get left hand landmarks
        mediapipe::Packet left_hand_packet;
        if (graph_->GetOutputStream("left_hand_landmarks")->GetPacket(&left_hand_packet)) {
            const auto& left_hand_landmarks = 
                left_hand_packet.Get<mediapipe::NormalizedLandmarkList>();
            results->left_hand_landmarks = ConvertLandmarks(
                left_hand_landmarks, &results->left_hand_count);
            results->hands_detected = true;
        }

        // Try to get right hand landmarks
        mediapipe::Packet right_hand_packet;
        if (graph_->GetOutputStream("right_hand_landmarks")->GetPacket(&right_hand_packet)) {
            const auto& right_hand_landmarks = 
                right_hand_packet.Get<mediapipe::NormalizedLandmarkList>();
            results->right_hand_landmarks = ConvertLandmarks(
                right_hand_landmarks, &results->right_hand_count);
            results->hands_detected = true;
        }

        // Try to get pose landmarks
        mediapipe::Packet pose_packet;
        if (graph_->GetOutputStream("pose_landmarks")->GetPacket(&pose_packet)) {
            const auto& pose_landmarks = 
                pose_packet.Get<mediapipe::NormalizedLandmarkList>();
            results->pose_landmarks = ConvertLandmarks(
                pose_landmarks, &results->pose_count);
            results->pose_detected = results->pose_count > 0;
        }

        // Try to get pose world landmarks (3D in meters)
        mediapipe::Packet pose_world_packet;
        if (graph_->GetOutputStream("pose_world_landmarks")->GetPacket(&pose_world_packet)) {
            const auto& pose_world_landmarks = 
                pose_world_packet.Get<mediapipe::LandmarkList>();
            results->pose_world_landmarks = ConvertWorldLandmarks(
                pose_world_landmarks, &results->pose_world_count);
        }
    }

    MPConfig config_;
    std::unique_ptr<mediapipe::CalculatorGraph> graph_;
    int64_t frame_count_;
};

// ============================================================================
// C API Implementation
// ============================================================================

MPHandle MP_Create(const MPConfig* config) {
    if (!config) {
        SetError(10, "Config is null");
        return nullptr;
    }

    try {
        auto* processor = new MediaPipeProcessor(config);
        ClearError();
        return static_cast<MPHandle>(processor);
    } catch (const std::exception& e) {
        SetError(11, std::string("Creation failed: ") + e.what());
        return nullptr;
    }
}

bool MP_Process(
    MPHandle handle,
    const uint8_t* pixels,
    int width,
    int height,
    MPResults* results
) {
    if (!handle) {
        SetError(20, "Invalid handle");
        return false;
    }

    auto* processor = static_cast<MediaPipeProcessor*>(handle);
    return processor->Process(pixels, width, height, results);
}

void MP_ReleaseResults(MPResults* results) {
    if (!results) return;

    delete[] results->face_landmarks;
    delete[] results->left_hand_landmarks;
    delete[] results->right_hand_landmarks;
    delete[] results->pose_landmarks;
    delete[] results->pose_world_landmarks;

    memset(results, 0, sizeof(MPResults));
}

MPError MP_GetLastError(MPHandle handle) {
    (void)handle; // Unused - we use thread-local storage
    return g_last_error;
}

void MP_Destroy(MPHandle handle) {
    if (handle) {
        auto* processor = static_cast<MediaPipeProcessor*>(handle);
        delete processor;
    }
}

const char* MP_GetVersion(void) {
    return "MediaPipe Bridge v1.0.0";
}

bool MP_IsGPUAvailable(void) {
#ifdef MEDIAPIPE_GPU_ENABLED
    return true;
#else
    return false;
#endif
}
