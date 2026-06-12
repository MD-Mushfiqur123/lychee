package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lychee/lychee/api"
	"github.com/lychee/lychee/envconfig"
	internalcloud "github.com/lychee/lychee/internal/cloud"
	"github.com/lychee/lychee/llm"
	"github.com/lychee/lychee/logutil"
	"github.com/lychee/lychee/model/parsers"
	"github.com/lychee/lychee/thinking"
	"github.com/lychee/lychee/tools"
	"github.com/lychee/lychee/types/errtypes"
	"github.com/lychee/lychee/types/model"
)

func writeChatResponseInternal(c *gin.Context, req api.ChatRequest, ch chan any) {
	if req.Stream != nil && !*req.Stream {
		var resp api.ChatResponse
		var toolCalls []api.ToolCall
		var allLogprobs []api.Logprob
		var sbThinking strings.Builder
		var sbContent strings.Builder
		for rr := range ch {
			switch t := rr.(type) {
			case api.ChatResponse:
				sbThinking.WriteString(t.Message.Thinking)
				sbContent.WriteString(t.Message.Content)
				resp = t
				if len(req.Tools) > 0 {
					toolCalls = append(toolCalls, t.Message.ToolCalls...)
				}
				if len(t.Logprobs) > 0 {
					allLogprobs = append(allLogprobs, t.Logprobs...)
				}
			case gin.H:
				msg, ok := t["error"].(string)
				if !ok {
					msg = "unexpected error format in response"
				}

				status, ok := t["status"].(int)
				if !ok {
					status = http.StatusInternalServerError
				}

				c.JSON(status, gin.H{"error": msg})
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected response"})
				return
			}
		}

		resp.Message.Content = sbContent.String()
		resp.Message.Thinking = sbThinking.String()
		resp.Logprobs = allLogprobs

		if len(toolCalls) > 0 {
			resp.Message.ToolCalls = toolCalls
		}

		if req.ValidateOutput {
			if err := ValidateJSONSchema(resp.Message.Content, req.Format); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("schema validation failed: %v", err)})
				return
			}
		}

		c.JSON(http.StatusOK, resp)
		return
	}

	streamResponse(c, ch)
}

func (s *Server) writeChatResponse(c *gin.Context, req api.ChatRequest, ch chan any) {
	if req.ConversationID == "" || s.memoryStore == nil {
		writeChatResponseInternal(c, req, ch)
		return
	}

	chOut := make(chan any)
	go func() {
		defer close(chOut)
		var assistantResponse strings.Builder
		var assistantThinking strings.Builder
		for msg := range ch {
			chOut <- msg
			switch t := msg.(type) {
			case api.ChatResponse:
				assistantResponse.WriteString(t.Message.Content)
				assistantThinking.WriteString(t.Message.Thinking)
			}
		}
		assistantMsg := api.Message{
			Role:     "assistant",
			Content:  assistantResponse.String(),
			Thinking: assistantThinking.String(),
		}
		_ = s.memoryStore.AppendMessage(req.ConversationID, assistantMsg)
	}()

	writeChatResponseInternal(c, req, chOut)
}

func validateChatRequest(req *api.ChatRequest) error {
	if req.TopLogprobs < 0 || req.TopLogprobs > 20 {
		return errors.New("top_logprobs must be between 0 and 20")
	}
	return nil
}

func (s *Server) ChatHandler(c *gin.Context) {
	checkpointStart := time.Now()

	var req api.ChatRequest
	if err := c.ShouldBindJSON(&req); errors.Is(err, io.EOF) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateChatRequest(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ConversationID != "" && s.memoryStore != nil {
		if len(req.Messages) == 0 {
			conv, err := s.memoryStore.Load(req.ConversationID)
			if err == nil && conv != nil {
				req.Messages = conv.Messages
			}
		} else {
			conv, err := s.memoryStore.Load(req.ConversationID)
			if err != nil || conv == nil {
				conv = &Conversation{
					ID:        req.ConversationID,
					Model:     req.Model,
					Messages:  req.Messages,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
			} else {
				conv.Messages = req.Messages
				conv.UpdatedAt = time.Now()
			}
			_ = s.memoryStore.Save(conv)
		}
	}

	if s.modelAliases != nil {
		req.Model = s.modelAliases.Resolve(req.Model)
	}

	var releaseRoute func() = func() {}
	if s.modelRouter != nil {
		originalModel := req.Model
		endpoint, release, err := s.modelRouter.Resolve(req.Model)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if endpoint != nil {
			releaseRoute = release
			req.Model = endpoint.Model
			if endpoint.Host != "" {
				defer releaseRoute()
				remoteURL, err := url.Parse(endpoint.Host)
				if err != nil {
					s.modelRouter.RecordFailure(originalModel, endpoint.Host)
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				contentType := "application/x-ndjson"
				if req.Stream != nil && !*req.Stream {
					contentType = "application/json; charset=utf-8"
				}
				c.Header("Content-Type", contentType)

				var assistantResponse strings.Builder
				var assistantThinking strings.Builder
				fn := func(resp api.ChatResponse) error {
					assistantResponse.WriteString(resp.Message.Content)
					assistantThinking.WriteString(resp.Message.Thinking)
					data, err := json.Marshal(resp)
					if err != nil {
						return err
					}
					if _, err = c.Writer.Write(append(data, '\n')); err != nil {
						return err
					}
					c.Writer.Flush()
					return nil
				}

				client := api.NewClient(remoteURL, http.DefaultClient)
				err = client.Chat(c, &req, fn)
				if err != nil {
					s.modelRouter.RecordFailure(originalModel, endpoint.Host)
					var apiError api.StatusError
					if errors.As(err, &apiError) {
						c.JSON(apiError.StatusCode, apiError)
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				s.modelRouter.RecordSuccess(originalModel, endpoint.Host)

				if req.ConversationID != "" && s.memoryStore != nil && (assistantResponse.Len() > 0 || assistantThinking.Len() > 0) {
					assistantMsg := api.Message{
						Role:     "assistant",
						Content:  assistantResponse.String(),
						Thinking: assistantThinking.String(),
					}
					_ = s.memoryStore.AppendMessage(req.ConversationID, assistantMsg)
				}
				return
			}
		}
	}
	defer releaseRoute()

	modelRef, err := parseAndValidateModelRef(req.Model)
	if err != nil {
		writeModelRefParseError(c, err, http.StatusBadRequest, "model is required")
		return
	}

	if modelRef.Source == modelSourceCloud {
		req.Model = modelRef.Base
		if c.GetBool(legacyCloudAnthropicKey) {
			proxyCloudJSONRequestWithPath(c, req, "/api/chat", cloudErrRemoteInferenceUnavailable)
			return
		}
		proxyCloudJSONRequest(c, req, cloudErrRemoteInferenceUnavailable)
		return
	}

	name := modelRef.Name

	name, err = getExistingName(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	m, err := GetModel(name.String())
	if err != nil {
		switch {
		case os.IsNotExist(err):
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", req.Model)})
		case err.Error() == errtypes.InvalidModelNameErrMsg:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if modelRef.Source == modelSourceLocal && m.Config.RemoteHost != "" && m.Config.RemoteModel != "" {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", req.Model)})
		return
	}

	// expire the runner
	if len(req.Messages) == 0 && req.KeepAlive != nil && req.KeepAlive.Duration == 0 {
		s.sched.expireRunner(m)

		c.JSON(http.StatusOK, api.ChatResponse{
			Model:      req.Model,
			CreatedAt:  time.Now().UTC(),
			Message:    api.Message{Role: "assistant"},
			Done:       true,
			DoneReason: "unload",
		})
		return
	}

	if m.Config.RemoteHost != "" && m.Config.RemoteModel != "" {
		if disabled, _ := internalcloud.Status(); disabled {
			c.JSON(http.StatusForbidden, gin.H{"error": internalcloud.DisabledError(cloudErrRemoteInferenceUnavailable)})
			return
		}

		origModel := req.Model

		remoteURL, err := url.Parse(m.Config.RemoteHost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !slices.Contains(envconfig.Remotes(), remoteURL.Hostname()) {
			slog.Info("remote model", "remotes", envconfig.Remotes(), "remoteURL", m.Config.RemoteHost, "hostname", remoteURL.Hostname())
			c.JSON(http.StatusBadRequest, gin.H{"error": "this server cannot run this remote model"})
			return
		}

		req.Model = m.Config.RemoteModel
		if req.Options == nil {
			req.Options = map[string]any{}
		}

		var msgs []api.Message
		if len(req.Messages) > 0 {
			msgs = append(m.Messages, req.Messages...)
			if req.Messages[0].Role != "system" && m.System != "" {
				msgs = append([]api.Message{{Role: "system", Content: m.System}}, msgs...)
			}
		}

		msgs = filterThinkTags(msgs, m)
		req.Messages = msgs

		for k, v := range m.Options {
			if _, ok := req.Options[k]; !ok {
				req.Options[k] = v
			}
		}

		contentType := "application/x-ndjson"
		if req.Stream != nil && !*req.Stream {
			contentType = "application/json; charset=utf-8"
		}
		c.Header("Content-Type", contentType)

		fn := func(resp api.ChatResponse) error {
			resp.Model = origModel
			resp.RemoteModel = m.Config.RemoteModel
			resp.RemoteHost = m.Config.RemoteHost

			data, err := json.Marshal(resp)
			if err != nil {
				return err
			}

			if _, err = c.Writer.Write(append(data, '\n')); err != nil {
				return err
			}
			c.Writer.Flush()
			return nil
		}

		client := api.NewClient(remoteURL, http.DefaultClient)
		err = client.Chat(c, &req, fn)
		if err != nil {
			var authError api.AuthorizationError
			if errors.As(err, &authError) {
				sURL, sErr := signinURL()
				if sErr != nil {
					slog.Error(sErr.Error())
					c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting authorization details"})
					return
				}

				c.JSON(authError.StatusCode, gin.H{"error": "unauthorized", "signin_url": sURL})
				return
			}
			var apiError api.StatusError
			if errors.As(err, &apiError) {
				c.JSON(apiError.StatusCode, apiError)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		return
	}

	caps := []model.Capability{model.CapabilityCompletion}
	if len(req.Tools) > 0 {
		caps = append(caps, model.CapabilityTools)
	}

	modelCaps := m.Capabilities()
	if slices.Contains(modelCaps, model.CapabilityThinking) {
		caps = append(caps, model.CapabilityThinking)
		if req.Think == nil {
			req.Think = &api.ThinkValue{Value: true}
		}
	} else {
		if req.Think != nil && req.Think.Bool() {
			// Set think to nil when being used with Anthropic API to connect to tools like claude code
			if _, ok := c.Get("relax_thinking"); ok {
				slog.Warn("model does not support thinking, relaxing thinking to nil", "model", req.Model)
				req.Think = nil
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%q does not support thinking", req.Model)})
				return
			}
		}
	}

	if HasCacheControlChat(&req) {
		trueVal := true
		req.Shift = &trueVal
	}

	r, m, opts, err := s.scheduleRunner(c.Request.Context(), name.String(), caps, req.Options, req.KeepAlive, req.Shift)
	if errors.Is(err, errCapabilityCompletion) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%q does not support chat", req.Model)})
		return
	} else if err != nil {
		handleScheduleError(c, req.Model, err)
		return
	}

	checkpointLoaded := time.Now()

	if len(req.Messages) == 0 {
		c.JSON(http.StatusOK, api.ChatResponse{
			Model:      req.Model,
			CreatedAt:  time.Now().UTC(),
			Message:    api.Message{Role: "assistant"},
			Done:       true,
			DoneReason: "load",
		})
		return
	}

	msgs := append(m.Messages, req.Messages...)
	if req.Messages[0].Role != "system" && m.System != "" {
		msgs = append([]api.Message{{Role: "system", Content: m.System}}, msgs...)
	}
	msgs = filterThinkTags(msgs, m)

	if shouldUseHarmony(m) {
		// harmony's Reasoning field only understands low/medium/high; map "max" to "high"
		if req.Think != nil {
			if s, ok := req.Think.Value.(string); ok && s == "max" {
				req.Think.Value = "high"
			}
		}
		if m.Config.Parser == "" {
			m.Config.Parser = "harmony"
		}
	}

	if chatModeForModel(m) == chatExecutionModeNative {
		s.handleNativeChat(c, req, m, r, opts, msgs, checkpointStart, checkpointLoaded)
		return
	}

	var builtinParser parsers.Parser
	processedTools := req.Tools

	if m.Config.Parser != "" {
		builtinParser = parsers.ParserForName(m.Config.Parser)
		if builtinParser != nil {
			// Determine last message for chat prefill
			var lastMessage *api.Message
			if len(msgs) > 0 {
				lastMessage = &msgs[len(msgs)-1]
			}
			// Initialize parser and get processed tools
			processedTools = builtinParser.Init(req.Tools, lastMessage, req.Think)
		}
	}

	truncate := req.Truncate == nil || *req.Truncate
	if m.IsMLX() {
		truncate = false
	}
	promptOpts := optionsForPrompt(opts, r)
	prompt, media, err := chatPrompt(c.Request.Context(), m, r.Tokenize, promptOpts, msgs, processedTools, req.Think, truncate)
	if err != nil {
		slog.Error("chat prompt error", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If debug mode is enabled, return the rendered template instead of calling the model
	if req.DebugRenderOnly {
		c.JSON(http.StatusOK, api.ChatResponse{
			Model:     req.Model,
			CreatedAt: time.Now().UTC(),
			DebugInfo: &api.DebugInfo{
				RenderedTemplate: prompt,
				ImageCount:       len(media),
			},
		})
		return
	}

	var thinkingState *thinking.Parser
	openingTag, closingTag := thinking.InferTags(m.Template.Template)
	if req.Think != nil && req.Think.Bool() && openingTag != "" && closingTag != "" {
		thinkingState = &thinking.Parser{
			OpeningTag: openingTag,
			ClosingTag: closingTag,
		}

		if strings.HasSuffix(strings.TrimSpace(prompt), openingTag) {
			thinkingState.AddContent(openingTag)
		}
	}

	var toolParser *tools.Parser
	if len(req.Tools) > 0 && (builtinParser == nil || !builtinParser.HasToolSupport()) {
		toolParser = tools.NewParser(m.Template.Template, req.Tools)
	}

	type structuredOutputsState int
	const (
		structuredOutputsState_None structuredOutputsState = iota
		structuredOutputsState_ReadyToApply
		structuredOutputsState_Applying
	)

	ch := make(chan any)
	go func() {
		defer close(ch)

		structuredOutputsState := structuredOutputsState_None

		for {
			var tb strings.Builder

			currentFormat := req.Format
			// structured outputs via double request is enabled when:
			// 1. the model supports the thinking capability and
			// 2. it uses a built-in parser or our generic thinking parser

			// Note that the current approach does not work for (potential future)
			// non-thinking models that emit anything before actual content. This
			// current approach uses the transition from parsed thinking content to
			// parsed non-thinking content as the signal to turn constraining on

			// TODO(parthsareen): temporary fix for https://github.com/lychee/lychee/issues/15260.
			// To revisit for other models and have a consistent pattern across models through parsers.
			forceImmediate := m.Config.Parser == "gemma4" && req.Think != nil && !req.Think.Bool()
			if req.Format != nil && structuredOutputsState == structuredOutputsState_None && !forceImmediate && ((builtinParser != nil || thinkingState != nil) && slices.Contains(m.Capabilities(), model.CapabilityThinking)) {
				currentFormat = nil
			}

			// sets up new context given parent context per request
			ctx, cancel := context.WithCancel(c.Request.Context())

			err := r.Completion(ctx, llm.CompletionRequest{
				Prompt:          prompt,
				Media:           media,
				Format:          currentFormat,
				Options:         opts,
				Shift:           req.Shift == nil || *req.Shift,
				Truncate:        truncate,
				Logprobs:        req.Logprobs,
				TopLogprobs:     req.TopLogprobs,
				PreservedTokens: preservedTokensForCompletion(builtinParser),
				ToolCallTag:     toolCallTagForCompletion(toolParser),
				LeadingBOS:      leadingBOSForModel(m),
			}, func(r llm.CompletionResponse) {
				res := api.ChatResponse{
					Model:     req.Model,
					CreatedAt: time.Now().UTC(),
					Message:   api.Message{Role: "assistant", Content: r.Content},
					Done:      r.Done,
					Metrics: api.Metrics{
						PromptEvalCount:    r.PromptEvalCount,
						PromptEvalDuration: r.PromptEvalDuration,
						EvalCount:          r.EvalCount,
						EvalDuration:       r.EvalDuration,
					},
					Logprobs: toAPILogprobs(r.Logprobs),
				}

				if r.Done {
					res.DoneReason = r.DoneReason.String()
					res.TotalDuration = time.Since(checkpointStart)
					res.LoadDuration = checkpointLoaded.Sub(checkpointStart)
				}

				if builtinParser != nil {
					slog.Log(context.TODO(), logutil.LevelTrace, "builtin parser input", "parser", m.Config.Parser, "content", r.Content)

					content, thinking, toolCalls, err := builtinParser.Add(r.Content, r.Done)
					if err != nil {
						ch <- gin.H{"error": err.Error()}
						return
					}

					res.Message.Content = content
					res.Message.Thinking = thinking
					for i := range toolCalls {
						toolCalls[i].ID = toolCallId()
					}
					res.Message.ToolCalls = toolCalls

					tb.WriteString(thinking)
					// we are now receiving content from the model - we should start applying structured outputs
					if structuredOutputsState == structuredOutputsState_None && req.Format != nil && tb.String() != "" && res.Message.Content != "" {
						structuredOutputsState = structuredOutputsState_ReadyToApply
						cancel()
						return
					}

					if res.Message.Content != "" || res.Message.Thinking != "" || len(res.Message.ToolCalls) > 0 || r.Done || len(res.Logprobs) > 0 {
						slog.Log(context.TODO(), logutil.LevelTrace, "builtin parser output", "parser", m.Config.Parser, "content", content, "thinking", thinking, "toolCalls", toolCalls, "done", r.Done)
						ch <- res
					} else {
						slog.Log(context.TODO(), logutil.LevelTrace, "builtin parser empty output", "parser", m.Config.Parser)
					}
					return
				}

				if thinkingState != nil {
					thinkingContent, remainingContent := thinkingState.AddContent(res.Message.Content)
					if thinkingContent == "" && remainingContent == "" && !r.Done {
						// need to accumulate more to decide what to send
						return
					}
					res.Message.Thinking = thinkingContent
					tb.WriteString(thinkingContent)
					// emit the collected thinking text before restarting with structured outputs and clear unstructured content
					// to avoid leaking mixed tokens like "</think>Hello"
					if structuredOutputsState == structuredOutputsState_None && req.Format != nil && tb.String() != "" && remainingContent != "" {
						structuredOutputsState = structuredOutputsState_ReadyToApply
						res.Message.Content = ""
						ch <- res
						cancel()
						return
					}
					res.Message.Content = remainingContent
				}

				if len(req.Tools) > 0 {
					toolCalls, content := toolParser.Add(res.Message.Content)
					if len(content) > 0 {
						res.Message.Content = content
					} else if len(toolCalls) > 0 {
						for i := range toolCalls {
							toolCalls[i].ID = toolCallId()
						}
						res.Message.ToolCalls = toolCalls
						res.Message.Content = ""
					} else if res.Message.Thinking != "" {
						// don't return, fall through to send
					} else {
						//  Send logprobs while content is being buffered by the parser for tool calls
						if len(res.Logprobs) > 0 && !r.Done {
							logprobRes := res
							logprobRes.Message.Content = ""
							logprobRes.Message.ToolCalls = nil
							ch <- logprobRes
						}

						if r.Done {
							res.Message.Content = toolParser.Content()
							ch <- res
						}
						return
					}
				}

				ch <- res
			})
			if err != nil {
				if structuredOutputsState == structuredOutputsState_ReadyToApply && strings.Contains(err.Error(), "context canceled") && c.Request.Context().Err() == nil {
					// only ignores error if it's a context cancellation due to setting structured outputs
				} else {
					s.sched.expireRunnersForRuntimeOOM(m, err)
					var serr api.StatusError
					if errors.As(err, &serr) {
						ch <- gin.H{"error": serr.ErrorMessage, "status": serr.StatusCode}
					} else {
						ch <- gin.H{"error": err.Error()}
					}
					return
				}
			}

			// ignored structured outputs cancellation falls through to here, start a new request with the structured outputs and updated prompt. use the
			if structuredOutputsState == structuredOutputsState_ReadyToApply {
				structuredOutputsState = structuredOutputsState_Applying
				msg := api.Message{
					Role:     "assistant",
					Thinking: tb.String(),
				}

				msgs = append(msgs, msg)
				prompt, _, err = chatPrompt(c.Request.Context(), m, r.Tokenize, promptOpts, msgs, processedTools, req.Think, truncate)
				if err != nil {
					slog.Error("chat prompt error applying structured outputs", "error", err)
					ch <- gin.H{"error": err.Error()}
					return
				}
				// force constraining by terminating thinking header, the parser is already at this state
				// when the last message is thinking, the rendered for gpt-oss cannot disambiguate between having the
				// model continue thinking or ending thinking and outputting the final message.
				// TODO(parthsareen): consider adding prefill disambiguation logic to the renderer for structured outputs.
				if shouldUseHarmony(m) || (builtinParser != nil && m.Config.Parser == "harmony") {
					prompt += "<|end|><|start|>assistant<|channel|>final<|message|>"
				}
				continue
			}

			break
		}
	}()

	s.writeChatResponse(c, req, ch)
}

func (s *Server) handleNativeChat(c *gin.Context, req api.ChatRequest, m *Model, r llm.LlamaServer, opts *api.Options, msgs []api.Message, checkpointStart, checkpointLoaded time.Time) {
	nativeReq := llm.ChatRequest{
		Messages:    msgs,
		Tools:       req.Tools,
		Format:      req.Format,
		Options:     opts,
		Think:       req.Think,
		Shift:       req.Shift == nil || *req.Shift,
		Logprobs:    req.Logprobs,
		TopLogprobs: req.TopLogprobs,
	}
	truncate := req.Truncate == nil || *req.Truncate
	var err error
	nativeReq.Messages, err = truncateNativeChatMessages(c.Request.Context(), m, r, optionsForPrompt(opts, r), nativeReq, truncate)
	if err != nil {
		slog.Error("chat template prompt error", "error", err)
		var serr api.StatusError
		if errors.As(err, &serr) {
			c.JSON(serr.StatusCode, gin.H{"error": serr.ErrorMessage})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if req.DebugRenderOnly {
		prompt, err := r.ApplyChatTemplate(c.Request.Context(), nativeReq)
		if err != nil {
			var serr api.StatusError
			if errors.As(err, &serr) {
				c.JSON(serr.StatusCode, gin.H{"error": serr.ErrorMessage})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, api.ChatResponse{
			Model:     req.Model,
			CreatedAt: time.Now().UTC(),
			DebugInfo: &api.DebugInfo{
				RenderedTemplate: prompt,
				ImageCount:       countChatImages(msgs),
			},
		})
		return
	}

	ch := make(chan any)
	go func() {
		defer close(ch)

		err := r.Chat(c.Request.Context(), nativeReq, func(r llm.ChatResponse) {
			res := api.ChatResponse{
				Model:     req.Model,
				CreatedAt: time.Now().UTC(),
				Message:   r.Message,
				Done:      r.Done,
				Metrics: api.Metrics{
					PromptEvalCount:    r.PromptEvalCount,
					PromptEvalDuration: r.PromptEvalDuration,
					EvalCount:          r.EvalCount,
					EvalDuration:       r.EvalDuration,
				},
				Logprobs: toAPILogprobs(r.Logprobs),
			}

			if res.Message.Role == "" {
				res.Message.Role = "assistant"
			}

			if r.Done {
				res.DoneReason = r.DoneReason.String()
				res.TotalDuration = time.Since(checkpointStart)
				res.LoadDuration = checkpointLoaded.Sub(checkpointStart)
			}

			ch <- res
		})
		if err != nil {
			s.sched.expireRunnersForRuntimeOOM(m, err)
			var serr api.StatusError
			if errors.As(err, &serr) {
				ch <- gin.H{"error": serr.ErrorMessage, "status": serr.StatusCode}
			} else {
				ch <- gin.H{"error": err.Error()}
			}
		}
	}()

	s.writeChatResponse(c, req, ch)
}

func truncateNativeChatMessages(ctx context.Context, m *Model, r llm.LlamaServer, opts *api.Options, req llm.ChatRequest, truncate bool) ([]api.Message, error) {
	if !truncate || opts == nil || opts.NumCtx <= 0 || len(req.Messages) <= 1 {
		return req.Messages, nil
	}

	lastMsgIdx := len(req.Messages) - 1
	currMsgIdx := 0
	var system []api.Message

	for i := 0; i <= lastMsgIdx; i++ {
		system = system[:0]
		for j := range i {
			if req.Messages[j].Role == "system" {
				system = append(system, req.Messages[j])
			}
		}

		renderReq := req
		renderReq.Messages = append(slices.Clone(system), req.Messages[i:]...)
		prompt, err := r.ApplyChatTemplate(ctx, renderReq)
		if err != nil {
			return nil, err
		}

		tokens, err := r.Tokenize(ctx, prompt)
		if err != nil {
			return nil, err
		}

		ctxLen := len(tokens)
		if m != nil && m.ProjectorPaths != nil {
			for _, msg := range renderReq.Messages {
				ctxLen += 768 * len(msg.Images)
			}
		}

		if ctxLen <= opts.NumCtx {
			currMsgIdx = i
			break
		}
		if i == lastMsgIdx {
			currMsgIdx = lastMsgIdx
			break
		}
	}

	if currMsgIdx > 0 {
		slog.Debug("truncating native chat messages which exceed context length", "truncated", currMsgIdx)
	}

	system = system[:0]
	for j := range currMsgIdx {
		if req.Messages[j].Role == "system" {
			system = append(system, req.Messages[j])
		}
	}
	return append(slices.Clone(system), req.Messages[currMsgIdx:]...), nil
}

func countChatImages(msgs []api.Message) int {
	var count int
	for _, msg := range msgs {
		count += len(msg.Images)
	}
	return count
}
