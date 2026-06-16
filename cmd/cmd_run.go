package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/lychee/lychee/api"
	"github.com/lychee/lychee/internal/modelref"
	"github.com/lychee/lychee/logutil"
	"github.com/lychee/lychee/progress"
	"github.com/lychee/lychee/readline"
	"github.com/lychee/lychee/types/model"
	"github.com/lychee/lychee/cmd/config"
	xcmd "github.com/lychee/lychee/x/cmd"
	"github.com/lychee/lychee/x/imagegen"
)

type generateContextKey string

type runOptions struct {
	Model          string
	ParentModel    string
	LoadedMessages []api.Message
	Prompt         string
	Messages       []api.Message
	WordWrap       bool
	Format         string
	System         string
	Images         []api.ImageData
	Options        map[string]any
	MultiModal     bool
	KeepAlive      *api.Duration
	Think          *api.ThinkValue
	HideThinking   bool
	ShowConnect    bool
}

func (r runOptions) Copy() runOptions {
	var loadedMessages []api.Message
	if r.LoadedMessages != nil {
		loadedMessages = make([]api.Message, len(r.LoadedMessages))
		copy(loadedMessages, r.LoadedMessages)
	}

	var messages []api.Message
	if r.Messages != nil {
		messages = make([]api.Message, len(r.Messages))
		copy(messages, r.Messages)
	}

	var images []api.ImageData
	if r.Images != nil {
		images = make([]api.ImageData, len(r.Images))
		copy(images, r.Images)
	}

	var opts map[string]any
	if r.Options != nil {
		opts = make(map[string]any, len(r.Options))
		for k, v := range r.Options {
			opts[k] = v
		}
	}

	var think *api.ThinkValue
	if r.Think != nil {
		cThink := *r.Think
		think = &cThink
	}

	return runOptions{
		Model:          r.Model,
		ParentModel:    r.ParentModel,
		LoadedMessages: loadedMessages,
		Prompt:         r.Prompt,
		Messages:       messages,
		WordWrap:       r.WordWrap,
		Format:         r.Format,
		System:         r.System,
		Images:         images,
		Options:        opts,
		MultiModal:     r.MultiModal,
		KeepAlive:      r.KeepAlive,
		Think:          think,
		HideThinking:   r.HideThinking,
		ShowConnect:    r.ShowConnect,
	}
}

func applyShowResponseToRunOptions(opts *runOptions, info *api.ShowResponse) {
	opts.ParentModel = info.Details.ParentModel
	opts.LoadedMessages = slices.Clone(info.Messages)
}

type displayResponseState struct {
	lineLength int
	wordBuffer string
}

func displayResponse(content string, wordWrap bool, state *displayResponseState) {
	termWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	if termWidth == 0 {
		termWidth = 80
	}
	if wordWrap && termWidth >= 10 {
		for _, ch := range content {
			if state.lineLength+1 > termWidth-5 {
				if runewidth.StringWidth(state.wordBuffer) > termWidth-10 {
					fmt.Printf("%s%c", state.wordBuffer, ch)
					state.wordBuffer = ""
					state.lineLength = 0
					continue
				}

				// backtrack the length of the last word and clear to the end of the line
				a := runewidth.StringWidth(state.wordBuffer)
				if a > 0 {
					fmt.Printf("\x1b[%dD", a)
				}
				fmt.Printf("\x1b[K\n")
				fmt.Printf("%s%c", state.wordBuffer, ch)
				chWidth := runewidth.RuneWidth(ch)

				state.lineLength = runewidth.StringWidth(state.wordBuffer) + chWidth
			} else {
				fmt.Print(string(ch))
				state.lineLength += runewidth.RuneWidth(ch)
				if runewidth.RuneWidth(ch) >= 2 {
					state.wordBuffer = ""
					continue
				}

				switch ch {
				case ' ', '\t':
					state.wordBuffer = ""
				case '\n', '\r':
					state.lineLength = 0
					state.wordBuffer = ""
				default:
					state.wordBuffer += string(ch)
				}
			}
		}
	} else {
		fmt.Printf("%s%s", state.wordBuffer, content)
		if len(state.wordBuffer) > 0 {
			state.wordBuffer = ""
		}
	}
}

func thinkingOutputOpeningText(plainText bool) string {
	text := "Thinking...\n"

	if plainText {
		return text
	}

	return readline.ColorGrey + readline.ColorBold + text + readline.ColorDefault + readline.ColorGrey
}

func thinkingOutputClosingText(plainText bool) string {
	text := "...done thinking.\n\n"

	if plainText {
		return text
	}

	return readline.ColorGrey + readline.ColorBold + text + readline.ColorDefault
}

func chat(cmd *cobra.Command, opts runOptions) (*api.Message, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	p := progress.NewProgress(os.Stderr)
	defer p.StopAndClear()

	spinner := progress.NewSpinner("")
	p.Add("", spinner)

	cancelCtx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	go func() {
		<-sigChan
		cancel()
	}()

	var state *displayResponseState = &displayResponseState{}
	var thinkingContent strings.Builder
	var latest api.ChatResponse
	var fullResponse strings.Builder
	var thinkTagOpened bool = false
	var thinkTagClosed bool = false

	role := "assistant"

	fn := func(response api.ChatResponse) error {
		if response.Message.Content != "" || !opts.HideThinking {
			p.StopAndClear()
		}

		latest = response

		role = response.Message.Role
		if response.Message.Thinking != "" && !opts.HideThinking {
			if !thinkTagOpened {
				fmt.Print(thinkingOutputOpeningText(false))
				thinkTagOpened = true
				thinkTagClosed = false
			}
			thinkingContent.WriteString(response.Message.Thinking)
			displayResponse(response.Message.Thinking, opts.WordWrap, state)
		}

		content := response.Message.Content
		if thinkTagOpened && !thinkTagClosed && (content != "" || len(response.Message.ToolCalls) > 0) {
			if !strings.HasSuffix(thinkingContent.String(), "\n") {
				fmt.Println()
			}
			fmt.Print(thinkingOutputClosingText(false))
			thinkTagOpened = false
			thinkTagClosed = true
			state = &displayResponseState{}
		}

		fullResponse.WriteString(content)

		if response.Message.ToolCalls != nil {
			toolCalls := response.Message.ToolCalls
			if len(toolCalls) > 0 {
				fmt.Print(renderToolCalls(toolCalls, false))
			}
		}

		displayResponse(content, opts.WordWrap, state)

		return nil
	}

	if opts.Format == "json" {
		opts.Format = `"` + opts.Format + `"`
	}

	req := &api.ChatRequest{
		Model:    opts.Model,
		Messages: opts.Messages,
		Format:   json.RawMessage(opts.Format),
		Options:  opts.Options,
		Think:    opts.Think,
	}

	if opts.KeepAlive != nil {
		req.KeepAlive = opts.KeepAlive
	}

	if err := client.Chat(cancelCtx, req, fn); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, nil
		}

		// this error should ideally be wrapped properly by the client
		if strings.Contains(err.Error(), "upstream error") {
			p.StopAndClear()
			fmt.Println("An error occurred while processing your message. Please try again.")
			fmt.Println()
			return nil, nil
		}
		return nil, err
	}

	if len(opts.Messages) > 0 {
		fmt.Println()
		fmt.Println()
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, err
	}

	if verbose {
		latest.Summary()
	}

	return &api.Message{Role: role, Thinking: thinkingContent.String(), Content: fullResponse.String()}, nil
}

func generate(cmd *cobra.Command, opts runOptions) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	p := progress.NewProgress(os.Stderr)
	defer p.StopAndClear()

	spinner := progress.NewSpinner("")
	p.Add("", spinner)

	var latest api.GenerateResponse

	generateContext, ok := cmd.Context().Value(generateContextKey("context")).([]int)
	if !ok {
		generateContext = []int{}
	}

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	go func() {
		<-sigChan
		cancel()
	}()

	var state *displayResponseState = &displayResponseState{}
	var thinkingContent strings.Builder
	var thinkTagOpened bool = false
	var thinkTagClosed bool = false

	plainText := !term.IsTerminal(int(os.Stdout.Fd()))

	fn := func(response api.GenerateResponse) error {
		latest = response
		content := response.Response

		if response.Response != "" || !opts.HideThinking {
			p.StopAndClear()
		}

		if response.Thinking != "" && !opts.HideThinking {
			if !thinkTagOpened {
				fmt.Print(thinkingOutputOpeningText(plainText))
				thinkTagOpened = true
				thinkTagClosed = false
			}
			thinkingContent.WriteString(response.Thinking)
			displayResponse(response.Thinking, opts.WordWrap, state)
		}

		if thinkTagOpened && !thinkTagClosed && (content != "" || len(response.ToolCalls) > 0) {
			if !strings.HasSuffix(thinkingContent.String(), "\n") {
				fmt.Println()
			}
			fmt.Print(thinkingOutputClosingText(plainText))
			thinkTagOpened = false
			thinkTagClosed = true
			state = &displayResponseState{}
		}

		displayResponse(content, opts.WordWrap, state)

		if response.ToolCalls != nil {
			toolCalls := response.ToolCalls
			if len(toolCalls) > 0 {
				fmt.Print(renderToolCalls(toolCalls, plainText))
			}
		}

		return nil
	}

	if opts.MultiModal {
		opts.Prompt, opts.Images, err = extractFileData(opts.Prompt)
		if err != nil {
			return err
		}
	}

	if opts.Format == "json" {
		opts.Format = `"` + opts.Format + `"`
	}

	request := api.GenerateRequest{
		Model:     opts.Model,
		Prompt:    opts.Prompt,
		Context:   generateContext,
		Images:    opts.Images,
		Format:    json.RawMessage(opts.Format),
		System:    opts.System,
		Options:   opts.Options,
		KeepAlive: opts.KeepAlive,
		Think:     opts.Think,
	}

	if err := client.Generate(ctx, &request, fn); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}

	if opts.Prompt != "" {
		fmt.Println()
		fmt.Println()
	}

	if !latest.Done {
		return nil
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}

	if verbose {
		latest.Summary()
	}

	ctx = context.WithValue(cmd.Context(), generateContextKey("context"), latest.Context)
	cmd.SetContext(ctx)

	return nil
}

func ensureThinkingSupport(ctx context.Context, client *api.Client, name string) {
	if name == "" {
		return
	}
	resp, err := client.Show(ctx, &api.ShowRequest{Model: name})
	if err != nil {
		return
	}
	if slices.Contains(resp.Capabilities, model.CapabilityThinking) {
		return
	}
	fmt.Fprintf(os.Stderr, "warning: model %q does not support thinking output\n", name)
}

func inferThinkingOption(caps *[]model.Capability, runOpts *runOptions, explicitlySetByUser bool) (*api.ThinkValue, error) {
	if explicitlySetByUser {
		return runOpts.Think, nil
	}

	if caps == nil {
		client, err := api.ClientFromEnvironment()
		if err != nil {
			return nil, err
		}
		ret, err := client.Show(context.Background(), &api.ShowRequest{
			Model: runOpts.Model,
		})
		if err != nil {
			return nil, err
		}
		caps = &ret.Capabilities
	}

	thinkingSupported := false
	for _, cap := range *caps {
		if cap == model.CapabilityThinking {
			thinkingSupported = true
		}
	}

	if thinkingSupported {
		return &api.ThinkValue{Value: true}, nil
	}

	return nil, nil
}

func renderToolCalls(toolCalls []api.ToolCall, plainText bool) string {
	out := ""
	formatExplanation := ""
	formatValues := ""
	if !plainText {
		formatExplanation = readline.ColorGrey + readline.ColorBold
		formatValues = readline.ColorDefault
		out += formatExplanation
	}
	for i, toolCall := range toolCalls {
		argsAsJSON, err := json.Marshal(toolCall.Function.Arguments)
		if err != nil {
			return ""
		}
		if i > 0 {
			out += "\n"
		}
		out += fmt.Sprintf("  Model called a non-existent function '%s()' with arguments: %s", formatValues+toolCall.Function.Name+formatExplanation, formatValues+string(argsAsJSON)+formatExplanation)
	}
	if !plainText {
		out += readline.ColorDefault
	}
	return out
}

func loadOrUnloadModel(cmd *cobra.Command, opts *runOptions) error {
	p := progress.NewProgress(os.Stderr)
	defer p.StopAndClear()

	spinner := progress.NewSpinner("")
	p.Add("", spinner)

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	requestedCloud := modelref.HasExplicitCloudSource(opts.Model)

	if info, err := client.Show(cmd.Context(), &api.ShowRequest{Model: opts.Model}); err != nil {
		return err
	} else if info.RemoteHost != "" || requestedCloud {
		isCloud := requestedCloud || strings.HasPrefix(info.RemoteHost, "https://lychee.com")

		if isCloud {
			if _, err := client.Whoami(cmd.Context()); err != nil {
				return err
			}
		}

		if opts.ShowConnect {
			p.StopAndClear()
			remoteModel := info.RemoteModel
			if remoteModel == "" {
				remoteModel = opts.Model
			}
			if isCloud {
				fmt.Fprintf(os.Stderr, "Connecting to '%s' on 'lychee.com' ⚡\n", remoteModel)
			} else {
				fmt.Fprintf(os.Stderr, "Connecting to '%s' on '%s'\n", remoteModel, info.RemoteHost)
			}
		}

		return nil
	}

	req := &api.GenerateRequest{
		Model:     opts.Model,
		KeepAlive: opts.KeepAlive,
		Think:     opts.Think,
	}

	return client.Generate(cmd.Context(), req, func(r api.GenerateResponse) error {
		return nil
	})
}

func generateEmbedding(cmd *cobra.Command, modelName, input string, keepAlive *api.Duration, truncate *bool, dimensions int) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	req := &api.EmbedRequest{
		Model: modelName,
		Input: input,
	}
	if keepAlive != nil {
		req.KeepAlive = keepAlive
	}
	if truncate != nil {
		req.Truncate = truncate
	}
	if dimensions > 0 {
		req.Dimensions = dimensions
	}

	resp, err := client.Embed(cmd.Context(), req)
	if err != nil {
		return err
	}

	if len(resp.Embeddings) == 0 {
		return errors.New("no embeddings returned")
	}

	output, err := json.Marshal(resp.Embeddings[0])
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

func handleCloudAuthorizationError(err error) bool {
	var authErr api.AuthorizationError
	if errors.As(err, &authErr) && authErr.StatusCode == http.StatusUnauthorized {
		fmt.Printf("You need to be signed in to Lychee to run Cloud models.\n\n")
		if authErr.SigninURL != "" {
			fmt.Printf(ConnectInstructions, authErr.SigninURL)
		}
		return true
	}

	return false
}

func ensureCloudStub(ctx context.Context, client *api.Client, modelName string) {
	if !modelref.HasExplicitCloudSource(modelName) {
		return
	}

	normalizedName, _, err := modelref.NormalizePullName(modelName)
	if err != nil {
		slog.Warn("failed to normalize pull name", "model", modelName, "error", err, "normalizedName", normalizedName)
		return
	}

	listResp, err := client.List(ctx)
	if err != nil {
		slog.Warn("failed to list models", "error", err)
		return
	}

	if hasListedModelName(listResp.Models, modelName) || hasListedModelName(listResp.Models, normalizedName) {
		return
	}

	logutil.Trace("pulling cloud stub", "model", modelName, "normalizedName", normalizedName)
	_ = client.Pull(ctx, &api.PullRequest{
		Model: normalizedName,
	}, func(api.ProgressResponse) error {
		return nil
	})
}

func hasListedModelName(models []api.ListModelResponse, name string) bool {
	for _, m := range models {
		if strings.EqualFold(m.Name, name) || strings.EqualFold(m.Model, name) {
			return true
		}
	}
	return false
}

func RunHandler(cmd *cobra.Command, args []string) error {
	interactive := true

	opts := runOptions{
		Model:       args[0],
		WordWrap:    os.Getenv("TERM") == "xterm-256color",
		Options:     map[string]any{},
		ShowConnect: true,
	}

	draftVal, err := cmd.Flags().GetString("draft")
	if err == nil && draftVal != "" {
		opts.Options["draft_model"] = draftVal
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}
	opts.Format = format

	thinkFlag := cmd.Flags().Lookup("think")
	if thinkFlag.Changed {
		thinkStr, err := cmd.Flags().GetString("think")
		if err != nil {
			return err
		}

		switch thinkStr {
		case "", "true":
			opts.Think = &api.ThinkValue{Value: true}
		case "false":
			opts.Think = &api.ThinkValue{Value: false}
		case "high", "medium", "low", "max":
			opts.Think = &api.ThinkValue{Value: thinkStr}
		default:
			return fmt.Errorf("invalid value for --think: %q (must be true, false, high, medium, low, or max)", thinkStr)
		}
	} else {
		opts.Think = nil
	}
	hidethinking, err := cmd.Flags().GetBool("hidethinking")
	if err != nil {
		return err
	}
	opts.HideThinking = hidethinking

	keepAlive, err := cmd.Flags().GetString("keepalive")
	if err != nil {
		return err
	}
	if keepAlive != "" {
		d, err := time.ParseDuration(keepAlive)
		if err != nil {
			return err
		}
		opts.KeepAlive = &api.Duration{Duration: d}
	}

	prompts := args[1:]
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		in, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		stdinContent := string(in)
		if len(stdinContent) > 0 {
			prompts = append([]string{stdinContent}, prompts...)
		}
		opts.ShowConnect = false
		opts.WordWrap = false
		interactive = false
	}
	opts.Prompt = strings.Join(prompts, " ")
	if len(prompts) > 0 {
		interactive = false
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		interactive = false
	}

	nowrap, err := cmd.Flags().GetBool("nowordwrap")
	if err != nil {
		return err
	}
	opts.WordWrap = !nowrap

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	name := args[0]
	requestedCloud := modelref.HasExplicitCloudSource(name)

	info, err := func() (*api.ShowResponse, error) {
		showReq := &api.ShowRequest{Name: name}
		info, err := client.Show(cmd.Context(), showReq)
		var se api.StatusError
		if errors.As(err, &se) && se.StatusCode == http.StatusNotFound {
			if requestedCloud {
				return nil, err
			}
			if err := PullHandler(cmd, []string{name}); err != nil {
				return nil, err
			}
			return client.Show(cmd.Context(), &api.ShowRequest{Name: name})
		}
		return info, err
	}()
	if err != nil {
		if handleCloudAuthorizationError(err) {
			return nil
		}
		return err
	}

	ensureCloudStub(cmd.Context(), client, name)

	opts.Think, err = inferThinkingOption(&info.Capabilities, &opts, thinkFlag.Changed)
	if err != nil {
		return err
	}

	audioCapable := slices.Contains(info.Capabilities, model.CapabilityAudio)
	opts.MultiModal = slices.Contains(info.Capabilities, model.CapabilityVision) || audioCapable

	if len(info.ProjectorInfo) != 0 {
		opts.MultiModal = true
	}
	for k := range info.ModelInfo {
		if strings.Contains(k, ".vision.") {
			opts.MultiModal = true
			break
		}
	}

	applyShowResponseToRunOptions(&opts, info)

	isEmbeddingModel := slices.Contains(info.Capabilities, model.CapabilityEmbedding)

	if isEmbeddingModel {
		if opts.Prompt == "" {
			return errors.New("embedding models require input text. Usage: lychee run " + name + " \"your text here\"")
		}

		var truncate *bool
		if truncateFlag, err := cmd.Flags().GetBool("truncate"); err == nil && cmd.Flags().Changed("truncate") {
			truncate = &truncateFlag
		}

		dimensions, err := cmd.Flags().GetInt("dimensions")
		if err != nil {
			return err
		}

		return generateEmbedding(cmd, name, opts.Prompt, opts.KeepAlive, truncate, dimensions)
	}

	if slices.Contains(info.Capabilities, model.CapabilityImage) {
		if opts.Prompt == "" && !interactive {
			return errors.New("image generation models require a prompt. Usage: lychee run " + name + " \"your prompt here\"")
		}
		return imagegen.RunCLI(cmd, name, opts.Prompt, interactive, opts.KeepAlive)
	}

	isExperimental, _ := cmd.Flags().GetBool("experimental")
	yoloMode, _ := cmd.Flags().GetBool("experimental-yolo")
	enableWebsearch, _ := cmd.Flags().GetBool("experimental-websearch")

	if interactive {
		if err := loadOrUnloadModel(cmd, &opts); err != nil {
			var sErr api.AuthorizationError
			if errors.As(err, &sErr) && sErr.StatusCode == http.StatusUnauthorized {
				fmt.Printf("You need to be signed in to Lychee to run Cloud models.\n\n")

				if sErr.SigninURL != "" {
					fmt.Printf(ConnectInstructions, sErr.SigninURL)
				}
				return nil
			}
			return err
		}

		for _, msg := range info.Messages {
			switch msg.Role {
			case "user":
				fmt.Printf(">>> %s\n", msg.Content)
			case "assistant":
				state := &displayResponseState{}
				displayResponse(msg.Content, opts.WordWrap, state)
				fmt.Println()
				fmt.Println()
			}
		}

		var runErr error
		if isExperimental {
			runErr = xcmd.GenerateInteractive(cmd, opts.Model, opts.WordWrap, opts.Options, opts.Think, opts.HideThinking, opts.KeepAlive, yoloMode, enableWebsearch)
		} else {
			runErr = generateInteractive(cmd, opts)
		}
		if runErr == nil {
			if term.IsTerminal(int(os.Stdout.Fd())) && !config.ShowedStarPrompt() {
				fmt.Fprintf(cmd.OutOrStdout(), "\n⭐ If you find Lychee useful, give us a star: https://github.com/lychee/lychee\n")
				_ = config.SetShowedStarPrompt(true)
			}
		}
		return runErr
	}
	if err := generate(cmd, opts); err != nil {
		if handleCloudAuthorizationError(err) {
			return nil
		}
		return err
	}
	if term.IsTerminal(int(os.Stdout.Fd())) && !config.ShowedStarPrompt() {
		fmt.Fprintf(cmd.OutOrStdout(), "\n⭐ If you find Lychee useful, give us a star: https://github.com/lychee/lychee\n")
		_ = config.SetShowedStarPrompt(true)
	}
	return nil
}

func launchInteractiveModel(cmd *cobra.Command, modelName string) error {
	opts := runOptions{
		Model:       modelName,
		WordWrap:    os.Getenv("TERM") == "xterm-256color",
		Options:     map[string]any{},
		ShowConnect: true,
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	requestedCloud := modelref.HasExplicitCloudSource(modelName)

	info, err := func() (*api.ShowResponse, error) {
		showReq := &api.ShowRequest{Name: modelName}
		info, err := client.Show(cmd.Context(), showReq)
		var se api.StatusError
		if errors.As(err, &se) && se.StatusCode == http.StatusNotFound {
			if requestedCloud {
				return nil, err
			}
			if err := PullHandler(cmd, []string{modelName}); err != nil {
				return nil, err
			}
			return client.Show(cmd.Context(), &api.ShowRequest{Name: modelName})
		}
		return info, err
	}()
	if err != nil {
		if handleCloudAuthorizationError(err) {
			return nil
		}
		return err
	}

	ensureCloudStub(cmd.Context(), client, modelName)

	opts.Think, err = inferThinkingOption(&info.Capabilities, &opts, false)
	if err != nil {
		return err
	}

	audioCapable := slices.Contains(info.Capabilities, model.CapabilityAudio)
	opts.MultiModal = slices.Contains(info.Capabilities, model.CapabilityVision) || audioCapable

	if len(info.ProjectorInfo) != 0 {
		opts.MultiModal = true
	}
	for k := range info.ModelInfo {
		if strings.Contains(k, ".vision.") {
			opts.MultiModal = true
			break
		}
	}

	applyShowResponseToRunOptions(&opts, info)

	if err := loadOrUnloadModel(cmd, &opts); err != nil {
		return fmt.Errorf("error loading model: %w", err)
	}
	if err := generateInteractive(cmd, opts); err != nil {
		return fmt.Errorf("error running model: %w", err)
	}
	return nil
}
