//go:build darwin

package imagegen

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/lychee/lychee/x/imagegen/manifest"
	"github.com/lychee/lychee/x/imagegen/mlx"
	"github.com/lychee/lychee/x/imagegen/models/flux2"
	"github.com/lychee/lychee/x/imagegen/models/zimage"
)

// ImageModel is the interface for image generation models.
type ImageModel interface {
	GenerateImage(ctx context.Context, prompt string, width, height int32, steps int, seed int64, progress func(step, total int)) (*mlx.Array, error)
}

var imageGenMu sync.Mutex

// loadImageModel loads an image generation model.
func (s *server) loadImageModel() error {
	// Check memory requirements before loading
	var requiredMemory uint64
	if modelManifest, err := manifest.LoadManifest(s.modelName); err == nil {
		requiredMemory = uint64(modelManifest.TotalTensorSize())
	}
	availableMemory := mlx.GetMemoryLimit()
	if availableMemory > 0 && requiredMemory > 0 && availableMemory < requiredMemory {
		return fmt.Errorf("insufficient memory for image generation: need %d GB, have %d GB",
			requiredMemory/(1024*1024*1024), availableMemory/(1024*1024*1024))
	}

	// Detect model type and load appropriate model
	modelType := DetectModelType(s.modelName)
	slog.Info("detected image model type", "type", modelType)

	var model ImageModel
	switch modelType {
	case "Flux2KleinPipeline":
		m := &flux2.Model{}
		if err := m.Load(s.modelName); err != nil {
			return fmt.Errorf("failed to load flux2 model: %w", err)
		}
		model = m
	default:
		// Default to Z-Image for ZImagePipeline, FluxPipeline, etc.
		m := &zimage.Model{}
		if err := m.Load(s.modelName); err != nil {
			return fmt.Errorf("failed to load zimage model: %w", err)
		}
		model = m
	}

	s.imageModel = model
	return nil
}

// handleImageCompletion handles image generation requests.
func (s *server) handleImageCompletion(w http.ResponseWriter, r *http.Request, req Request) {
	// Serialize generation requests - MLX model may not handle concurrent generation
	imageGenMu.Lock()
	defer imageGenMu.Unlock()

	// Set seed if not provided
	if req.Seed <= 0 {
		req.Seed = time.Now().UnixNano()
	}

	// Set up streaming response
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Transfer-Encoding", "chunked")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	enc := json.NewEncoder(w)

	// Progress callback streams step updates
	progress := func(step, total int) {
		resp := Response{Step: step, Total: total}
		if err := enc.Encode(resp); err != nil {
			slog.Error("failed to encode progress response", "error", err)
			return
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			slog.Error("failed to write progress newline", "error", err)
			return
		}
		flusher.Flush()
	}

	// Generate image
	img, err := s.imageModel.GenerateImage(ctx, req.Prompt, req.Width, req.Height, req.Steps, req.Seed, progress)
	if err != nil {
		// Don't send error for cancellation
		if ctx.Err() != nil {
			return
		}
		resp := Response{Content: fmt.Sprintf("error: %v", err), Done: true}
		data, err := json.Marshal(resp)
		if err != nil {
			slog.Error("failed to marshal error response", "error", err)
			return
		}
		if _, err := w.Write(data); err != nil {
			slog.Error("failed to write error response", "error", err)
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			slog.Error("failed to write error newline", "error", err)
		}
		return
	}

	// Encode image as base64 PNG
	imageData, err := EncodeImageBase64(img)
	if err != nil {
		resp := Response{Content: fmt.Sprintf("error encoding: %v", err), Done: true}
		data, err := json.Marshal(resp)
		if err != nil {
			slog.Error("failed to marshal encode error response", "error", err)
			return
		}
		if _, err := w.Write(data); err != nil {
			slog.Error("failed to write encode error response", "error", err)
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			slog.Error("failed to write encode error newline", "error", err)
		}
		return
	}

	// Free the generated image array and clean up MLX state
	img.Free()
	mlx.ClearCache()
	mlx.MetalResetPeakMemory()

	// Send final response with image data
	resp := Response{
		Image: imageData,
		Done:  true,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		slog.Error("failed to marshal final response", "error", err)
		return
	}
	if _, err := w.Write(data); err != nil {
		slog.Error("failed to write final response", "error", err)
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		slog.Error("failed to write final response newline", "error", err)
	}
	flusher.Flush()
}
