# Getting the Real MediaPipe Holistic Graph Configuration

## The Problem

The `holistic_config.cc` file currently contains a **PLACEHOLDER** graph configuration.
You need to replace it with the actual MediaPipe Holistic graph definition.

## Option 1: Copy from MediaPipe Repository (Recommended)

### Step 1: Locate the Graph File

After cloning MediaPipe:
```bash
cd path/to/mediapipe
find . -name "holistic_tracking_cpu.pbtxt"
```

File location:
```
mediapipe/graphs/holistic_tracking/holistic_tracking_cpu.pbtxt
```

### Step 2: View the Content

```bash
cat mediapipe/graphs/holistic_tracking/holistic_tracking_cpu.pbtxt
```

### Step 3: Copy to C++ String

Open `cpp_core/holistic_config.cc` and replace the content of `kHolisticGraphConfig`:

```cpp
const char* kHolisticGraphConfig = R"pb(
// Paste the ENTIRE content of holistic_tracking_cpu.pbtxt here
// It should start with comments and input_stream definitions
// It should have many node { } blocks defining the pipeline
)pb";
```

## Option 2: Load from File at Runtime

If you prefer dynamic loading, modify `mediapipe_bridge.cc`:

```cpp
#include <fstream>
#include <sstream>

MediaPipeProcessor::MediaPipeProcessor(const MPConfig* config)
    : config_(*config), frame_count_(0) {
    
    // Load graph from file
    std::string graph_path = "/path/to/holistic_tracking_cpu.pbtxt";
    std::ifstream graph_file(graph_path);
    if (!graph_file.is_open()) {
        throw std::runtime_error("Cannot open graph file: " + graph_path);
    }
    
    std::stringstream buffer;
    buffer << graph_file.rdbuf();
    std::string graph_str = buffer.str();
    
    // Parse graph
    mediapipe::CalculatorGraphConfig graph_config;
    if (!mediapipe::ParseTextProto<mediapipe::CalculatorGraphConfig>(
            graph_str, &graph_config)) {
        throw std::runtime_error("Failed to parse graph config");
    }
    
    // ... rest of initialization
}
```

## Option 3: Use MediaPipe Python to Generate

If you have MediaPipe Python installed:

```python
import mediapipe as mp

# This will show you the graph structure
mp_holistic = mp.solutions.holistic
print(mp_holistic.HOLISTIC_GRAPH)
```

However, the Python API abstracts away the actual graph definition,
so **Option 1 is strongly recommended**.

## What the Real Graph Contains

The actual graph is ~200-300 lines and includes:

1. **Input Streams**
   - `input_video` - RGB image frames

2. **Face Detection & Mesh**
   - Face detector for initial detection
   - Face mesh with 468 landmarks
   - Optional refinement for lips and eyes

3. **Pose Detection**
   - Pose detector (BlazePose)
   - 33 pose landmarks
   - World coordinates output

4. **Hand Tracking**
   - Hand ROI extraction from pose wrists
   - Left/right hand landmark detection
   - 21 landmarks per hand

5. **Output Streams**
   - `face_landmarks` - NormalizedLandmarkList
   - `pose_landmarks` - NormalizedLandmarkList
   - `pose_world_landmarks` - LandmarkList
   - `left_hand_landmarks` - NormalizedLandmarkList
   - `right_hand_landmarks` - NormalizedLandmarkList

6. **Calculators Used**
   - `ImageTransformationCalculator`
   - `FaceLandmarksFromPoseCalculator`
   - `HandLandmarksFromPoseCalculator`
   - `FlowLimiterCalculator`
   - And many more...

## Example: First 50 Lines of Real Config

Here's what the beginning looks like:

```protobuf
# MediaPipe graph that performs holistic landmark detection with TensorFlow Lite on CPU.

# CPU image. (ImageFrame)
input_stream: "input_video"

# Detected face landmarks. (NormalizedLandmarkList)
output_stream: "face_landmarks"
# Detected pose landmarks. (NormalizedLandmarkList)
output_stream: "pose_landmarks"
# Detected pose world landmarks. (LandmarkList)
output_stream: "pose_world_landmarks"
# Detected left hand landmarks. (NormalizedLandmarkList)
output_stream: "left_hand_landmarks"
# Detected right hand landmarks. (NormalizedLandmarkList)
output_stream: "right_hand_landmarks"

# Throttles the images flowing downstream for flow control. It passes through
# the very first incoming image unaltered, and waits for downstream nodes
# (calculators and subgraphs) in the graph to finish their tasks before it
# passes through another image. All images that come in while waiting are
# dropped, limiting the number of in-flight images in most part of the graph to
# 1. This prevents the downstream nodes from queuing up incoming images and data
# excessively, which leads to increased latency and memory usage, unwanted in
# real-time mobile applications.
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

# ... continues for 200+ more lines
```

## Verification

After adding the real config, rebuild and test:

```bash
cd cpp_core
bazel clean
./build.sh
./bazel-bin/bridge_test
```

**Success looks like:**
```
MediaPipe Bridge Test
=====================
Version: MediaPipe Bridge v1.0.0
GPU Available: No

Initializing processor...
Processor created successfully!

Processing test frame (640x480)...

Processing Results:
-------------------
Processing time: 45.2 ms
Timestamp: 0 ms

Face detected: Yes
  - Face landmarks: 468
  - Example landmark[0]: (0.501, 0.498, -0.023)
...
```

## Important Notes

⚠️ The placeholder config in `holistic_config.cc` **WILL NOT WORK** as-is.
It's intentionally incomplete to avoid legal/licensing issues with copying Google's config.

✅ Once you have the real config, everything else is ready to go!

✅ The C++ code is designed to work with the official MediaPipe graph without modification.

## GPU Version

For GPU support, use:
```
mediapipe/graphs/holistic_tracking/holistic_tracking_gpu.pbtxt
```

And build with:
```bash
./build.sh gpu
```

## Questions?

Check the [MediaPipe Holistic documentation](https://google.github.io/mediapipe/solutions/holistic.html)
for official information about the graph structure.
