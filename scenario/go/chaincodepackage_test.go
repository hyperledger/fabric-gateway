// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package scenario

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// A chaincode package is a gzipped tar file containing a metadata descriptor and a gzipped tar file of the chaincode
// source. Source files appear below a "src" directory, with the exception of chaincode metadata files, which appear
// below "META-INF". This is the packaging carried out by the "peer lifecycle chaincode package" command.
const (
	metadataFileName    = "metadata.json"
	codePackageFileName = "code.tar.gz"
	sourcePathPrefix    = "src"
	metadataPathPrefix  = "META-INF"

	// Directories that never contain chaincode source. Dependencies of Node chaincode are installed by the peer when
	// it builds the chaincode.
	nodeModulesDirName = "node_modules"

	golangChaincodeLanguage = "golang"
)

// chaincodeMetadata is the content of the metadata descriptor within a chaincode package.
type chaincodeMetadata struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

// tarEntry is a file to be written to a tar archive.
type tarEntry struct {
	name    string
	content []byte
}

// newChaincodePackage creates a chaincode package for the chaincode source in a specific directory. The package
// content is deterministic so the same source always yields the same chaincode package identifier.
func newChaincodePackage(language string, sourceDir string, label string) ([]byte, error) {
	codePackage, err := newCodePackage(sourceDir)
	if err != nil {
		return nil, err
	}

	ccPath, err := chaincodePath(language, sourceDir)
	if err != nil {
		return nil, err
	}

	metadata, err := json.Marshal(&chaincodeMetadata{
		Path:  ccPath,
		Type:  language,
		Label: label,
	})
	if err != nil {
		return nil, err
	}

	return newGzipTar([]tarEntry{
		{name: metadataFileName, content: metadata},
		{name: codePackageFileName, content: codePackage},
	})
}

// chaincodePath is recorded in the chaincode package metadata and used by the peer to build the chaincode. Go
// chaincode is built from its module path; other languages are built from the packaged source so the value is only
// informational.
func chaincodePath(language string, sourceDir string) (string, error) {
	if language != golangChaincodeLanguage {
		return sourceDir, nil
	}

	return goModulePath(filepath.Join(sourceDir, "go.mod"))
}

var goModulePattern = regexp.MustCompile(`(?m)^\s*module\s+(\S+)`)

func goModulePath(goModFile string) (string, error) {
	content, err := os.ReadFile(goModFile) // #nosec G304
	if err != nil {
		return "", err
	}

	match := goModulePattern.FindSubmatch(content)
	if match == nil {
		return "", fmt.Errorf("no module path found in %s", goModFile)
	}

	return strings.Trim(string(match[1]), `"`), nil
}

// newCodePackage creates the gzipped tar file of chaincode source that is embedded within a chaincode package.
func newCodePackage(sourceDir string) ([]byte, error) {
	root := filepath.Clean(sourceDir)
	metadataDir := filepath.Join(root, metadataPathPrefix)

	var entries []tarEntry

	walk := func(file string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if dirEntry.IsDir() {
			if file != root && isExcludedDir(dirEntry.Name()) {
				return fs.SkipDir
			}
			return nil
		}

		entry, err := newSourceEntry(root, metadataDir, file)
		if err != nil {
			return err
		}

		entries = append(entries, entry)
		return nil
	}

	if err := filepath.WalkDir(root, walk); err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no chaincode source files found in %s", sourceDir)
	}

	return newGzipTar(entries)
}

func isExcludedDir(name string) bool {
	return strings.HasPrefix(name, ".") || name == nodeModulesDirName
}

func newSourceEntry(root string, metadataDir string, file string) (tarEntry, error) {
	relativeRoot, prefix := root, sourcePathPrefix
	if isInDir(file, metadataDir) {
		relativeRoot, prefix = metadataDir, metadataPathPrefix
	}

	relativePath, err := filepath.Rel(relativeRoot, file)
	if err != nil {
		return tarEntry{}, err
	}

	content, err := os.ReadFile(file) // #nosec G304
	if err != nil {
		return tarEntry{}, err
	}

	return tarEntry{
		name:    path.Join(prefix, filepath.ToSlash(relativePath)),
		content: content,
	}, nil
}

func isInDir(file string, dir string) bool {
	return strings.HasPrefix(filepath.Clean(file), filepath.Clean(dir)+string(filepath.Separator))
}

func newGzipTar(entries []tarEntry) ([]byte, error) {
	buffer := new(bytes.Buffer)
	gzipWriter := gzip.NewWriter(buffer)
	tarWriter := tar.NewWriter(gzipWriter)

	for _, entry := range entries {
		if err := writeTarEntry(tarWriter, entry); err != nil {
			return nil, err
		}
	}

	if err := tarWriter.Close(); err != nil {
		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// writeTarEntry writes a file to a tar archive. File attributes are fixed values, both to match the chaincode
// packaging carried out by the Fabric peer CLI, and to ensure the archive content is deterministic.
func writeTarEntry(writer *tar.Writer, entry tarEntry) error {
	header := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     entry.name,
		Size:     int64(len(entry.content)),
		Mode:     0o100644,
		Uid:      500,
		Gid:      500,
	}
	if err := writer.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header for %s: %w", entry.name, err)
	}

	if _, err := writer.Write(entry.content); err != nil {
		return fmt.Errorf("failed to write tar entry %s: %w", entry.name, err)
	}

	return nil
}
