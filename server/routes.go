package server

import (
	"bytes"
	"cmp"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/image/webp"

	"github.com/lychee/lychee/api"
	"github.com/lychee/lychee/auth"
	"github.com/lychee/lychee/discover"
	"github.com/lychee/lychee/envconfig"
	"github.com/lychee/lychee/format"
	internalcloud "github.com/lychee/lychee/internal/cloud"
	"github.com/lychee/lychee/llm"
	"github.com/lychee/lychee/logutil"
	"github.com/lychee/lychee/manifest"
	"github.com/lychee/lychee/model/parsers"
	"github.com/lychee/lychee/model/renderers"
	"github.com/lychee/lychee/server/internal/client/lychee"
	"github.com/lychee/lychee/server/internal/registry"
	"github.com/lychee/lychee/thinking"
	"github.com/lychee/lychee/tools"
	"github.com/lychee/lychee/types/errtypes"
	"github.com/lychee/lychee/types/model"
	"github.com/lychee/lychee/version"
)

const signinURLStr = "https://lychee.github.io/lychee/connect?name=%s&key=%s"

const (
	cloudErrRemoteInferenceUnavailable    = "remote model is unavailable"
	cloudErrRemoteModelDetailsUnavailable = "remote model details are unavailable"
	cloudErrWebSearchUnavailable          = "web search is unavailable"
	cloudErrWebFetchUnavailable           = "web fetch is unavailable"
	copilotChatUserAgentPrefix            = "GitHubCopilotChat/"
)

func writeModelRefParseError(c *gin.Context, err error, fallbackStatus int, fallbackMessage string) {
	switch {
	case errors.Is(err, errConflictingModelSource):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrUnqualifiedName):
		c.JSON(http.StatusBadRequest, gin.H{"error": errtypes.InvalidModelNameErrMsg})
	default:
		c.JSON(fallbackStatus, gin.H{"error": fallbackMessage})
	}
}

func shouldUseHarmony(model *Model) bool {
	if slices.Contains([]string{"gptoss", "gpt-oss"}, model.Config.ModelFamily) {
		// heuristic to check whether the template expects to be parsed via harmony:
		// search for harmony tags that are nearly always used
		if model.Template.Contains("<|start|>") && model.Template.Contains("<|end|>") {
			return true
		}
	}

	return false
}

func experimentEnabled(name string) bool {
	return slices.Contains(strings.Split(os.Getenv("LYCHEE_EXPERIMENT"), ","), name)
}

var useClient2 = experimentEnabled("client2")

var mode string = gin.DebugMode

type Server struct {
	addr          net.Addr
	sched         *Scheduler
	defaultNumCtx int
	requestLogger *inferenceRequestLogger
	modelCaches   *modelCaches
	memoryStore   *MemoryStore
	modelRouter   *ModelRouter
	modelAliases  *ModelAliases
}

func init() {
	switch mode {
	case gin.DebugMode:
	case gin.ReleaseMode:
	case gin.TestMode:
	default:
		mode = gin.DebugMode
	}

	gin.SetMode(mode)

	// Tell renderers to use [img] tags
	renderers.RenderImgTags = true
}

var (
	errRequired    = errors.New("is required")
	errBadTemplate = errors.New("template error")
)

func (s *Server) modelOptions(model *Model, requestOpts map[string]any) (api.Options, error) {
	return s.modelOptionsWithEmbeddingBatchDefault(model, requestOpts, shouldApplyEmbeddingBatchDefault(model, requestOpts))
}

func (s *Server) modelOptionsWithEmbeddingBatchDefault(model *Model, requestOpts map[string]any, applyEmbeddingBatchDefault bool) (api.Options, error) {
	opts := api.DefaultOptions()
	if opts.NumCtx == 0 {
		opts.NumCtx = s.defaultNumCtx
	}

	// api.Options stores defaulted values, so lower layers cannot distinguish
	// an unset draft_num_predict from the default. Track that while we still
	// have the raw model/request option maps.
	draftNumPredictSet := hasOption(requestOpts, "draft_num_predict")
	if model != nil {
		draftNumPredictSet = draftNumPredictSet || hasOption(model.Options, "draft_num_predict")
		if err := opts.FromMap(model.Options); err != nil {
			return api.Options{}, err
		}
	}

	if err := opts.FromMap(requestOpts); err != nil {
		return api.Options{}, err
	}

	if applyEmbeddingBatchDefault {
		opts = llm.WithDefaultEmbeddingNumBatch(opts)
	}

	if model != nil && model.DraftPath == "" && !draftNumPredictSet {
		opts.DraftNumPredict = 0
	}

	return opts, nil
}

func shouldApplyEmbeddingBatchDefault(m *Model, requestOpts map[string]any) bool {
	if m == nil || hasOption(m.Options, "num_batch") || hasOption(requestOpts, "num_batch") {
		return false
	}
	if slices.Contains(m.Config.Capabilities, string(model.CapabilityEmbedding)) {
		return true
	}
	return m.ModelPath != "" && m.CheckCapabilities(model.CapabilityEmbedding) == nil
}

func hasOption(opts map[string]any, name string) bool {
	_, ok := opts[name]
	return ok
}

func usesAutomaticNumCtx(model *Model, requestOpts map[string]any) bool {
	if _, ok := requestOpts["num_ctx"]; ok {
		return false
	}
	if model != nil {
		if _, ok := model.Options["num_ctx"]; ok {
			return false
		}
	}
	return envconfig.ContextLength() == 0
}

func usesAutomaticNumBatch(model *Model, requestOpts map[string]any) bool {
	if _, ok := requestOpts["num_batch"]; ok {
		return false
	}
	if model != nil {
		if _, ok := model.Options["num_batch"]; ok {
			return false
		}
	}
	return true
}

// scheduleRunner schedules a runner after validating inputs such as capabilities and model options.
// It returns the allocated runner, model instance, and consolidated options if successful and error otherwise.
func (s *Server) scheduleRunner(ctx context.Context, name string, caps []model.Capability, requestOpts map[string]any, keepAlive *api.Duration, shift *bool) (llm.LlamaServer, *Model, *api.Options, error) {
	if name == "" {
		return nil, nil, nil, fmt.Errorf("model %w", errRequired)
	}

	model, err := GetModel(name)
	if err != nil {
		return nil, nil, nil, err
	}

	if slices.Contains(model.Config.ModelFamilies, "mllama") && len(model.ProjectorPaths) > 0 {
		return nil, nil, nil, fmt.Errorf("'llama3.2-vision' is no longer compatible with your version of Lychee and has been replaced by a newer version. To re-download, run 'lychee pull llama3.2-vision'")
	}

	if err := model.CheckCapabilities(caps...); err != nil {
		return nil, nil, nil, fmt.Errorf("%s %w", name, err)
	}

	// Deprecated runner override option; ignore if present.
	delete(requestOpts, "use_imagegen_runner")

	numCtxAuto := usesAutomaticNumCtx(model, requestOpts)
	embeddingBatchDefault := shouldApplyEmbeddingBatchDefault(model, requestOpts)
	numBatchAuto := usesAutomaticNumBatch(model, requestOpts) && !embeddingBatchDefault
	opts, err := s.modelOptionsWithEmbeddingBatchDefault(model, requestOpts, embeddingBatchDefault)
	if err != nil {
		return nil, nil, nil, err
	}

	runnerCh, errCh := s.sched.getRunner(ctx, model, opts, keepAlive, numCtxAuto, numBatchAuto, shift)
	var runner *runnerRef
	select {
	case runner = <-runnerCh:
	case err = <-errCh:
		return nil, nil, nil, err
	}

	return runner.llama, model, &opts, nil
}

func signinURL() (string, error) {
	pubKey, err := auth.GetPublicKey()
	if err != nil {
		return "", err
	}

	encKey := base64.RawURLEncoding.EncodeToString([]byte(pubKey))
	h, _ := os.Hostname()
	return fmt.Sprintf(signinURLStr, url.PathEscape(h), encKey), nil
}







func isLocalIP(ip netip.Addr) bool {
	if interfaces, err := net.Interfaces(); err == nil {
		for _, iface := range interfaces {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			for _, a := range addrs {
				if parsed, _, err := net.ParseCIDR(a.String()); err == nil {
					if parsed.String() == ip.String() {
						return true
					}
				}
			}
		}
	}

	return false
}

func allowedHost(host string) bool {
	host = strings.ToLower(host)

	if host == "" || host == "localhost" {
		return true
	}

	if hostname, err := os.Hostname(); err == nil && host == strings.ToLower(hostname) {
		return true
	}

	tlds := []string{
		"localhost",
		"local",
		"internal",
	}

	// check if the host is a local TLD
	for _, tld := range tlds {
		if strings.HasSuffix(host, "."+tld) {
			return true
		}
	}

	return false
}

func allowedHostsMiddleware(addr net.Addr) gin.HandlerFunc {
	return func(c *gin.Context) {
		if addr == nil {
			c.Next()
			return
		}

		if addr, err := netip.ParseAddrPort(addr.String()); err == nil && !addr.Addr().IsLoopback() {
			c.Next()
			return
		}

		host, _, err := net.SplitHostPort(c.Request.Host)
		if err != nil {
			host = c.Request.Host
		}

		if addr, err := netip.ParseAddr(host); err == nil {
			if addr.IsLoopback() || addr.IsPrivate() || addr.IsUnspecified() || isLocalIP(addr) {
				c.Next()
				return
			}
		}

		if allowedHost(host) {
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}

			c.Next()
			return
		}

		c.AbortWithStatus(http.StatusForbidden)
	}
}

func (s *Server) GenerateRoutes(rc *lychee.Registry) (http.Handler, error) {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowWildcard = true
	corsConfig.AllowBrowserExtensions = true
	corsConfig.AllowHeaders = []string{
		"Authorization",
		"Content-Type",
		"User-Agent",
		"Accept",
		"X-Requested-With",

		// OpenAI compatibility headers
		"OpenAI-Beta",
		"x-stainless-arch",
		"x-stainless-async",
		"x-stainless-custom-poll-interval",
		"x-stainless-helper-method",
		"x-stainless-lang",
		"x-stainless-os",
		"x-stainless-package-version",
		"x-stainless-poll-helper",
		"x-stainless-retry-count",
		"x-stainless-runtime",
		"x-stainless-runtime-version",
		"x-stainless-timeout",
	}
	corsConfig.AllowOrigins = envconfig.AllowedOrigins()

	r := gin.Default()
	r.HandleMethodNotAllowed = true
	r.Use(
		cors.New(corsConfig),
		allowedHostsMiddleware(s.addr),
		func(c *gin.Context) {
			c.Header("X-Lychee-API-Version", "1.0.0")
			c.Next()
		},
	)

	// General
	r.HEAD("/", func(c *gin.Context) { c.String(http.StatusOK, "Lychee is running") })
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "Lychee is running") })
	r.HEAD("/api/version", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"version": version.Version}) })
	r.GET("/api/version", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"version": version.Version}) })
	r.GET("/api/status", s.StatusHandler)

	// Local model cache management
	s.registerModelRoutes(r)

	r.POST("/api/me", s.WhoamiHandler)

	r.POST("/api/signout", s.SignoutHandler)
	// deprecated
	r.DELETE("/api/user/keys/:encodedKey", s.SignoutHandler)

	// Create
	r.POST("/api/create", s.CreateHandler)
	r.POST("/api/blobs/:digest", s.CreateBlobHandler)
	r.HEAD("/api/blobs/:digest", s.HeadBlobHandler)
	r.POST("/api/copy", s.CopyHandler)
	r.POST("/api/experimental/web_search", s.WebSearchExperimentalHandler)
	r.POST("/api/experimental/web_fetch", s.WebFetchExperimentalHandler)
	r.GET("/api/experimental/model-recommendations", s.ModelRecommendationsExperimentalHandler)

	// Inference
	r.GET("/api/ps", s.PsHandler)
	r.POST("/api/generate", s.withInferenceRequestLogging("/api/generate", s.GenerateHandler)...)
	r.POST("/api/chat", s.withInferenceRequestLogging("/api/chat", s.ChatHandler)...)
	r.POST("/api/embed", s.EmbedHandler)
	r.POST("/api/embeddings", s.EmbeddingsHandler)
	r.POST("/api/compose", s.ComposeHandler)
	r.POST("/api/structured", s.StructuredHandler)
	r.GET("/api/conversations", s.ListConversationsHandler)
	r.GET("/api/conversations/:id", s.GetConversationHandler)
	r.POST("/api/conversations", s.CreateConversationHandler)
	r.DELETE("/api/conversations/:id", s.DeleteConversationHandler)
	r.GET("/api/conversations/:id/export", s.ExportConversationHandler)
	r.POST("/api/conversations/import", s.ImportConversationHandler)

	r.POST("/api/routes", s.CreateRouteHandler)
	r.GET("/api/routes", s.ListRoutesHandler)
	r.DELETE("/api/routes/:name", s.DeleteRouteHandler)
	r.GET("/api/routes/:name/status", s.RouteStatusHandler)

	r.POST("/api/aliases", s.SetAliasHandler)
	r.GET("/api/aliases", s.ListAliasesHandler)
	r.DELETE("/api/aliases/:name", s.DeleteAliasHandler)

	// OpenAI compatibility endpoints
	s.registerOpenAIRoutes(r)

	// Anthropic compatibility endpoints
	s.registerAnthropicRoutes(r)

	// Dashboard UI
	s.registerDashboardRoutes(r)

	if rc != nil {
		// wrap old with new
		rs := &registry.Local{
			Client:   rc,
			Logger:   slog.Default(), // TODO(bmizerany): Take a logger, do not use slog.Default()
			Fallback: r,

			Prune: PruneLayers,
		}
		return rs, nil
	}

	return r, nil
}

func (s *Server) ModelRecommendationsExperimentalHandler(c *gin.Context) {
	recs := defaultModelRecommendations
	source := "default"
	if s.modelCaches != nil && s.modelCaches.recommendations != nil {
		ctx := context.Background()
		if c.Request != nil {
			ctx = c.Request.Context()
		}
		recs = s.modelCaches.recommendations.GetSWR(ctx)
		source = "cache"
	}

	slog.Debug("serving model recommendations", "recommendation_source", source, "count", len(recs))
	c.JSON(http.StatusOK, api.ModelRecommendationsResponse{
		Recommendations: recs,
	})
}

func Serve(ln net.Listener) error {
	slog.SetDefault(logutil.NewLogger(os.Stderr, envconfig.LogLevel()))
	slog.Info("server config", "env", envconfig.Values())
	cloudDisabled, _ := internalcloud.Status()
	slog.Info(fmt.Sprintf("Lychee cloud disabled: %t", cloudDisabled))

	blobsDir, err := manifest.BlobsPath("")
	if err != nil {
		return err
	}
	if err := fixBlobs(blobsDir); err != nil {
		return err
	}

	if !envconfig.NoPrune() {
		if _, err := manifest.Manifests(false); err != nil {
			slog.Warn("corrupt manifests detected, skipping prune operation.  Re-pull or delete to clear", "error", err)
		} else {
			// clean up unused layers and manifests
			if err := PruneLayers(); err != nil {
				return err
			}

			manifestsPath, err := manifest.Path()
			if err != nil {
				return err
			}

			if err := manifest.PruneDirectory(manifestsPath); err != nil {
				return err
			}
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	s := &Server{
		addr:         ln.Addr(),
		modelCaches:  newModelCaches(),
		memoryStore:  NewMemoryStore(filepath.Join(home, ".lychee", "conversations")),
		modelRouter:  NewModelRouter(filepath.Join(home, ".lychee", "routes.json")),
		modelAliases: NewModelAliases(filepath.Join(home, ".lychee", "aliases.json")),
	}
	if s.modelRouter != nil {
		go s.modelRouter.StartHealthChecks(context.Background(), 30*time.Second)
	}
	if err := s.initRequestLogging(); err != nil {
		return err
	}

	var rc *lychee.Registry
	if useClient2 {
		var err error
		rc, err = lychee.DefaultRegistry()
		if err != nil {
			return err
		}
	}

	h, err := s.GenerateRoutes(rc)
	if err != nil {
		return err
	}

	http.Handle("/", h)

	ctx, done := context.WithCancel(context.Background())
	schedCtx, schedDone := context.WithCancel(ctx)
	sched := InitScheduler(schedCtx)
	s.sched = sched
	s.modelCaches.Start(ctx)

	slog.Info(fmt.Sprintf("Listening on %s (version %s)", ln.Addr(), version.Version))
	h2s := &http2.Server{}
	srvr := &http.Server{
		// Use http.DefaultServeMux so we get net/http/pprof for
		// free.
		//
		// TODO(bmizerany): Decide if we want to make this
		// configurable so it is not exposed by default, or allow
		// users to bind it to a different port. This was a quick
		// and easy way to get pprof, but it may not be the best
		// way.
		Handler: h2c.NewHandler(http.DefaultServeMux, h2s),
	}

	// listen for a ctrl+c and stop any loaded llm
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		srvr.Close()
		schedDone()
		sched.unloadAllRunners()
		done()
	}()

	s.sched.Run(schedCtx)

	// register the experimental webp decoder
	// so webp images can be used in multimodal inputs
	image.RegisterFormat("webp", "RIFF????WEBP", webp.Decode, webp.DecodeConfig)

	// At startup we retrieve GPU information so we can get log messages before loading a model
	// This will log warnings to the log in case we have problems with detected GPUs
	gpus := discover.GPUDevices(ctx, nil)
	discover.LogDetails(gpus)

	var totalVRAM uint64
	for _, gpu := range gpus {
		totalVRAM += gpu.TotalMemory - envconfig.GpuOverhead()
	}

	// Set default context based on VRAM tier
	// Use slightly lower thresholds (47/23 GiB vs. 48/24 GiB) to account for small differences in the exact value
	switch {
	case totalVRAM >= 47*format.GibiByte:
		s.defaultNumCtx = 262144
	case totalVRAM >= 23*format.GibiByte:
		s.defaultNumCtx = 32768
	default:
		s.defaultNumCtx = 4096
	}
	slog.Info("vram-based default context", "total_vram", format.HumanBytes2(totalVRAM), "default_num_ctx", s.defaultNumCtx)

	err = srvr.Serve(ln)
	// If server is closed from the signal handler, wait for the ctx to be done
	// otherwise error out quickly
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	<-ctx.Done()
	return nil
}

func waitForStream(c *gin.Context, ch chan any) {
	c.Header("Content-Type", "application/json")
	var latest api.ProgressResponse
	for resp := range ch {
		switch r := resp.(type) {
		case api.ProgressResponse:
			latest = r
		case gin.H:
			status, ok := r["status"].(int)
			if !ok {
				status = http.StatusInternalServerError
			}
			errorMsg, ok := r["error"].(string)
			if !ok {
				errorMsg = "unknown error"
			}
			c.JSON(status, gin.H{"error": errorMsg})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown message type"})
			return
		}
	}

	c.JSON(http.StatusOK, latest)
}

var jsonBufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func streamResponse(c *gin.Context, ch chan any) {
	c.Header("Content-Type", "application/x-ndjson")
	c.Stream(func(w io.Writer) bool {
		val, ok := <-ch
		if !ok {
			return false
		}

		// errors are provided as a gin.H with an "error" field and
		// an optional "status" field.  For errors that are streamed
		// before any content, we need to set the status code and
		// content type for the error.
		if h, ok := val.(gin.H); ok {
			if e, ok := h["error"].(string); ok {
				status, ok := h["status"].(int)
				if !ok {
					status = http.StatusInternalServerError
				}

				if !c.Writer.Written() {
					c.Header("Content-Type", "application/json")
					c.JSON(status, gin.H{"error": e})
				} else {
					buf := jsonBufferPool.Get().(*bytes.Buffer)
					buf.Reset()
					if err := json.NewEncoder(buf).Encode(gin.H{"error": e}); err != nil {
						slog.Error("streamResponse failed to encode json error", "error", err)
					} else {
						if _, err := w.Write(buf.Bytes()); err != nil {
							slog.Error("streamResponse failed to write json error", "error", err)
						}
					}
					jsonBufferPool.Put(buf)
				}

				return false
			}
		}

		buf := jsonBufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		err := json.NewEncoder(buf).Encode(val)
		if err != nil {
			jsonBufferPool.Put(buf)
			slog.Info(fmt.Sprintf("streamResponse: json encode failed with %s", err))
			return false
		}

		_, err = w.Write(buf.Bytes())
		jsonBufferPool.Put(buf)
		if err != nil {
			slog.Info(fmt.Sprintf("streamResponse: w.Write failed with %s", err))
			return false
		}

		return true
	})
}

func (s *Server) StatusHandler(c *gin.Context) {
	disabled, source := internalcloud.Status()
	c.JSON(http.StatusOK, api.StatusResponse{
		Cloud: api.CloudStatus{
			Disabled: disabled,
			Source:   source,
		},
	})
}

func (s *Server) WebSearchExperimentalHandler(c *gin.Context) {
	s.webExperimentalProxyHandler(c, "/api/web_search", cloudErrWebSearchUnavailable)
}

func (s *Server) WebFetchExperimentalHandler(c *gin.Context) {
	s.webExperimentalProxyHandler(c, "/api/web_fetch", cloudErrWebFetchUnavailable)
}

func (s *Server) webExperimentalProxyHandler(c *gin.Context, proxyPath, disabledOperation string) {
	body, err := readRequestBody(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(bytes.TrimSpace(body)) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	}

	proxyCloudRequestWithPath(c, body, proxyPath, disabledOperation)
}

func (s *Server) WhoamiHandler(c *gin.Context) {
	// todo allow other hosts
	u, err := url.Parse("https://lychee.com")
	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "URL parse error"})
		return
	}

	client := api.NewClient(u, http.DefaultClient)
	user, err := client.Whoami(c)
	if err != nil {
		var authErr api.AuthorizationError
		if errors.As(err, &authErr) && authErr.StatusCode == http.StatusUnauthorized {
			// Preserve an actionable sign-in response for launch; other failures
			// below mean account or plan verification is temporarily unavailable.
			sURL := authErr.SigninURL
			if sURL == "" {
				var sErr error
				sURL, sErr = signinURL()
				if sErr != nil {
					slog.Error(sErr.Error())
					c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting authorization details"})
					return
				}
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "signin_url": sURL})
			return
		}

		slog.Error(err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "account unavailable"})
		return
	}

	if user == nil || user.Name == "" {
		sURL, sErr := signinURL()
		if sErr != nil {
			slog.Error(sErr.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting authorization details"})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "signin_url": sURL})
		return
	}

	if strings.TrimSpace(user.Plan) == "" {
		slog.Warn("account plan was not set; defaulting to free")
		user.Plan = "free"
	}
	c.JSON(http.StatusOK, user)
}

func (s *Server) SignoutHandler(c *gin.Context) {
	pubKey, err := auth.GetPublicKey()
	if err != nil {
		slog.Error("couldn't get public key", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "there was an error signing out"})
		return
	}

	encKey := base64.RawURLEncoding.EncodeToString([]byte(pubKey))

	// todo allow other hosts
	u, err := url.Parse("https://lychee.com")
	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "URL parse error"})
		return
	}

	client := api.NewClient(u, http.DefaultClient)
	err = client.Disconnect(c, encKey)
	if err != nil {
		var authError api.AuthorizationError
		if errors.As(err, &authError) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not currently signed in"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "there was an error signing out"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (s *Server) PsHandler(c *gin.Context) {
	models := []api.ProcessModelResponse{}

	for _, v := range s.sched.loaded {
		m := v.model
		displayName := model.ParseName(m.ShortName).DisplayShortest()
		modelDetails := api.ModelDetails{
			Format:            m.Config.ModelFormat,
			Family:            m.Config.ModelFamily,
			Families:          m.Config.ModelFamilies,
			ParameterSize:     m.Config.ModelType,
			QuantizationLevel: m.Config.FileType,
		}

		mr := api.ProcessModelResponse{
			Model:     displayName,
			Name:      displayName,
			Size:      int64(v.totalSize),
			SizeVRAM:  int64(v.vramSize),
			Digest:    m.Digest,
			Details:   modelDetails,
			ExpiresAt: v.expiresAt,
		}
		if v.llama != nil {
			mr.ContextLength = v.llama.ContextLength()
			total, vram := v.llama.MemorySize()
			mr.Size = int64(total)
			mr.SizeVRAM = int64(vram)
		}
		// The scheduler waits to set expiresAt, so if a model is loading it's
		// possible that it will be set to the unix epoch. For those cases, just
		// calculate the time w/ the sessionDuration instead.
		var epoch time.Time
		if v.expiresAt == epoch {
			mr.ExpiresAt = time.Now().Add(v.sessionDuration)
		}

		models = append(models, mr)
	}

	slices.SortStableFunc(models, func(i, j api.ProcessModelResponse) int {
		// longest duration remaining listed first
		return cmp.Compare(j.ExpiresAt.Unix(), i.ExpiresAt.Unix())
	})

	c.JSON(http.StatusOK, api.ProcessResponse{Models: models})
}

func toolCallId() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return "call_" + strings.ToLower(string(b))
}

func preservedTokensForCompletion(builtinParser parsers.Parser) []string {
	if builtinParser != nil {
		return builtinParser.PreservedTokens()
	}
	return nil
}

func toolCallTagForCompletion(toolParser *tools.Parser) string {
	if toolParser == nil {
		return ""
	}
	return toolParser.Tag()
}

func leadingBOSForModel(m *Model) string {
	if m == nil || m.Config.Renderer == "" {
		return ""
	}

	return renderers.LeadingBOSForRenderer(resolveRendererName(m))
}

func optionsForPrompt(opts *api.Options, runner llm.LlamaServer) *api.Options {
	if opts == nil || runner == nil {
		return opts
	}

	if ctxLen := runner.ContextLength(); ctxLen > 0 && opts.NumCtx > ctxLen {
		copied := *opts
		copied.NumCtx = ctxLen
		return &copied
	}

	return opts
}

type chatExecutionMode int

const (
	chatExecutionModeNative chatExecutionMode = iota
	chatExecutionModeRendered
)

func chatModeForModel(m *Model) chatExecutionMode {
	if m.IsMLX() || usesLycheeRenderedChat(m) {
		return chatExecutionModeRendered
	}

	return chatExecutionModeNative
}

func llamaServerConfigForModel(m *Model) llm.LlamaServerConfig {
	return llm.LlamaServerConfig{
		DisableJinja:   usesLycheeRenderedChat(m),
		DraftModelPath: m.DraftPath,
	}
}

func usesLycheeRenderedChat(m *Model) bool {
	return m != nil && (m.Config.Renderer != "" || m.Config.Parser != "" || shouldUseHarmony(m) || shouldUseGoTemplate(m))
}

func shouldUseGoTemplate(m *Model) bool {
	if !m.HasGoTemplate {
		return false
	}
	if goTemplateEnvSet() {
		return envconfig.GoTemplate(true)
	}

	return !m.PreferChatTemplate && envconfig.GoTemplate(true)
}



func handleScheduleError(c *gin.Context, name string, err error) {
	switch {
	case errors.Is(err, errCapabilities), errors.Is(err, errRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, context.Canceled):
		c.JSON(499, gin.H{"error": "request canceled"})
	case errors.Is(err, ErrMaxQueue):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	case errors.Is(err, os.ErrNotExist):
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model %q not found, try pulling it first", name)})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func filterThinkTags(msgs []api.Message, m *Model) []api.Message {
	if m.Config.ModelFamily == "qwen3" || model.ParseName(m.Name).Model == "deepseek-r1" {
		finalUserIndex := -1
		for i, msg := range msgs {
			if msg.Role == "user" {
				finalUserIndex = i
			}
		}

		for i, msg := range msgs {
			if msg.Role == "assistant" && i < finalUserIndex {
				// TODO(drifkin): this is from before we added proper thinking support.
				// However, even if thinking is not enabled (and therefore we shouldn't
				// change the user output), we should probably perform this filtering
				// for all thinking models (not just qwen3 & deepseek-r1) since it tends
				// to save tokens and improve quality.
				thinkingState := &thinking.Parser{
					OpeningTag: "<think>",
					ClosingTag: "</think>",
				}
				_, content := thinkingState.AddContent(msg.Content)
				msgs[i].Content = content
			}
		}
	}
	return msgs
}

// handleImageGenerate handles image generation requests within GenerateHandler.
// This is called when the model has the Image capability.
func (s *Server) handleImageGenerate(c *gin.Context, req api.GenerateRequest, modelName string, checkpointStart time.Time) {
	// Validate image dimensions
	const maxDimension int32 = 4096
	if req.Width > maxDimension || req.Height > maxDimension {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("width and height must be <= %d", maxDimension)})
		return
	}

	// Schedule the runner for image generation
	runner, m, _, err := s.scheduleRunner(c.Request.Context(), modelName, []model.Capability{model.CapabilityImage}, nil, req.KeepAlive, nil)
	if err != nil {
		handleScheduleError(c, req.Model, err)
		return
	}

	checkpointLoaded := time.Now()

	// Handle load-only request (empty prompt)
	if req.Prompt == "" {
		c.JSON(http.StatusOK, api.GenerateResponse{
			Model:      req.Model,
			CreatedAt:  time.Now().UTC(),
			Done:       true,
			DoneReason: "load",
		})
		return
	}

	// Check streaming preference
	isStreaming := req.Stream == nil || *req.Stream

	contentType := "application/x-ndjson"
	if !isStreaming {
		contentType = "application/json; charset=utf-8"
	}
	c.Header("Content-Type", contentType)

	// Get seed from options if provided
	var seed int64
	if s, ok := req.Options["seed"]; ok {
		switch v := s.(type) {
		case int:
			seed = int64(v)
		case int64:
			seed = v
		case float64:
			seed = int64(v)
		}
	}

	var media []llm.MediaData
	for i, imgData := range req.Images {
		media = append(media, llm.NewMediaData(i, imgData))
	}

	var streamStarted bool
	var finalResponse api.GenerateResponse

	if err := runner.Completion(c.Request.Context(), llm.CompletionRequest{
		Prompt: req.Prompt,
		Width:  req.Width,
		Height: req.Height,
		Steps:  req.Steps,
		Seed:   seed,
		Media:  media,
	}, func(cr llm.CompletionResponse) {
		streamStarted = true
		res := api.GenerateResponse{
			Model:     req.Model,
			CreatedAt: time.Now().UTC(),
			Done:      cr.Done,
		}

		if cr.TotalSteps > 0 {
			res.Completed = int64(cr.Step)
			res.Total = int64(cr.TotalSteps)
		}

		if cr.Image != "" {
			res.Image = cr.Image
		}

		if cr.Done {
			res.DoneReason = cr.DoneReason.String()
			res.Metrics.TotalDuration = time.Since(checkpointStart)
			res.Metrics.LoadDuration = checkpointLoaded.Sub(checkpointStart)
		}

		if !isStreaming {
			finalResponse = res
			return
		}

		data, _ := json.Marshal(res)
		c.Writer.Write(append(data, '\n'))
		c.Writer.Flush()
	}); err != nil {
		s.sched.expireRunnersForRuntimeOOM(m, err)
		// Only send JSON error if streaming hasn't started yet
		// (once streaming starts, headers are committed and we can't change status code)
		if !isStreaming || !streamStarted {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			data, _ := json.Marshal(gin.H{"error": err.Error()})
			c.Writer.Write(append(data, '\n'))
			c.Writer.Flush()
		}
		return
	}

	if !isStreaming {
		c.JSON(http.StatusOK, finalResponse)
	}
}
