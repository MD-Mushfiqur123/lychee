package server

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/lychee/lychee/api"
	"github.com/lychee/lychee/envconfig"
	"github.com/lychee/lychee/format"
	"github.com/lychee/lychee/fs/ggml"
	internalcloud "github.com/lychee/lychee/internal/cloud"
	"github.com/lychee/lychee/llm"
	"github.com/lychee/lychee/manifest"
	"github.com/lychee/lychee/types/errtypes"
	"github.com/lychee/lychee/types/model"
	imagegenmanifest "github.com/lychee/lychee/x/imagegen/manifest"
	xserver "github.com/lychee/lychee/x/server"
)

func (s *Server) PullHandler(c *gin.Context) {
	var req api.PullRequest
	err := c.ShouldBindJSON(&req)
	switch {
	case errors.Is(err, io.EOF):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	case err != nil:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TEMP(drifkin): we're temporarily allowing to continue pulling cloud model
	// stub-files until we integrate cloud models into `/api/tags` (in which case
	// this roundabout way of "adding" cloud models won't be needed anymore). So
	// right here normalize any `:cloud` models into the legacy-style suffixes
	// `:<tag>-cloud` and `:cloud`
	modelRef, err := parseNormalizePullModelRef(cmp.Or(req.Model, req.Name))
	if err != nil {
		writeModelRefParseError(c, err, http.StatusBadRequest, errtypes.InvalidModelNameErrMsg)
		return
	}

	name := modelRef.Name

	name, err = getExistingName(name)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch := make(chan any)
	go func() {
		defer close(ch)
		fn := func(r api.ProgressResponse) {
			ch <- r
		}

		regOpts := &registryOptions{
			Insecure: req.Insecure,
		}

		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()

		if err := PullModel(ctx, name.DisplayShortest(), regOpts, fn); err != nil {
			ch <- gin.H{"error": err.Error()}
			return
		}

		s.refreshModelListCache(name)
	}()

	if req.Stream != nil && !*req.Stream {
		waitForStream(c, ch)
		return
	}

	streamResponse(c, ch)
}

func (s *Server) PushHandler(c *gin.Context) {
	var req api.PushRequest
	err := c.ShouldBindJSON(&req)
	switch {
	case errors.Is(err, io.EOF):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	case err != nil:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var mname string
	if req.Model != "" {
		mname = req.Model
	} else if req.Name != "" {
		mname = req.Name
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	ch := make(chan any)
	go func() {
		defer close(ch)
		fn := func(r api.ProgressResponse) {
			ch <- r
		}

		regOpts := &registryOptions{
			Insecure: req.Insecure,
		}

		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()

		name, err := getExistingName(model.ParseName(mname))
		if err != nil {
			ch <- gin.H{"error": err.Error()}
			return
		}

		if err := PushModel(ctx, name.DisplayShortest(), regOpts, fn); err != nil {
			ch <- gin.H{"error": err.Error()}
		}
	}()

	if req.Stream != nil && !*req.Stream {
		waitForStream(c, ch)
		return
	}

	streamResponse(c, ch)
}

// getExistingName searches the models directory for the longest prefix match of
// the input name and returns the input name with all existing parts replaced
// with each part found. If no parts are found, the input name is returned as
// is.
func getExistingName(n model.Name) (model.Name, error) {
	var zero model.Name
	existing, err := manifest.Manifests(true)
	if err != nil {
		return zero, err
	}
	var set model.Name // tracks parts already canonicalized
	for e := range existing {
		if set.Host == "" && strings.EqualFold(e.Host, n.Host) {
			n.Host = e.Host
		}
		if set.Namespace == "" && strings.EqualFold(e.Namespace, n.Namespace) {
			n.Namespace = e.Namespace
		}
		if set.Model == "" && strings.EqualFold(e.Model, n.Model) {
			n.Model = e.Model
		}
		if set.Tag == "" && strings.EqualFold(e.Tag, n.Tag) {
			n.Tag = e.Tag
		}
	}

	return n, nil
}

func (s *Server) DeleteHandler(c *gin.Context) {
	var r api.DeleteRequest
	if err := c.ShouldBindJSON(&r); errors.Is(err, io.EOF) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	modelRef, err := parseNormalizePullModelRef(cmp.Or(r.Model, r.Name))
	if err != nil {
		switch {
		case errors.Is(err, errConflictingModelSource):
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, model.ErrUnqualifiedName):
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("name %q is invalid", cmp.Or(r.Model, r.Name))})
		default:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	n, err := getExistingName(modelRef.Name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", cmp.Or(r.Model, r.Name))})
		return
	}

	m, err := manifest.ParseNamedManifest(n)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", cmp.Or(r.Model, r.Name))})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := m.Remove(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.deleteModelListCache(n)

	if err := m.RemoveLayers(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func (s *Server) ShowHandler(c *gin.Context) {
	var req api.ShowRequest
	err := c.ShouldBindJSON(&req)
	switch {
	case errors.Is(err, io.EOF):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	case err != nil:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Model != "" {
		// noop
	} else if req.Name != "" {
		req.Model = req.Name
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}
	requestedModel := req.Model

	modelRef, err := parseAndValidateModelRef(req.Model)
	if err != nil {
		writeModelRefParseError(c, err, http.StatusBadRequest, err.Error())
		return
	}

	if modelRef.Source == modelSourceCloud {
		req.Model = modelRef.Base
		if modelShowCacheable(req) && s.modelCaches != nil && s.modelCaches.show != nil {
			if disabled, _ := internalcloud.Status(); disabled {
				c.JSON(http.StatusForbidden, gin.H{"error": internalcloud.DisabledError(cloudErrRemoteModelDetailsUnavailable)})
				return
			}

			ctx := context.Background()
			if c.Request != nil {
				ctx = c.Request.Context()
			}
			if resp, ok := s.modelCaches.show.GetCloudSWR(ctx, req); ok {
				c.JSON(http.StatusOK, resp)
				return
			}
		}
		proxyCloudJSONRequest(c, req, cloudErrRemoteModelDetailsUnavailable)
		return
	}

	name := modelRef.Name
	name, err = getExistingName(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Model = name.DisplayShortest()

	var resp *api.ShowResponse
	if modelShowCacheable(req) && s.modelCaches != nil && s.modelCaches.show != nil {
		resp, err = s.modelCaches.show.GetLocal(req)
	} else {
		resp, err = GetModelInfo(req)
	}
	if err != nil {
		var statusErr api.StatusError
		switch {
		case os.IsNotExist(err):
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", req.Model)})
		case errors.As(err, &statusErr):
			c.JSON(statusErr.StatusCode, gin.H{"error": statusErr.ErrorMessage})
		case err.Error() == errtypes.InvalidModelNameErrMsg:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if modelRef.Source == modelSourceLocal && resp.RemoteHost != "" {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", modelRef.Original)})
		return
	}

	userAgent := c.Request.UserAgent()
	if strings.HasPrefix(userAgent, copilotChatUserAgentPrefix) {
		if resp.ModelInfo == nil {
			resp.ModelInfo = map[string]any{}
		}
		// Copilot Chat prefers `general.basename`, but this is usually not what
		// users are familiar with, so echo back the requested model name.
		resp.ModelInfo["general.basename"] = requestedModel
	}

	c.JSON(http.StatusOK, resp)
}

func GetModelInfo(req api.ShowRequest) (*api.ShowResponse, error) {
	name := model.ParseName(req.Model)
	if !name.IsValid() {
		return nil, model.Unqualified(name)
	}
	name, err := getExistingName(name)
	if err != nil {
		return nil, err
	}

	m, err := GetModel(name.String())
	if err != nil {
		return nil, err
	}

	if m.Config.RemoteHost != "" {
		if disabled, _ := internalcloud.Status(); disabled {
			return nil, api.StatusError{
				StatusCode:   http.StatusForbidden,
				ErrorMessage: internalcloud.DisabledError(cloudErrRemoteModelDetailsUnavailable),
			}
		}
	}

	modelDetails := api.ModelDetails{
		ParentModel:       m.ParentModel,
		Format:            m.Config.ModelFormat,
		Family:            m.Config.ModelFamily,
		Families:          m.Config.ModelFamilies,
		ParameterSize:     m.Config.ModelType,
		QuantizationLevel: m.Config.FileType,
	}

	// For image generation models, populate details from imagegen package
	if slices.Contains(m.Capabilities(), model.CapabilityImage) {
		if info, err := imagegenmanifest.GetModelInfo(name.String()); err == nil {
			modelDetails.Family = info.Architecture
			modelDetails.ParameterSize = format.HumanNumber(uint64(info.ParameterCount))
			modelDetails.QuantizationLevel = info.Quantization
		}
	}

	// For safetensors LLM models (experimental), populate details from config.json
	if m.Config.ModelFormat == "safetensors" && slices.Contains(m.Config.Capabilities, "completion") {
		if info, err := xserver.GetSafetensorsLLMInfo(name); err == nil {
			if arch, ok := info["general.architecture"].(string); ok && arch != "" {
				modelDetails.Family = arch
			}
			if paramCount, ok := info["general.parameter_count"].(int64); ok && paramCount > 0 {
				modelDetails.ParameterSize = format.HumanNumber(uint64(paramCount))
			}
		}
		// Older manifests may not have file_type populated for safetensors models.
		if modelDetails.QuantizationLevel == "" {
			if dtype, err := xserver.GetSafetensorsDtype(name); err == nil && dtype != "" {
				modelDetails.QuantizationLevel = dtype
			}
		}
	}

	if req.System != "" {
		m.System = req.System
	}

	msgs := make([]api.Message, len(m.Messages))
	for i, msg := range m.Messages {
		msgs[i] = api.Message{Role: msg.Role, Content: msg.Content}
	}

	mf, err := manifest.ParseNamedManifest(name)
	if err != nil {
		return nil, err
	}

	resp := &api.ShowResponse{
		License:      strings.Join(m.License, "\n"),
		System:       m.System,
		Template:     m.Template.String(),
		Details:      modelDetails,
		Messages:     msgs,
		Capabilities: m.Capabilities(),
		ModifiedAt:   mf.FileInfo().ModTime(),
		Requires:     m.Config.Requires,
		// Several integrations crash on a nil/omitempty+empty ModelInfo, so by
		// default we return an empty map.
		ModelInfo: make(map[string]any),
	}

	if m.Config.RemoteHost != "" {
		resp.RemoteHost = m.Config.RemoteHost
		resp.RemoteModel = m.Config.RemoteModel

		if m.Config.ModelFamily != "" {
			resp.ModelInfo = make(map[string]any)
			resp.ModelInfo["general.architecture"] = m.Config.ModelFamily

			if m.Config.BaseName != "" {
				resp.ModelInfo["general.basename"] = m.Config.BaseName
			}

			if m.Config.ContextLen > 0 {
				resp.ModelInfo[fmt.Sprintf("%s.context_length", m.Config.ModelFamily)] = m.Config.ContextLen
			}

			if m.Config.EmbedLen > 0 {
				resp.ModelInfo[fmt.Sprintf("%s.embedding_length", m.Config.ModelFamily)] = m.Config.EmbedLen
			}
		}
	}

	var params []string
	cs := 30
	for k, v := range m.Options {
		switch val := v.(type) {
		case []any:
			for _, nv := range val {
				params = append(params, fmt.Sprintf("%-*s %#v", cs, k, nv))
			}
		default:
			params = append(params, fmt.Sprintf("%-*s %#v", cs, k, v))
		}
	}
	resp.Parameters = strings.Join(params, "\n")

	if len(req.Options) > 0 {
		if m.Options == nil {
			m.Options = make(map[string]any)
		}
		for k, v := range req.Options {
			m.Options[k] = v
		}
	}

	var sb strings.Builder
	fmt.Fprintln(&sb, "# Modelfile generated by \"lychee show\"")
	modelfile := m.String()
	if m.IsMLX() {
		fmt.Fprintf(&sb, "FROM %s\n", m.ShortName)
		if _, rest, ok := strings.Cut(modelfile, "\n"); ok {
			fmt.Fprint(&sb, rest)
		}
	} else {
		fmt.Fprintln(&sb, "# To build a new Modelfile based on this, replace FROM with:")
		fmt.Fprintf(&sb, "# FROM %s\n\n", m.ShortName)
		fmt.Fprint(&sb, modelfile)
	}
	resp.Modelfile = sb.String()

	// skip loading tensor information if this is a remote model
	if m.Config.RemoteHost != "" && m.Config.RemoteModel != "" {
		return resp, nil
	}

	if slices.Contains(m.Capabilities(), model.CapabilityImage) {
		// Populate tensor info if verbose
		if req.Verbose {
			if tensors, err := xserver.GetSafetensorsTensorInfo(name); err == nil {
				resp.Tensors = tensors
			}
		}
		return resp, nil
	}

	// For safetensors LLM models (experimental), populate ModelInfo from config.json
	if m.Config.ModelFormat == "safetensors" && slices.Contains(m.Config.Capabilities, "completion") {
		if info, err := xserver.GetSafetensorsLLMInfo(name); err == nil {
			resp.ModelInfo = info
		}
		// Populate tensor info if verbose
		if req.Verbose {
			if tensors, err := xserver.GetSafetensorsTensorInfo(name); err == nil {
				resp.Tensors = tensors
			}
		}
		return resp, nil
	}

	kvData, tensors, err := getModelData(m.ModelPath, req.Verbose)
	if err != nil {
		return nil, err
	}

	resp.Template = selectedModelTemplate(m, kvData)
	if isUnknownQuantization(resp.Details.QuantizationLevel) {
		if fileType := kvData.FileType().String(); !isUnknownQuantization(fileType) {
			resp.Details.QuantizationLevel = fileType
		}
	}

	delete(kvData, "general.name")
	delete(kvData, "tokenizer.chat_template")
	resp.ModelInfo = kvData

	tensorData := make([]api.Tensor, len(tensors.Items()))
	for cnt, t := range tensors.Items() {
		tensorData[cnt] = api.Tensor{Name: t.Name, Type: t.Type(), Shape: t.Shape}
	}
	resp.Tensors = tensorData

	if len(m.ProjectorPaths) > 0 {
		projectorData, _, err := getModelData(m.ProjectorPaths[0], req.Verbose)
		if err != nil {
			return nil, err
		}
		resp.ProjectorInfo = projectorData
	}

	return resp, nil
}

func getModelData(digest string, verbose bool) (ggml.KV, ggml.Tensors, error) {
	maxArraySize := 0
	if verbose {
		maxArraySize = -1
	}
	data, err := llm.LoadModel(digest, maxArraySize)
	if err != nil {
		return nil, ggml.Tensors{}, err
	}

	kv := data.KV()

	if !verbose {
		for k := range kv {
			if t, ok := kv[k].([]any); len(t) > 5 && ok {
				kv[k] = []any{}
			}
		}
	}

	return kv, data.Tensors(), nil
}

func selectedModelTemplate(m *Model, kv ggml.KV) string {
	if m.HasChatTemplate && chatModeForModel(m) == chatExecutionModeNative {
		if chatTemplate := kv.String("tokenizer.chat_template"); chatTemplate != "" {
			return chatTemplate
		}
	}
	return m.Template.String()
}

func (s *Server) ListHandler(c *gin.Context) {
	if s.modelCaches == nil || s.modelCaches.modelList == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model list cache unavailable"})
		return
	}

	models, err := s.modelCaches.modelList.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, api.ListResponse{Models: models})
}

func (s *Server) CopyHandler(c *gin.Context) {
	var r api.CopyRequest
	if err := c.ShouldBindJSON(&r); errors.Is(err, io.EOF) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	src := model.ParseName(r.Source)
	if !src.IsValid() {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("source %q is invalid", r.Source)})
		return
	}
	src, err := getExistingName(src)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dst := model.ParseName(r.Destination)
	if !dst.IsValid() {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("destination %q is invalid", r.Destination)})
		return
	}
	dst, err = getExistingName(dst)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := CopyModel(src, dst); errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model %q not found", r.Source)})
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		s.refreshModelListCache(dst)
	}
}

func (s *Server) HeadBlobHandler(c *gin.Context) {
	path, err := manifest.BlobsPath(c.Param("digest"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := os.Stat(path); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("blob %q not found", c.Param("digest"))})
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) CreateBlobHandler(c *gin.Context) {
	if ib, ok := intermediateBlobs[c.Param("digest")]; ok {
		p, err := manifest.BlobsPath(ib)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
			slog.Info("evicting intermediate blob which no longer exists", "digest", ib)
			delete(intermediateBlobs, c.Param("digest"))
		} else if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.Status(http.StatusOK)
			return
		}
	}

	path, err := manifest.BlobsPath(c.Param("digest"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = os.Stat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// noop
	case err != nil:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	default:
		c.Status(http.StatusOK)
		return
	}

	layer, err := manifest.NewLayer(c.Request.Body, "")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if layer.Digest != c.Param("digest") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("digest mismatch, expected %q, got %q", c.Param("digest"), layer.Digest)})
		return
	}

	c.Status(http.StatusCreated)
}
