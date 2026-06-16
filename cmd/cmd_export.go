package cmd

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lychee/lychee/envconfig"
	"github.com/lychee/lychee/format"
	"github.com/lychee/lychee/manifest"
	"github.com/lychee/lychee/types/model"
)

func NewExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export MODEL OUTPUT_FILE.lychee",
		Short: "Export a model and all its layer blobs into a portable archive",
		Long:  `Gzips and tars the model manifest and all corresponding blobs (weights, template, system, config) for sharing.`,
		Args:  cobra.ExactArgs(2),
		ValidArgsFunction: autocompleteInstalledModels,
		RunE: func(cmd *cobra.Command, args []string) error {
			modelName := args[0]
			outputPath := args[1]

			if !strings.HasSuffix(outputPath, ".lychee") {
				outputPath += ".lychee"
			}

			name := model.ParseName(modelName)
			m, err := manifest.ParseNamedManifest(name)
			if err != nil {
				return fmt.Errorf("model %q not found locally: %w", modelName, err)
			}

			fmt.Printf("Exporting model %s -> %s...\n", modelName, outputPath)

			out, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			defer out.Close()

			gw := gzip.NewWriter(out)
			defer gw.Close()

			tw := tar.NewWriter(gw)
			defer tw.Close()

			// 1. Write manifest contents
			manifestBytes, err := os.ReadFile(m.FileInfo().Name())
			if err != nil {
				// Read file path directly
				manifestPath, err := manifest.PathForName(name)
				if err != nil {
					return err
				}
				manifestBytes, err = os.ReadFile(manifestPath)
				if err != nil {
					// Fallback to marshal
					manifestBytes, err = json.Marshal(m)
					if err != nil {
						return err
					}
				}
			}

			hdr := &tar.Header{
				Name: "manifest.json",
				Size: int64(len(manifestBytes)),
				Mode: 0644,
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			if _, err := tw.Write(manifestBytes); err != nil {
				return err
			}

			// 2. Write model path file
			relPath := name.Filepath()
			hdrPath := &tar.Header{
				Name: "model_name.txt",
				Size: int64(len(relPath)),
				Mode: 0644,
			}
			if err := tw.WriteHeader(hdrPath); err != nil {
				return err
			}
			if _, err := tw.Write([]byte(relPath)); err != nil {
				return err
			}

			// 3. Write blobs
			allLayers := append(m.Layers, m.Config)
			for _, layer := range allLayers {
				if layer.Digest == "" {
					continue
				}
				blobPath, err := manifest.BlobsPath(layer.Digest)
				if err != nil {
					return err
				}

				blobFile, err := os.Open(blobPath)
				if err != nil {
					return fmt.Errorf("blob file not found: %s: %w", layer.Digest, err)
				}

				stat, err := blobFile.Stat()
				if err != nil {
					blobFile.Close()
					return err
				}

				hdrBlob := &tar.Header{
					Name: "blobs/" + strings.ReplaceAll(layer.Digest, ":", "-"),
					Size: stat.Size(),
					Mode: 0644,
				}
				if err := tw.WriteHeader(hdrBlob); err != nil {
					blobFile.Close()
					return err
				}

				fmt.Printf("Packing layer blob %s (%s)...\n", layer.Digest[:12], format.HumanBytes(stat.Size()))
				if _, err := io.Copy(tw, blobFile); err != nil {
					blobFile.Close()
					return err
				}
				blobFile.Close()
			}

			fmt.Println("✅ Model successfully exported!")
			return nil
		},
	}
	return cmd
}

func NewImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import ARCHIVE_FILE.lychee",
		Short: "Import a portable model archive into your local catalog",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			archivePath := args[0]
			if _, err := os.Stat(archivePath); err != nil {
				return fmt.Errorf("archive file not found: %w", err)
			}

			fmt.Printf("Importing model from %s...\n", archivePath)

			f, err := os.Open(archivePath)
			if err != nil {
				return err
			}
			defer f.Close()

			gr, err := gzip.NewReader(f)
			if err != nil {
				return err
			}
			defer gr.Close()

			tr := tar.NewReader(gr)

			var manifestBytes []byte
			var modelNameStr string

			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}

				// Zip Slip / Directory Traversal Guard
				cleanedName := filepath.Clean(hdr.Name)
				if strings.HasPrefix(cleanedName, "..") || strings.HasPrefix(cleanedName, "/") || filepath.IsAbs(cleanedName) {
					return fmt.Errorf("dangerous path detected in archive: %s", hdr.Name)
				}

				if hdr.Name == "manifest.json" {
					manifestBytes, err = io.ReadAll(tr)
					if err != nil {
						return err
					}
				} else if hdr.Name == "model_name.txt" {
					nameBytes, err := io.ReadAll(tr)
					if err != nil {
						return err
					}
					modelNameStr = string(nameBytes)
				} else if strings.HasPrefix(hdr.Name, "blobs/") {
					digest := strings.TrimPrefix(hdr.Name, "blobs/")
					digest = strings.ReplaceAll(digest, "-", ":")

					blobPath, err := manifest.BlobsPath(digest)
					if err != nil {
						return err
					}

					// Verify target file path lies strictly within blobs subdirectory
					modelsDir := envconfig.Models()
					absBlobsDir, err := filepath.Abs(filepath.Join(modelsDir, "blobs"))
					if err != nil {
						return err
					}
					absBlobPath, err := filepath.Abs(blobPath)
					if err != nil {
						return err
					}
					if !strings.HasPrefix(absBlobPath, absBlobsDir) {
						return fmt.Errorf("dangerous blob path detected: %s", blobPath)
					}

					fmt.Printf("Extracting blob %s...\n", digest[:12])
					outBlob, err := os.Create(blobPath)
					if err != nil {
						return err
					}

					hasher := sha256.New()
					mw := io.MultiWriter(outBlob, hasher)

					if _, err := io.Copy(mw, tr); err != nil {
						outBlob.Close()
						_ = os.Remove(blobPath)
						return err
					}
					outBlob.Close()

					// Check SHA256 integrity
					actualDigest := "sha256:" + hex.EncodeToString(hasher.Sum(nil))
					if !strings.EqualFold(actualDigest, digest) {
						_ = os.Remove(blobPath)
						return fmt.Errorf("integrity violation for %s: expected digest %s, got %s", hdr.Name, digest, actualDigest)
					}
				}
			}

			if len(manifestBytes) == 0 || modelNameStr == "" {
				return fmt.Errorf("corrupted model archive: missing metadata")
			}

			name := model.ParseNameFromFilepath(filepath.FromSlash(modelNameStr))
			if !name.IsFullyQualified() {
				name = model.ParseName(modelNameStr)
			}

			manifestPath, err := manifest.PathForName(name)
			if err != nil {
				return err
			}

			if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
				return err
			}

			if err := os.WriteFile(manifestPath, manifestBytes, 0644); err != nil {
				return err
			}

			fmt.Printf("✅ Model %s successfully imported!\n", name.String())
			return nil
		},
	}
	return cmd
}
