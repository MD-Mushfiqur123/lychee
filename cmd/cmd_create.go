package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/progress"
	"github.com/MD-Mushfiqur123/lychee/parser"
	"github.com/MD-Mushfiqur123/lychee/types/model"
	"github.com/MD-Mushfiqur123/lychee/types/syncmap"
	xcreateclient "github.com/MD-Mushfiqur123/lychee/x/create/client"
)

func CreateHandler(cmd *cobra.Command, args []string) error {
	p := progress.NewProgress(os.Stderr)
	defer p.Stop()

	// Validate model name early to fail fast
	modelName := args[0]
	name := model.ParseName(modelName)
	if !name.IsValid() {
		return fmt.Errorf("invalid model name: %s", modelName)
	}

	// Check for --experimental flag for safetensors model creation
	// This gates both safetensors LLM and imagegen model creation
	experimental, _ := cmd.Flags().GetBool("experimental")
	draftQuantize, _ := cmd.Flags().GetString("draft-quantize")
	if experimental {
		if !isLocalhost() {
			return errors.New("remote safetensor model creation not yet supported")
		}

		// Get Modelfile content - either from -f flag or default to "FROM ."
		var reader io.Reader
		filename, err := getModelfileName(cmd)
		if os.IsNotExist(err) || filename == "" {
			// No Modelfile specified or found - use default
			reader = strings.NewReader("FROM .\n")
		} else if err != nil {
			return err
		} else {
			f, err := os.Open(filename)
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()
			reader = f
		}

		// Parse the Modelfile
		modelfile, err := parser.ParseFile(reader)
		if err != nil {
			return fmt.Errorf("failed to parse Modelfile: %w", err)
		}

		modelDir, mfConfig, err := xcreateclient.ConfigFromModelfile(modelfile)
		if err != nil {
			return err
		}

		modelDir = resolveExperimentalLocalModelDir(modelDir, filename)
		if mfConfig.Draft != "" {
			draftDir, err := resolveExperimentalDraftDir(mfConfig.Draft, filename)
			if err != nil {
				return err
			}
			mfConfig.Draft = draftDir
		}

		quantize, _ := cmd.Flags().GetString("quantize")
		return xcreateclient.CreateModel(xcreateclient.CreateOptions{
			ModelName:     modelName,
			ModelDir:      modelDir,
			Quantize:      quantize,
			DraftQuantize: draftQuantize,
			Modelfile:     mfConfig,
		}, p)
	}

	// Standard Modelfile + API path
	var reader io.Reader

	filename, err := getModelfileName(cmd)
	if os.IsNotExist(err) {
		if filename == "" {
			reader = strings.NewReader("FROM .\n")
		} else {
			return errModelfileNotFound
		}
	} else if err != nil {
		return err
	} else {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}

		reader = f
		defer func() { _ = f.Close() }()
	}

	modelfile, err := parser.ParseFile(reader)
	if err != nil {
		return err
	}

	status := "gathering model components"
	spinner := progress.NewSpinner(status)
	p.Add(status, spinner)

	req, err := modelfile.CreateRequest(filepath.Dir(filename))
	if err != nil {
		return err
	}
	spinner.Stop()

	req.Model = modelName
	quantize, _ := cmd.Flags().GetString("quantize")
	if quantize != "" {
		req.Quantize = quantize
	}
	if draftQuantize != "" {
		if len(req.DraftFiles) == 0 {
			return errors.New("--draft-quantize requires a DRAFT model")
		}
		req.DraftQuantize = draftQuantize
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	var g errgroup.Group
	g.SetLimit(max(runtime.GOMAXPROCS(0)-1, 1))

	files := syncmap.NewSyncMap[string, string]()
	fileNames := createRequestFileNames(req.Files)
	for f, digest := range req.Files {
		g.Go(func() error {
			if _, err := createBlob(cmd, client, f, digest, p); err != nil {
				return err
			}

			files.Store(fileNames[f], digest)
			return nil
		})
	}

	adapters := syncmap.NewSyncMap[string, string]()
	adapterNames := createRequestFileNames(req.Adapters)
	for f, digest := range req.Adapters {
		g.Go(func() error {
			if _, err := createBlob(cmd, client, f, digest, p); err != nil {
				return err
			}

			adapters.Store(adapterNames[f], digest)
			return nil
		})
	}

	draftFiles := syncmap.NewSyncMap[string, string]()
	draftFileNames := createRequestFileNames(req.DraftFiles)
	for f, digest := range req.DraftFiles {
		g.Go(func() error {
			if _, err := createBlob(cmd, client, f, digest, p); err != nil {
				return err
			}

			draftFiles.Store(draftFileNames[f], digest)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	req.Files = files.Items()
	req.Adapters = adapters.Items()
	req.DraftFiles = draftFiles.Items()

	bars := make(map[string]*progress.Bar)
	fn := func(resp api.ProgressResponse) error {
		if resp.Digest != "" {
			bar, ok := bars[resp.Digest]
			if !ok {
				msg := resp.Status
				if msg == "" {
					msg = fmt.Sprintf("pulling %s...", resp.Digest[7:19])
				}
				bar = progress.NewBar(msg, resp.Total, resp.Completed)
				bars[resp.Digest] = bar
				p.Add(resp.Digest, bar)
			}

			bar.Set(resp.Completed)
		} else if status != resp.Status {
			spinner.Stop()

			status = resp.Status
			spinner = progress.NewSpinner(status)
			p.Add(status, spinner)
		}

		return nil
	}

	if err := client.Create(cmd.Context(), req, fn); err != nil {
		if strings.Contains(err.Error(), "path or Modelfile are required") {
			return fmt.Errorf("the lychee server must be updated to use `lychee create` with this client")
		}
		return err
	}

	return nil
}

func createRequestFileNames(files map[string]string) map[string]string {
	names := make(map[string]string, len(files))
	root, ok := commonFileRoot(files)
	for f := range files {
		name := filepath.Base(f)
		if ok {
			abs, err := filepath.Abs(f)
			if err == nil {
				if rel, err := filepath.Rel(root, abs); err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
					name = rel
				}
			}
		}
		names[f] = path.Clean(filepath.ToSlash(name))
	}
	return names
}

func commonFileRoot(files map[string]string) (string, bool) {
	if len(files) < 2 {
		return "", false
	}

	var root string
	var volume string
	for f := range files {
		abs, err := filepath.Abs(f)
		if err != nil {
			return "", false
		}
		if nextVolume := filepath.VolumeName(abs); volume == "" {
			volume = nextVolume
		} else if !strings.EqualFold(volume, nextVolume) {
			return "", false
		}

		dir := filepath.Dir(abs)
		if root == "" {
			root = dir
			continue
		}

		for {
			rel, err := filepath.Rel(root, dir)
			if err == nil && (rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))) {
				break
			}

			parent := filepath.Dir(root)
			if parent == root {
				return "", false
			}
			root = parent
		}
	}

	return root, root != ""
}

func createBlob(cmd *cobra.Command, client *api.Client, path string, digest string, p *progress.Progress) (string, error) {
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}

	bin, err := os.Open(realPath)
	if err != nil {
		return "", err
	}
	defer func() { _ = bin.Close() }()

	// Get file info to retrieve the size
	fileInfo, err := bin.Stat()
	if err != nil {
		return "", err
	}
	fileSize := fileInfo.Size()

	var pw progressWriter
	status := fmt.Sprintf("copying file %s 0%%", digest)
	spinner := progress.NewSpinner(status)
	p.Add(status, spinner)
	defer spinner.Stop()

	done := make(chan struct{})
	defer close(done)

	go func() {
		ticker := time.NewTicker(60 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				spinner.SetMessage(fmt.Sprintf("copying file %s %d%%", digest, int(100*pw.n.Load()/fileSize)))
			case <-done:
				spinner.SetMessage(fmt.Sprintf("copying file %s 100%%", digest))
				return
			}
		}
	}()

	if err := client.CreateBlob(cmd.Context(), digest, io.TeeReader(bin, &pw)); err != nil {
		return "", err
	}
	return digest, nil
}

type progressWriter struct {
	n atomic.Int64
}

func (w *progressWriter) Write(p []byte) (n int, err error) {
	w.n.Add(int64(len(p)))
	return len(p), nil
}
