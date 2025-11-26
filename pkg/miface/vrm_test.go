package miface

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"
)

// createTestVRM creates a minimal VRM binary for testing.
func createTestVRM(t *testing.T) []byte {
	t.Helper()

	// Create glTF JSON with VRM extension
	gltf := map[string]interface{}{
		"asset": map[string]interface{}{
			"version": "2.0",
		},
		"nodes": []map[string]interface{}{
			{
				"name":        "Armature",
				"children":    []int{1, 2},
				"translation": []float64{0, 0, 0},
			},
			{
				"name":        "Hips",
				"children":    []int{3},
				"translation": []float64{0, 1.0, 0},
			},
			{
				"name":        "Head",
				"translation": []float64{0, 1.7, 0},
			},
			{
				"name":        "Spine",
				"translation": []float64{0, 1.2, 0},
			},
		},
		"extensions": map[string]interface{}{
			"VRM": map[string]interface{}{
				"humanoid": map[string]interface{}{
					"humanBones": []map[string]interface{}{
						{"bone": "hips", "node": 1},
						{"bone": "head", "node": 2},
						{"bone": "spine", "node": 3},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(gltf)
	if err != nil {
		t.Fatalf("failed to marshal test glTF: %v", err)
	}

	// Pad JSON to 4-byte boundary
	padding := (4 - len(jsonData)%4) % 4
	for i := 0; i < padding; i++ {
		jsonData = append(jsonData, ' ')
	}

	// Build glTF binary
	var buf bytes.Buffer

	// Header
	buf.Write([]byte("glTF"))                                    // magic
	_ = binary.Write(&buf, binary.LittleEndian, uint32(2))           // version
	_ = binary.Write(&buf, binary.LittleEndian, uint32(12+8+len(jsonData))) // total length

	// JSON chunk
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(jsonData))) // chunk length
	buf.Write([]byte("JSON"))                                      // chunk type
	buf.Write(jsonData)

	return buf.Bytes()
}

func TestParseVRMSkeleton(t *testing.T) {
	data := createTestVRM(t)
	reader := bytes.NewReader(data)

	skeleton, err := ParseVRMSkeleton(reader)
	if err != nil {
		t.Fatalf("failed to parse VRM: %v", err)
	}

	if skeleton == nil {
		t.Fatal("expected non-nil skeleton")
	}

	// Check bones were extracted
	if len(skeleton.Bones) != 4 {
		t.Errorf("expected 4 bones, got %d", len(skeleton.Bones))
	}

	// Check humanoid mapping
	if len(skeleton.HumanBones) != 3 {
		t.Errorf("expected 3 human bones, got %d", len(skeleton.HumanBones))
	}

	if _, ok := skeleton.HumanBones["hips"]; !ok {
		t.Error("expected 'hips' in humanoid mapping")
	}

	if _, ok := skeleton.HumanBones["head"]; !ok {
		t.Error("expected 'head' in humanoid mapping")
	}
}

func TestParseVRMSkeletonInvalidMagic(t *testing.T) {
	data := []byte("NOTG\x02\x00\x00\x00\x14\x00\x00\x00")
	reader := bytes.NewReader(data)

	_, err := ParseVRMSkeleton(reader)
	if err == nil {
		t.Error("expected error for invalid magic number")
	}
}

func TestParseVRMSkeletonInvalidVersion(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte("glTF"))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(1)) // version 1 not supported
	_ = binary.Write(&buf, binary.LittleEndian, uint32(12))

	reader := bytes.NewReader(buf.Bytes())
	_, err := ParseVRMSkeleton(reader)
	if err == nil {
		t.Error("expected error for unsupported version")
	}
}

func TestVRMSkeletonGetBonePosition(t *testing.T) {
	data := createTestVRM(t)
	reader := bytes.NewReader(data)

	skeleton, err := ParseVRMSkeleton(reader)
	if err != nil {
		t.Fatalf("failed to parse VRM: %v", err)
	}

	pos, ok := skeleton.GetBonePosition("hips")
	if !ok {
		t.Error("expected to find hips bone")
	}
	if pos.Y != 1.0 {
		t.Errorf("expected hips Y=1.0, got %f", pos.Y)
	}

	_, ok = skeleton.GetBonePosition("nonexistent")
	if ok {
		t.Error("expected false for nonexistent bone")
	}
}

func TestVRMSkeletonListHumanBones(t *testing.T) {
	data := createTestVRM(t)
	reader := bytes.NewReader(data)

	skeleton, err := ParseVRMSkeleton(reader)
	if err != nil {
		t.Fatalf("failed to parse VRM: %v", err)
	}

	bones := skeleton.ListHumanBones()
	if len(bones) != 3 {
		t.Errorf("expected 3 bones, got %d", len(bones))
	}
}

func TestVRMSkeletonGetProportions(t *testing.T) {
	data := createTestVRM(t)
	reader := bytes.NewReader(data)

	skeleton, err := ParseVRMSkeleton(reader)
	if err != nil {
		t.Fatalf("failed to parse VRM: %v", err)
	}

	props := skeleton.GetProportions()
	if props == nil {
		t.Fatal("expected non-nil proportions")
	}
}

func TestDistance(t *testing.T) {
	a := Point3D{X: 0, Y: 0, Z: 0}
	b := Point3D{X: 3, Y: 4, Z: 0}

	d := distance(a, b)
	if d < 4.99 || d > 5.01 {
		t.Errorf("expected distance ~5, got %f", d)
	}
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0, 0},
		{1, 1},
		{4, 2},
		{9, 3},
		{25, 5},
	}

	for _, tt := range tests {
		result := sqrt(tt.input)
		diff := result - tt.expected
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.0001 {
			t.Errorf("sqrt(%f) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}

// Test VRM 1.0 format
func createTestVRM1(t *testing.T) []byte {
	t.Helper()

	gltf := map[string]interface{}{
		"asset": map[string]interface{}{
			"version": "2.0",
		},
		"nodes": []map[string]interface{}{
			{
				"name":        "Root",
				"translation": []float64{0, 0, 0},
			},
			{
				"name":        "Hips",
				"translation": []float64{0, 1.0, 0},
			},
		},
		"extensions": map[string]interface{}{
			"VRMC_vrm": map[string]interface{}{
				"humanoid": map[string]interface{}{
					"humanBones": map[string]interface{}{
						"hips": map[string]interface{}{"node": 1},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(gltf)
	if err != nil {
		t.Fatalf("failed to marshal test glTF: %v", err)
	}

	padding := (4 - len(jsonData)%4) % 4
	for i := 0; i < padding; i++ {
		jsonData = append(jsonData, ' ')
	}

	var buf bytes.Buffer
	buf.Write([]byte("glTF"))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(2))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(12+8+len(jsonData)))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(jsonData)))
	buf.Write([]byte("JSON"))
	buf.Write(jsonData)

	return buf.Bytes()
}

func TestParseVRM1Skeleton(t *testing.T) {
	data := createTestVRM1(t)
	reader := bytes.NewReader(data)

	skeleton, err := ParseVRMSkeleton(reader)
	if err != nil {
		t.Fatalf("failed to parse VRM 1.0: %v", err)
	}

	if _, ok := skeleton.HumanBones["hips"]; !ok {
		t.Error("expected 'hips' in VRM 1.0 humanoid mapping")
	}
}
