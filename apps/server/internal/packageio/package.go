package packageio

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
)

const (
	PackageVersion       = "1.0.0"
	MaxPackageSize int64 = 500 << 20
	MaxEntrySize   int64 = 25 << 20
	MaxEntries           = 2000
)

var ErrInvalidPackage = errors.New("invalid ddivination package")

type Manifest struct {
	PackageVersion string               `json:"packageVersion"`
	SchemaVersion  string               `json:"schemaVersion"`
	AdventureID    string               `json:"adventureId"`
	AdventureHash  string               `json:"adventureHash"`
	CreatedAt      time.Time            `json:"createdAt"`
	Attributions   []domain.Attribution `json:"attributions"`
}

func Export(doc domain.AdventureDocument) ([]byte, error) {
	adventureJSON, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(adventureJSON)
	manifest := Manifest{
		PackageVersion: PackageVersion,
		SchemaVersion:  doc.SchemaVersion,
		AdventureID:    doc.ID,
		AdventureHash:  hex.EncodeToString(hash[:]),
		CreatedAt:      time.Now().UTC(),
		Attributions:   doc.Attributions,
	}
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	if err := writeEntry(writer, "manifest.json", manifestJSON); err != nil {
		return nil, err
	}
	if err := writeEntry(writer, "adventure.json", adventureJSON); err != nil {
		return nil, err
	}
	if err := writeEntry(writer, "ATTRIBUTION.md", []byte(attributionMarkdown(doc.Attributions))); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func Import(reader io.ReaderAt, size int64) (domain.AdventureDocument, error) {
	if size < 1 || size > MaxPackageSize {
		return domain.AdventureDocument{}, fmt.Errorf("%w: package size exceeds limit", ErrInvalidPackage)
	}
	archive, err := zip.NewReader(reader, size)
	if err != nil {
		return domain.AdventureDocument{}, fmt.Errorf("%w: %v", ErrInvalidPackage, err)
	}
	if len(archive.File) > MaxEntries {
		return domain.AdventureDocument{}, fmt.Errorf("%w: too many entries", ErrInvalidPackage)
	}
	files := make(map[string]*zip.File, len(archive.File))
	var uncompressed uint64
	for _, file := range archive.File {
		clean := path.Clean(strings.ReplaceAll(file.Name, "\\", "/"))
		if clean != file.Name || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
			return domain.AdventureDocument{}, fmt.Errorf("%w: unsafe entry path", ErrInvalidPackage)
		}
		if file.UncompressedSize64 > uint64(MaxEntrySize) {
			return domain.AdventureDocument{}, fmt.Errorf("%w: entry too large", ErrInvalidPackage)
		}
		uncompressed += file.UncompressedSize64
		if uncompressed > uint64(MaxPackageSize) {
			return domain.AdventureDocument{}, fmt.Errorf("%w: uncompressed content exceeds limit", ErrInvalidPackage)
		}
		files[clean] = file
	}
	manifestBytes, err := readEntry(files["manifest.json"])
	if err != nil {
		return domain.AdventureDocument{}, fmt.Errorf("%w: missing manifest", ErrInvalidPackage)
	}
	adventureBytes, err := readEntry(files["adventure.json"])
	if err != nil {
		return domain.AdventureDocument{}, fmt.Errorf("%w: missing adventure", ErrInvalidPackage)
	}
	var manifest Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return domain.AdventureDocument{}, fmt.Errorf("%w: malformed manifest", ErrInvalidPackage)
	}
	if manifest.PackageVersion != PackageVersion || manifest.SchemaVersion != domain.SchemaVersion {
		return domain.AdventureDocument{}, fmt.Errorf("%w: unsupported version", ErrInvalidPackage)
	}
	hash := sha256.Sum256(adventureBytes)
	if hex.EncodeToString(hash[:]) != manifest.AdventureHash {
		return domain.AdventureDocument{}, fmt.Errorf("%w: adventure hash mismatch", ErrInvalidPackage)
	}
	var doc domain.AdventureDocument
	if err := json.Unmarshal(adventureBytes, &doc); err != nil {
		return domain.AdventureDocument{}, fmt.Errorf("%w: malformed adventure", ErrInvalidPackage)
	}
	if doc.ID != manifest.AdventureID || doc.SchemaVersion != domain.SchemaVersion {
		return domain.AdventureDocument{}, fmt.Errorf("%w: manifest mismatch", ErrInvalidPackage)
	}
	doc.Version++
	doc.UpdatedAt = time.Now().UTC()
	return doc, nil
}

func writeEntry(writer *zip.Writer, name string, content []byte) error {
	entry, err := writer.Create(name)
	if err != nil {
		return err
	}
	_, err = entry.Write(content)
	return err
}

func readEntry(file *zip.File) ([]byte, error) {
	if file == nil {
		return nil, io.EOF
	}
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(io.LimitReader(reader, MaxEntrySize+1))
}

func attributionMarkdown(attributions []domain.Attribution) string {
	var builder strings.Builder
	builder.WriteString("# Attributions\n\n")
	for _, attribution := range attributions {
		fmt.Fprintf(&builder, "## %s\n\n- Creator: %s\n- Source: %s\n- License: %s\n\n%s\n\n",
			attribution.Title, attribution.Creator, attribution.Source, attribution.License, attribution.Notice)
	}
	return builder.String()
}
