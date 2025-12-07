# Bazel BUILD file for MediaPipe Bridge

load("@rules_cc//cc:defs.bzl", "cc_library", "cc_binary", "cc_test")

package(default_visibility = ["//visibility:public"])

# Main bridge library (shared object)
cc_library(
    name = "mediapipe_bridge_lib",
    srcs = [
        "mediapipe_bridge.cc",
        "holistic_config.cc",
    ],
    hdrs = ["mediapipe_bridge.h"],
    deps = [
        "@mediapipe//mediapipe/framework:calculator_framework",
        "@mediapipe//mediapipe/framework/formats:image_frame",
        "@mediapipe//mediapipe/framework/formats:image_frame_opencv",
        "@mediapipe//mediapipe/framework/formats:landmark_cc_proto",
        "@mediapipe//mediapipe/framework/port:parse_text_proto",
        "@mediapipe//mediapipe/framework/port:status",
        # Holistic tracking dependencies
        "@mediapipe//mediapipe/graphs/holistic_tracking:holistic_tracking_cpu_graph_deps",
        # OpenCV
        "@linux_opencv//:opencv",
    ],
    copts = [
        "-std=c++17",
        "-fPIC",
        "-Wall",
        "-Wextra",
    ],
    linkopts = [
        "-lpthread",
    ],
)

# Shared library output
cc_binary(
    name = "libmediapipe_bridge.so",
    linkshared = True,
    linkstatic = False,
    deps = [":mediapipe_bridge_lib"],
)

# Test program
cc_binary(
    name = "bridge_test",
    srcs = ["bridge_test.cc"],
    deps = [":mediapipe_bridge_lib"],
    copts = ["-std=c++17"],
)

# For GPU support, add alternative target
cc_library(
    name = "mediapipe_bridge_gpu_lib",
    srcs = [
        "mediapipe_bridge.cc",
        "holistic_config.cc",
    ],
    hdrs = ["mediapipe_bridge.h"],
    deps = [
        "@mediapipe//mediapipe/framework:calculator_framework",
        "@mediapipe//mediapipe/framework/formats:image_frame",
        "@mediapipe//mediapipe/framework/formats:image_frame_opencv",
        "@mediapipe//mediapipe/framework/formats:landmark_cc_proto",
        "@mediapipe//mediapipe/framework/port:parse_text_proto",
        "@mediapipe//mediapipe/framework/port:status",
        "@mediapipe//mediapipe/graphs/holistic_tracking:holistic_tracking_gpu_graph_deps",
        "@mediapipe//mediapipe/gpu:gpu_buffer",
        "@mediapipe//mediapipe/gpu:gpu_shared_data_internal",
        "@linux_opencv//:opencv",
    ],
    defines = ["MEDIAPIPE_GPU_ENABLED"],
    copts = [
        "-std=c++17",
        "-fPIC",
        "-Wall",
        "-Wextra",
    ],
    linkopts = [
        "-lpthread",
        "-lGLESv2",
        "-lEGL",
    ],
)

cc_binary(
    name = "libmediapipe_bridge_gpu.so",
    linkshared = True,
    linkstatic = False,
    deps = [":mediapipe_bridge_gpu_lib"],
)
