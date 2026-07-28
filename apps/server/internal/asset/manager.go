package asset

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
)

const (
	MaxImageSize int64 = 10 << 20
	MaxGLBSize   int64 = 25 << 20
)

var ErrInvalidAsset = errors.New("invalid asset")

type persistence interface {
	SaveAsset(context.Context, domain.AssetRef, string) error
}

type Manager struct {
	root  string
	store persistence
}

func NewManager(root string, store persistence) *Manager {
	return &Manager{root: root, store: store}
}

func (m *Manager) Import(
	ctx context.Context,
	fileName string,
	reader io.Reader,
	creator string,
	license string,
) (domain.AssetRef, error) {
	content, err := io.ReadAll(io.LimitReader(reader, MaxGLBSize+1))
	if err != nil {
		return domain.AssetRef{}, err
	}
	mediaType, kind, extension, err := validate(content)
	if err != nil {
		return domain.AssetRef{}, err
	}
	limit := MaxGLBSize
	if kind == "image" {
		limit = MaxImageSize
	}
	if int64(len(content)) > limit {
		return domain.AssetRef{}, fmt.Errorf("%w: file exceeds %d bytes", ErrInvalidAsset, limit)
	}
	if strings.TrimSpace(creator) == "" {
		creator = "User-provided asset"
	}
	if strings.TrimSpace(license) == "" {
		license = "User-managed"
	}
	hash := sha256.Sum256(content)
	hashText := hex.EncodeToString(hash[:])
	id := "asset-" + hashText[:24]
	storedName := hashText + extension
	if err := os.MkdirAll(m.root, 0o755); err != nil {
		return domain.AssetRef{}, err
	}
	target := filepath.Join(m.root, storedName)
	if err := os.WriteFile(target, content, 0o644); err != nil {
		return domain.AssetRef{}, err
	}
	ref := domain.AssetRef{
		ID:        id,
		SHA256:    hashText,
		FileName:  cleanFileName(fileName),
		MediaType: mediaType,
		Size:      int64(len(content)),
		Kind:      kind,
		License: domain.Attribution{
			Title:   cleanFileName(fileName),
			Creator: strings.TrimSpace(creator),
			License: strings.TrimSpace(license),
			Notice:  "Imported by the user; redistribution rights must be verified by the user.",
		},
	}
	if err := m.store.SaveAsset(ctx, ref, storedName); err != nil {
		return domain.AssetRef{}, err
	}
	return ref, nil
}

func validate(content []byte) (mediaType, kind, extension string, err error) {
	detected := http.DetectContentType(content)
	switch detected {
	case "image/png":
		config, _, decodeErr := image.DecodeConfig(bytes.NewReader(content))
		if decodeErr != nil || config.Width < 1 || config.Height < 1 ||
			config.Width > 8192 || config.Height > 8192 {
			return "", "", "", fmt.Errorf("%w: invalid or oversized PNG", ErrInvalidAsset)
		}
		return detected, "image", ".png", nil
	case "image/webp":
		if len(content) < 16 || string(content[:4]) != "RIFF" || string(content[8:12]) != "WEBP" {
			return "", "", "", fmt.Errorf("%w: malformed WebP", ErrInvalidAsset)
		}
		width, height, dimensionsOK := webpDimensions(content)
		if !dimensionsOK || width < 1 || height < 1 || width > 8192 || height > 8192 {
			return "", "", "", fmt.Errorf("%w: invalid or oversized WebP", ErrInvalidAsset)
		}
		return detected, "image", ".webp", nil
	case "model/gltf-binary", "application/octet-stream":
		if err := validateGLB(content); err != nil {
			return "", "", "", err
		}
		return "model/gltf-binary", "model", ".glb", nil
	default:
		return "", "", "", fmt.Errorf("%w: unsupported media type %s", ErrInvalidAsset, detected)
	}
}

func webpDimensions(content []byte) (int, int, bool) {
	if len(content) < 30 {
		return 0, 0, false
	}
	switch string(content[12:16]) {
	case "VP8X":
		width := 1 + int(content[24]) + int(content[25])<<8 + int(content[26])<<16
		height := 1 + int(content[27]) + int(content[28])<<8 + int(content[29])<<16
		return width, height, true
	case "VP8L":
		if len(content) < 25 || content[20] != 0x2f {
			return 0, 0, false
		}
		width := 1 + int(content[21]) + int(content[22]&0x3f)<<8
		height := 1 + int(content[22]>>6) + int(content[23])<<2 + int(content[24]&0x0f)<<10
		return width, height, true
	case "VP8 ":
		if len(content) < 30 || !bytes.Equal(content[23:26], []byte{0x9d, 0x01, 0x2a}) {
			return 0, 0, false
		}
		width := int(binary.LittleEndian.Uint16(content[26:28]) & 0x3fff)
		height := int(binary.LittleEndian.Uint16(content[28:30]) & 0x3fff)
		return width, height, true
	default:
		return 0, 0, false
	}
}

func validateGLB(content []byte) error {
	if len(content) < 20 || string(content[:4]) != "glTF" ||
		binary.LittleEndian.Uint32(content[4:8]) != 2 ||
		int(binary.LittleEndian.Uint32(content[8:12])) != len(content) {
		return fmt.Errorf("%w: malformed GLB header", ErrInvalidAsset)
	}
	offset := 12
	foundJSON := false
	for offset+8 <= len(content) {
		length := int(binary.LittleEndian.Uint32(content[offset : offset+4]))
		chunkType := binary.LittleEndian.Uint32(content[offset+4 : offset+8])
		offset += 8
		if length < 0 || offset+length > len(content) {
			return fmt.Errorf("%w: malformed GLB chunk", ErrInvalidAsset)
		}
		if chunkType == 0x4E4F534A {
			foundJSON = true
			var document any
			if err := json.Unmarshal(bytes.TrimRight(content[offset:offset+length], " \x00"), &document); err != nil {
				return fmt.Errorf("%w: malformed GLB JSON", ErrInvalidAsset)
			}
			if externalURI(document) {
				return fmt.Errorf("%w: GLB contains an external resource", ErrInvalidAsset)
			}
		}
		offset += length
	}
	if !foundJSON || offset != len(content) {
		return fmt.Errorf("%w: incomplete GLB", ErrInvalidAsset)
	}
	return nil
}

func externalURI(value any) bool {
	switch current := value.(type) {
	case map[string]any:
		for key, child := range current {
			if key == "uri" {
				if uri, ok := child.(string); ok && !strings.HasPrefix(uri, "data:") {
					return true
				}
			}
			if externalURI(child) {
				return true
			}
		}
	case []any:
		for _, child := range current {
			if externalURI(child) {
				return true
			}
		}
	}
	return false
}

func cleanFileName(value string) string {
	value = strings.TrimSpace(filepath.Base(value))
	if value == "" || value == "." {
		return "asset-" + time.Now().UTC().Format("20060102-150405")
	}
	if len(value) > 180 {
		value = value[:180]
	}
	return value
}
