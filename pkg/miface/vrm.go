package miface

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// VRMBone represents a single bone in the VRM skeleton.
type VRMBone struct {
	// Name is the VRM bone name (e.g., "head", "leftUpperArm").
	Name string
	// NodeIndex is the index of the node in the glTF scene.
	NodeIndex int
	// Position is the bone's local position.
	Position Point3D
	// Rotation is the bone's local rotation.
	Rotation Quaternion
	// Scale is the bone's local scale.
	Scale Point3D
	// ParentIndex is the index of the parent bone (-1 if root).
	ParentIndex int
	// Children contains indices of child bones.
	Children []int
}

// VRMSkeleton represents the bone hierarchy extracted from a VRM file.
type VRMSkeleton struct {
	// Bones is a map of VRM bone names to bone data.
	Bones map[string]*VRMBone
	// HumanBones maps VRM humanoid bone names to node indices.
	HumanBones map[string]int
	// ArmSpan is the distance between left and right hand in T-pose.
	ArmSpan float64
	// Height is the estimated model height.
	Height float64
	// HeadSize is the estimated head size (distance from chin to top).
	HeadSize float64
}

// BoneProportions contains calculated bone proportions for tracking calibration.
type BoneProportions struct {
	// ArmLength is the total arm length (shoulder to wrist).
	ArmLength float64
	// UpperArmLength is the upper arm length (shoulder to elbow).
	UpperArmLength float64
	// LowerArmLength is the lower arm length (elbow to wrist).
	LowerArmLength float64
	// SpineLength is the spine length (hips to chest).
	SpineLength float64
	// NeckLength is the neck length (chest to head).
	NeckLength float64
	// HeadSize is the head size.
	HeadSize float64
	// ShoulderWidth is the distance between shoulders.
	ShoulderWidth float64
}

// LoadVRMSkeleton loads bone data from a VRM file without loading meshes or textures.
// This is minimal parsing for calibration purposes only.
func LoadVRMSkeleton(path string) (*VRMSkeleton, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening VRM file: %w", err)
	}
	defer f.Close()

	return ParseVRMSkeleton(f)
}

// ParseVRMSkeleton parses bone data from a VRM file reader.
func ParseVRMSkeleton(r io.Reader) (*VRMSkeleton, error) {
	// Read glTF binary header
	header := make([]byte, 12)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, fmt.Errorf("reading glTF header: %w", err)
	}

	// Verify magic number (glTF)
	magic := binary.LittleEndian.Uint32(header[0:4])
	if magic != 0x46546C67 { // "glTF" in little-endian
		return nil, fmt.Errorf("invalid glTF magic number: %x", magic)
	}

	// Verify version
	version := binary.LittleEndian.Uint32(header[4:8])
	if version != 2 {
		return nil, fmt.Errorf("unsupported glTF version: %d", version)
	}

	// Read JSON chunk header
	chunkHeader := make([]byte, 8)
	if _, err := io.ReadFull(r, chunkHeader); err != nil {
		return nil, fmt.Errorf("reading chunk header: %w", err)
	}

	chunkLength := binary.LittleEndian.Uint32(chunkHeader[0:4])
	chunkType := binary.LittleEndian.Uint32(chunkHeader[4:8])

	if chunkType != 0x4E4F534A { // "JSON" in little-endian
		return nil, fmt.Errorf("expected JSON chunk, got %x", chunkType)
	}

	// Read JSON data
	jsonData := make([]byte, chunkLength)
	if _, err := io.ReadFull(r, jsonData); err != nil {
		return nil, fmt.Errorf("reading JSON chunk: %w", err)
	}

	// Parse glTF JSON
	var gltf gltfDocument
	if err := json.Unmarshal(jsonData, &gltf); err != nil {
		return nil, fmt.Errorf("parsing glTF JSON: %w", err)
	}

	return extractSkeleton(&gltf)
}

// gltfDocument represents the minimal glTF JSON structure needed for skeleton extraction.
type gltfDocument struct {
	Nodes      []gltfNode      `json:"nodes"`
	Extensions gltfExtensions  `json:"extensions"`
}

type gltfNode struct {
	Name        string    `json:"name"`
	Children    []int     `json:"children"`
	Translation []float64 `json:"translation"`
	Rotation    []float64 `json:"rotation"`
	Scale       []float64 `json:"scale"`
}

type gltfExtensions struct {
	VRM  *vrmExtension  `json:"VRM"`
	VRMC *vrmcExtension `json:"VRMC_vrm"`
}

// VRM 0.x extension
type vrmExtension struct {
	Humanoid *vrmHumanoid `json:"humanoid"`
}

type vrmHumanoid struct {
	HumanBones []vrmHumanBone `json:"humanBones"`
}

type vrmHumanBone struct {
	Bone string `json:"bone"`
	Node int    `json:"node"`
}

// VRM 1.0 extension
type vrmcExtension struct {
	Humanoid *vrmcHumanoid `json:"humanoid"`
}

type vrmcHumanoid struct {
	HumanBones map[string]vrmcHumanBone `json:"humanBones"`
}

type vrmcHumanBone struct {
	Node int `json:"node"`
}

// extractSkeleton extracts skeleton data from parsed glTF.
func extractSkeleton(gltf *gltfDocument) (*VRMSkeleton, error) {
	skeleton := &VRMSkeleton{
		Bones:      make(map[string]*VRMBone),
		HumanBones: make(map[string]int),
	}

	// Build parent-child relationships
	parentMap := make(map[int]int)
	for i, node := range gltf.Nodes {
		for _, child := range node.Children {
			parentMap[child] = i
		}
	}

	// Extract nodes as bones
	for i, node := range gltf.Nodes {
		bone := &VRMBone{
			Name:        node.Name,
			NodeIndex:   i,
			ParentIndex: -1,
			Children:    node.Children,
		}

		if parent, ok := parentMap[i]; ok {
			bone.ParentIndex = parent
		}

		// Parse transform
		if len(node.Translation) >= 3 {
			bone.Position = Point3D{
				X: node.Translation[0],
				Y: node.Translation[1],
				Z: node.Translation[2],
			}
		}

		if len(node.Rotation) >= 4 {
			bone.Rotation = Quaternion{
				X: node.Rotation[0],
				Y: node.Rotation[1],
				Z: node.Rotation[2],
				W: node.Rotation[3],
			}
		} else {
			bone.Rotation = Quaternion{W: 1} // Identity
		}

		if len(node.Scale) >= 3 {
			bone.Scale = Point3D{
				X: node.Scale[0],
				Y: node.Scale[1],
				Z: node.Scale[2],
			}
		} else {
			bone.Scale = Point3D{X: 1, Y: 1, Z: 1}
		}

		skeleton.Bones[node.Name] = bone
	}

	// Extract VRM humanoid bone mappings
	if gltf.Extensions.VRM != nil && gltf.Extensions.VRM.Humanoid != nil {
		// VRM 0.x format
		for _, hb := range gltf.Extensions.VRM.Humanoid.HumanBones {
			skeleton.HumanBones[hb.Bone] = hb.Node
		}
	} else if gltf.Extensions.VRMC != nil && gltf.Extensions.VRMC.Humanoid != nil {
		// VRM 1.0 format
		for name, hb := range gltf.Extensions.VRMC.Humanoid.HumanBones {
			skeleton.HumanBones[name] = hb.Node
		}
	}

	// Calculate model proportions
	skeleton.calculateProportions(gltf.Nodes)

	return skeleton, nil
}

// calculateProportions calculates body proportions from bone positions.
func (s *VRMSkeleton) calculateProportions(nodes []gltfNode) {
	// Get key bone positions
	getWorldPos := func(boneName string) (Point3D, bool) {
		nodeIdx, ok := s.HumanBones[boneName]
		if !ok || nodeIdx >= len(nodes) {
			return Point3D{}, false
		}
		
		// For simplicity, use local position (proper implementation would compute world transforms)
		node := nodes[nodeIdx]
		if len(node.Translation) >= 3 {
			return Point3D{
				X: node.Translation[0],
				Y: node.Translation[1],
				Z: node.Translation[2],
			}, true
		}
		return Point3D{}, false
	}

	// Calculate arm span
	if leftHand, ok := getWorldPos("leftHand"); ok {
		if rightHand, ok := getWorldPos("rightHand"); ok {
			s.ArmSpan = distance(leftHand, rightHand)
		}
	}

	// Estimate height from hips to head
	if hips, ok := getWorldPos("hips"); ok {
		if head, ok := getWorldPos("head"); ok {
			s.Height = head.Y - hips.Y
			// Add estimated leg length (roughly equal to upper body)
			s.Height *= 2
		}
	}

	// Estimate head size from head to neck
	if head, ok := getWorldPos("head"); ok {
		if neck, ok := getWorldPos("neck"); ok {
			s.HeadSize = distance(head, neck) * 1.5 // Approximate full head size
		}
	}
}

// GetProportions calculates detailed bone proportions for tracking calibration.
func (s *VRMSkeleton) GetProportions() *BoneProportions {
	props := &BoneProportions{
		HeadSize: s.HeadSize,
	}

	// Helper to get bone by VRM name
	getBone := func(name string) *VRMBone {
		if nodeIdx, ok := s.HumanBones[name]; ok {
			for _, bone := range s.Bones {
				if bone.NodeIndex == nodeIdx {
					return bone
				}
			}
		}
		return nil
	}

	// Calculate arm proportions (using left arm as reference)
	if shoulder := getBone("leftUpperArm"); shoulder != nil {
		if elbow := getBone("leftLowerArm"); elbow != nil {
			props.UpperArmLength = distance(shoulder.Position, elbow.Position)
			if wrist := getBone("leftHand"); wrist != nil {
				props.LowerArmLength = distance(elbow.Position, wrist.Position)
				props.ArmLength = props.UpperArmLength + props.LowerArmLength
			}
		}
	}

	// Calculate spine length
	if hips := getBone("hips"); hips != nil {
		if chest := getBone("chest"); chest != nil {
			props.SpineLength = distance(hips.Position, chest.Position)
		}
	}

	// Calculate neck length
	if chest := getBone("chest"); chest != nil {
		if head := getBone("head"); head != nil {
			props.NeckLength = distance(chest.Position, head.Position)
		}
	}

	// Calculate shoulder width
	if leftShoulder := getBone("leftUpperArm"); leftShoulder != nil {
		if rightShoulder := getBone("rightUpperArm"); rightShoulder != nil {
			props.ShoulderWidth = distance(leftShoulder.Position, rightShoulder.Position)
		}
	}

	return props
}

// GetBonePosition returns the world position of a VRM bone by name.
func (s *VRMSkeleton) GetBonePosition(boneName string) (Point3D, bool) {
	nodeIdx, ok := s.HumanBones[boneName]
	if !ok {
		return Point3D{}, false
	}

	for _, bone := range s.Bones {
		if bone.NodeIndex == nodeIdx {
			return bone.Position, true
		}
	}

	return Point3D{}, false
}

// ListHumanBones returns a list of all available humanoid bone names.
func (s *VRMSkeleton) ListHumanBones() []string {
	names := make([]string, 0, len(s.HumanBones))
	for name := range s.HumanBones {
		names = append(names, name)
	}
	return names
}

// distance calculates the Euclidean distance between two 3D points.
func distance(a, b Point3D) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return sqrt(dx*dx + dy*dy + dz*dz)
}

// sqrt is a simple square root approximation using Newton's method.
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}
