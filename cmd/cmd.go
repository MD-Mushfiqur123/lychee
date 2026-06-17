package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/containerd/console"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/cmd/config"
	"github.com/MD-Mushfiqur123/lychee/cmd/launch"
	"github.com/MD-Mushfiqur123/lychee/cmd/tui"
	"github.com/MD-Mushfiqur123/lychee/discover"
	"github.com/MD-Mushfiqur123/lychee/envconfig"
	"github.com/MD-Mushfiqur123/lychee/runner"
	"github.com/MD-Mushfiqur123/lychee/version"
	xcreate "github.com/MD-Mushfiqur123/lychee/x/create"
	"github.com/MD-Mushfiqur123/lychee/x/imagegen"
)

func init() {
	// Override default selectors to use Bubbletea TUI instead of raw terminal I/O.
	launch.DefaultSingleSelector = func(title string, items []launch.SelectionItem, current string) (string, error) {
		return runTUISingleSelector(title, items, current, nil)
	}

	launch.DefaultSingleSelectorWithUpdates = func(title string, items []launch.SelectionItem, current string, updates <-chan []launch.SelectionItem) (string, error) {
		return runTUISingleSelector(title, items, current, updates)
	}

	launch.DefaultMultiSelector = func(title string, items []launch.SelectionItem, preChecked []string) ([]string, error) {
		return runTUIMultiSelector(title, items, preChecked, nil)
	}

	launch.DefaultMultiSelectorWithUpdates = func(title string, items []launch.SelectionItem, preChecked []string, updates <-chan []launch.SelectionItem) ([]string, error) {
		return runTUIMultiSelector(title, items, preChecked, updates)
	}

	launch.DefaultSignIn = func(modelName, signInURL string) (string, error) {
		userName, err := tui.RunSignIn(modelName, signInURL)
		if errors.Is(err, tui.ErrCancelled) {
			return "", launch.ErrCancelled
		}
		return userName, err
	}

	launch.DefaultUpgrade = func(modelName, requiredPlan string) (string, error) {
		plan, err := tui.RunUpgrade(modelName, requiredPlan)
		if errors.Is(err, tui.ErrCancelled) {
			return "", launch.ErrCancelled
		}
		return plan, err
	}

	launch.DefaultConfirmPrompt = tui.RunConfirmWithOptions
}

func runTUISingleSelector(title string, items []launch.SelectionItem, current string, updates <-chan []launch.SelectionItem) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return "", fmt.Errorf("model selection requires an interactive terminal; use --model to run in headless mode")
	}
	tuiItems := tui.ReorderItems(tui.ConvertItems(items))
	result, err := tui.SelectSingleWithUpdates(title, tuiItems, current, convertSelectionItemUpdates(updates))
	if errors.Is(err, tui.ErrCancelled) {
		return "", launch.ErrCancelled
	}
	return result, err
}

func runTUIMultiSelector(title string, items []launch.SelectionItem, preChecked []string, updates <-chan []launch.SelectionItem) ([]string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return nil, fmt.Errorf("model selection requires an interactive terminal; use --model to run in headless mode")
	}
	tuiItems := tui.ReorderItems(tui.ConvertItems(items))
	result, err := tui.SelectMultipleWithUpdates(title, tuiItems, preChecked, convertSelectionItemUpdates(updates))
	if errors.Is(err, tui.ErrCancelled) {
		return nil, launch.ErrCancelled
	}
	return result, err
}

func convertSelectionItemUpdates(updates <-chan []launch.SelectionItem) <-chan []tui.SelectItem {
	if updates == nil {
		return nil
	}
	out := make(chan []tui.SelectItem, 1)
	go func() {
		defer close(out)
		for items := range updates {
			out <- tui.ReorderItems(tui.ConvertItems(items))
		}
	}()
	return out
}

const ConnectInstructions = "If your browser did not open, navigate to:\n    %s\n\n"

var errModelfileNotFound = errors.New("specified Modelfile wasn't found")

func getModelfileName(cmd *cobra.Command) (string, error) {
	filename, _ := cmd.Flags().GetString("file")

	if filename == "" {
		filename = "Modelfile"
	}

	absName, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}

	_, err = os.Stat(absName)
	if err != nil {
		return "", err
	}

	return absName, nil
}

// isLocalhost returns true if the configured Lychee host is a loopback or unspecified address.
func isLocalhost() bool {
	host := envconfig.Host()
	h, _, _ := net.SplitHostPort(host.Host)
	if h == "localhost" {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && (ip.IsLoopback() || ip.IsUnspecified())
}

func resolveExperimentalLocalModelDir(ref, filename string) string {
	if ref == "" || filepath.IsAbs(ref) || filename == "" {
		return ref
	}

	candidate := filepath.Join(filepath.Dir(filename), ref)
	if xcreate.IsSafetensorsModelDir(candidate) || xcreate.IsTensorModelDir(candidate) {
		return candidate
	}

	return ref
}

func resolveExperimentalDraftDir(ref, filename string) (string, error) {
	if ref == "" || filepath.IsAbs(ref) || filename == "" {
		return ref, nil
	}

	candidate := filepath.Join(filepath.Dir(filename), ref)
	if xcreate.IsSafetensorsModelDir(candidate) || xcreate.IsTensorModelDir(candidate) {
		return candidate, nil
	}

	return ref, nil
}

func checkServerHeartbeat(cmd *cobra.Command, _ []string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}
	if err := client.Heartbeat(cmd.Context()); err != nil {
		if !(strings.Contains(err.Error(), " refused") || strings.Contains(err.Error(), "could not connect")) {
			return err
		}
		if err := startApp(cmd.Context(), client); err != nil {
			return err
		}
	}
	return nil
}

func versionHandler(cmd *cobra.Command, _ []string) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return
	}

	serverVersion, err := client.Version(cmd.Context())
	if err != nil {
		fmt.Println("Warning: could not connect to a running Lychee instance")
	}

	if serverVersion != "" {
		fmt.Printf("lychee version is %s\n", serverVersion)
	}

	if serverVersion != version.Version {
		fmt.Printf("Warning: client version is %s\n", version.Version)
	}
}

func appendEnvDocs(cmd *cobra.Command, envs []envconfig.EnvVar) {
	if len(envs) == 0 {
		return
	}

	envUsage := `
Environment Variables:
`
	for _, e := range envs {
		envUsage += fmt.Sprintf("      %-27s   %s\n", e.Name, e.Description)
	}

	cmd.SetUsageTemplate(cmd.UsageTemplate() + envUsage)
}

// ensureServerRunning checks if the lychee server is running and starts it in the background if not.
func ensureServerRunning(ctx context.Context) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	// Check if server is already running
	if err := client.Heartbeat(ctx); err == nil {
		return nil // server is already running
	}

	// Server not running, start it in the background
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find executable: %w", err)
	}

	serverCmd := exec.CommandContext(ctx, exe, "serve")
	serverCmd.Env = os.Environ()
	serverCmd.SysProcAttr = backgroundServerSysProcAttr()
	if err := serverCmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for the server to be ready
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		if err := client.Heartbeat(ctx); err == nil {
			return nil // server has started
		}
	}
	return fmt.Errorf("server did not start within 15 seconds — run 'lychee serve' manually")
}

// runInteractiveTUI runs the main interactive TUI menu.
func runInteractiveTUI(cmd *cobra.Command) {
	// Ensure the server is running before showing the TUI
	if err := ensureServerRunning(cmd.Context()); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		return
	}

	accountPrefetch := launch.StartAccountStatePrefetch(cmd.Context())
	deps := launcherDeps{
		buildState:          launch.BuildLauncherState,
		runMenu:             tui.RunMenu,
		resolveRunModel:     launch.ResolveRunModel,
		launchIntegration:   launch.LaunchIntegration,
		runModel:            launchInteractiveModel,
		accountState:        accountPrefetch.StateIfReady,
		accountStateUpdates: accountPrefetch.StateUpdates,
	}

	for {
		continueLoop, err := runInteractiveTUIStep(cmd, deps)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		if !continueLoop {
			return
		}
	}
}

type launcherDeps struct {
	buildState          func(context.Context) (*launch.LauncherState, error)
	runMenu             func(*launch.LauncherState) (tui.TUIAction, error)
	resolveRunModel     func(context.Context, launch.RunModelRequest) (string, error)
	launchIntegration   func(context.Context, launch.IntegrationLaunchRequest) error
	runModel            func(*cobra.Command, string) error
	accountState        func() *launch.AccountState
	accountStateUpdates func(context.Context) <-chan *launch.AccountState
}

func runInteractiveTUIStep(cmd *cobra.Command, deps launcherDeps) (bool, error) {
	state, err := deps.buildState(cmd.Context())
	if err != nil {
		return false, fmt.Errorf("build launcher state: %w", err)
	}
	if state != nil && deps.accountState != nil {
		state.AccountState = deps.accountState()
	}

	action, err := deps.runMenu(state)
	if err != nil {
		return false, fmt.Errorf("run launcher menu: %w", err)
	}

	return runLauncherAction(cmd, action, deps)
}

func saveLauncherSelection(action tui.TUIAction) {
	// Best effort only: this affects menu recall, not launch correctness.
	_ = config.SetLastSelection(action.LastSelection())
}

func runLauncherAction(cmd *cobra.Command, action tui.TUIAction, deps launcherDeps) (bool, error) {
	switch action.Kind {
	case tui.TUIActionNone:
		return false, nil
	case tui.TUIActionRunModel:
		saveLauncherSelection(action)
		req := action.RunModelRequest()
		if deps.accountState != nil {
			req.AccountState = deps.accountState()
			req.AccountStateProvider = deps.accountState
		}
		req.AccountStateUpdates = deps.accountStateUpdates
		modelName, err := deps.resolveRunModel(cmd.Context(), req)
		if errors.Is(err, launch.ErrCancelled) {
			return true, nil
		}
		if err != nil {
			return true, fmt.Errorf("selecting model: %w", err)
		}
		if err := deps.runModel(cmd, modelName); err != nil {
			return true, err
		}
		return true, nil
	case tui.TUIActionLaunchIntegration:
		saveLauncherSelection(action)
		req := action.IntegrationLaunchRequest()
		if deps.accountState != nil {
			req.AccountState = deps.accountState()
			req.AccountStateProvider = deps.accountState
		}
		req.AccountStateUpdates = deps.accountStateUpdates
		err := deps.launchIntegration(cmd.Context(), req)
		if errors.Is(err, launch.ErrCancelled) {
			return true, nil
		}
		if err != nil {
			return true, fmt.Errorf("launching %s: %w", action.Integration, err)
		}
		if launcherActionExitsLoop(action.Integration) {
			return false, nil
		}
		return true, nil
	default:
		return false, fmt.Errorf("unknown launcher action: %d", action.Kind)
	}
}

func launcherActionExitsLoop(integration string) bool {
	switch integration {
	case "codex-app", "vscode":
		return true
	default:
		return false
	}
}

func NewCLI() *cobra.Command {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cobra.EnableCommandSorting = false

	if runtime.GOOS == "windows" && term.IsTerminal(int(os.Stdout.Fd())) {
		console.ConsoleFromFile(os.Stdin) //nolint:errcheck
	}

	rootCmd := &cobra.Command{
		Use:   "lychee",
		Short: "Large language model runner",
		Long: `Lychee is a tool for running large language models locally.

Usage:
  lychee [command]

Available Commands:
  run         Run a model
  pull        Pull a model from a registry
  create      Create a model from a Modelfile
  list        List models
  serve       Start Lychee
  completion  Generate shell completion scripts

Use "lychee [command] --help" for more information about a command.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			if version, _ := cmd.Flags().GetBool("version"); version {
				versionHandler(cmd, args)
				return
			}

			runInteractiveTUI(cmd)
		},
	}

	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	rootCmd.Flags().Bool("verbose", false, "Show timings for response")
	rootCmd.Flags().Bool("nowordwrap", false, "Don't wrap words to the next line automatically")

	createCmd := &cobra.Command{
		Use:   "create MODEL",
		Short: "Create a model",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip server check for experimental mode (writes directly to disk)
			if experimental, _ := cmd.Flags().GetBool("experimental"); experimental {
				return nil
			}
			return checkServerHeartbeat(cmd, args)
		},
		RunE: CreateHandler,
	}

	createCmd.Flags().StringP("file", "f", "", "Name of the Modelfile (default \"Modelfile\")")
	createCmd.Flags().StringP("quantize", "q", "", "Quantize model to this level (e.g. q4_K_M)")
	createCmd.Flags().String("draft-quantize", "", "Quantize draft model to this level")
	createCmd.Flags().Bool("experimental", false, "Enable experimental safetensors model creation")

	showCmd := &cobra.Command{
		Use:               "show MODEL",
		Short:             "Show information for a model",
		Args:              cobra.ExactArgs(1),
		PreRunE:           checkServerHeartbeat,
		RunE:              ShowHandler,
		ValidArgsFunction: autocompleteInstalledModels,
	}

	showCmd.Flags().Bool("license", false, "Show license of a model")
	showCmd.Flags().Bool("modelfile", false, "Show Modelfile of a model")
	showCmd.Flags().Bool("parameters", false, "Show parameters of a model")
	showCmd.Flags().Bool("template", false, "Show template of a model")
	showCmd.Flags().Bool("system", false, "Show system message of a model")
	showCmd.Flags().BoolP("verbose", "v", false, "Show detailed model information")

	runCmd := &cobra.Command{
		Use:               "run MODEL [PROMPT]",
		Short:             "Run a model",
		Args:              cobra.MinimumNArgs(1),
		PreRunE:           checkServerHeartbeat,
		RunE:              RunHandler,
		ValidArgsFunction: autocompleteInstalledModels,
	}

	runCmd.Flags().String("keepalive", "", "Duration to keep a model loaded (e.g. 5m)")
	runCmd.Flags().String("draft", "", "Draft model for speculative decoding overrides")
	runCmd.Flags().Bool("verbose", false, "Show timings for response")
	runCmd.Flags().Bool("insecure", false, "Use an insecure registry")
	runCmd.Flags().Bool("nowordwrap", false, "Don't wrap words to the next line automatically")
	runCmd.Flags().String("format", "", "Response format (e.g. json)")
	runCmd.Flags().String("think", "", "Enable thinking mode: true/false or high/medium/low for supported models")
	runCmd.Flags().Lookup("think").NoOptDefVal = "true"
	runCmd.Flags().Bool("hidethinking", false, "Hide thinking output (if provided)")
	runCmd.Flags().Bool("truncate", false, "For embedding models: truncate inputs exceeding context length (default: true). Set --truncate=false to error instead")
	runCmd.Flags().Int("dimensions", 0, "Truncate output embeddings to specified dimension (embedding models only)")
	runCmd.Flags().Bool("experimental", false, "Enable experimental agent loop with tools")
	runCmd.Flags().Bool("experimental-yolo", false, "Skip all tool approval prompts (use with caution)")
	runCmd.Flags().Bool("experimental-websearch", false, "Enable web search tool in experimental mode")

	// Image generation flags (width, height, steps, seed, etc.)
	imagegen.RegisterFlags(runCmd)

	runCmd.Flags().Bool("imagegen", false, "Use the imagegen runner for LLM inference")
	runCmd.Flags().MarkHidden("imagegen")

	stopCmd := &cobra.Command{
		Use:               "stop MODEL",
		Short:             "Stop a running model",
		Args:              cobra.ExactArgs(1),
		PreRunE:           checkServerHeartbeat,
		RunE:              StopHandler,
		ValidArgsFunction: autocompleteRunningModels,
	}

	serveCmd := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"start"},
		Short:   "Start Lychee",
		Args:    cobra.ExactArgs(0),
		RunE:    RunServer,
	}

	serveCmd.Flags().Int("parallel", 0, "Number of parallel request slots (env: LYCHEE_NUM_PARALLEL)")
	serveCmd.Flags().Int("slots", 0, "Number of parallel request slots (alias for --parallel)")
	serveCmd.Flags().String("log-level", "", "Log level: debug, info, warn, error")

	pullCmd := &cobra.Command{
		Use:     "pull MODEL",
		Short:   "Pull a model from a registry",
		Args:    cobra.ExactArgs(1),
		PreRunE: checkServerHeartbeat,
		RunE:    PullHandler,
	}

	pullCmd.Flags().Bool("insecure", false, "Use an insecure registry")
	pullCmd.Flags().String("quant", "", "Quantization to prefer (e.g. q4_k_m, q5_k_m) — for HuggingFace models")
	pullCmd.Flags().Bool("list", false, "List all available quantizations then exit — for HuggingFace models")
	pullCmd.Flags().Bool("run", false, "After pull, automatically create and start an interactive chat")

	pushCmd := &cobra.Command{
		Use:               "push MODEL",
		Short:             "Push a model to a registry",
		Args:              cobra.ExactArgs(1),
		PreRunE:           checkServerHeartbeat,
		RunE:              PushHandler,
		ValidArgsFunction: autocompleteInstalledModels,
	}

	pushCmd.Flags().Bool("insecure", false, "Use an insecure registry")

	signinCmd := &cobra.Command{
		Use:     "signin",
		Short:   "Sign in to lychee.com",
		Args:    cobra.ExactArgs(0),
		PreRunE: checkServerHeartbeat,
		RunE:    SigninHandler,
	}

	loginCmd := &cobra.Command{
		Use:     "login",
		Short:   "Sign in to lychee.com",
		Hidden:  true,
		Args:    cobra.ExactArgs(0),
		PreRunE: checkServerHeartbeat,
		RunE:    SigninHandler,
	}

	signoutCmd := &cobra.Command{
		Use:     "signout",
		Short:   "Sign out from lychee.com",
		Args:    cobra.ExactArgs(0),
		PreRunE: checkServerHeartbeat,
		RunE:    SignoutHandler,
	}

	logoutCmd := &cobra.Command{
		Use:     "logout",
		Short:   "Sign out from lychee.com",
		Hidden:  true,
		Args:    cobra.ExactArgs(0),
		PreRunE: checkServerHeartbeat,
		RunE:    SignoutHandler,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List models",
		PreRunE: checkServerHeartbeat,
		RunE:    ListHandler,
	}

	psCmd := &cobra.Command{
		Use:     "ps",
		Short:   "List running models",
		PreRunE: checkServerHeartbeat,
		RunE:    ListRunningHandler,
	}
	copyCmd := &cobra.Command{
		Use:     "cp SOURCE DESTINATION",
		Short:   "Copy a model",
		Args:    cobra.ExactArgs(2),
		PreRunE: checkServerHeartbeat,
		RunE:    CopyHandler,
	}

	deleteCmd := &cobra.Command{
		Use:               "rm MODEL [MODEL...]",
		Short:             "Remove a model",
		Args:              cobra.MinimumNArgs(1),
		PreRunE:           checkServerHeartbeat,
		RunE:              DeleteHandler,
		ValidArgsFunction: autocompleteInstalledModels,
	}

	runnerCmd := &cobra.Command{
		Use:    "runner",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.Execute(os.Args[1:])
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	runnerCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_ = runner.Execute(args[1:])
	})

	var gpuDiscoverLibDirs []string
	gpuDiscoverCmd := &cobra.Command{
		Use:    "gpu-discover",
		Hidden: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return discover.RunNativeProbeCommand(cmd.Context(), gpuDiscoverLibDirs, os.Stdout)
		},
	}
	gpuDiscoverCmd.Flags().StringArrayVar(&gpuDiscoverLibDirs, "lib-dir", nil, "Lychee runtime library directory")

	envVars := envconfig.AsMap()

	envs := []envconfig.EnvVar{envVars["LYCHEE_HOST"]}

	for _, cmd := range []*cobra.Command{
		createCmd,
		showCmd,
		runCmd,
		stopCmd,
		pullCmd,
		pushCmd,
		listCmd,
		psCmd,
		copyCmd,
		deleteCmd,
		serveCmd,
	} {
		switch cmd {
		case runCmd:
			imagegen.AppendFlagsDocs(cmd)
			appendEnvDocs(cmd, []envconfig.EnvVar{envVars["LYCHEE_EDITOR"], envVars["LYCHEE_HOST"], envVars["LYCHEE_NOHISTORY"]})
		case serveCmd:
			appendEnvDocs(cmd, []envconfig.EnvVar{
				envVars["LYCHEE_DEBUG"],
				envVars["LYCHEE_HOST"],
				envVars["LYCHEE_CONTEXT_LENGTH"],
				envVars["LYCHEE_KEEP_ALIVE"],
				envVars["LYCHEE_MAX_LOADED_MODELS"],
				envVars["LYCHEE_MAX_TRANSFER_STREAMS"],
				envVars["LYCHEE_MAX_QUEUE"],
				envVars["LYCHEE_MODELS"],
				envVars["LYCHEE_NUM_PARALLEL"],
				envVars["LYCHEE_NO_CLOUD"],
				envVars["LYCHEE_NOPRUNE"],
				envVars["LYCHEE_ORIGINS"],
				envVars["LYCHEE_SCHED_SPREAD"],
				envVars["LYCHEE_FLASH_ATTENTION"],
				envVars["LYCHEE_KV_CACHE_TYPE"],
				envVars["LYCHEE_LLM_LIBRARY"],
				envVars["LYCHEE_GPU_OVERHEAD"],
				envVars["LYCHEE_IGPU_ENABLE"],
				envVars["LLAMA_ARG_FIT"],
				envVars["LLAMA_ARG_FIT_TARGET"],
				envVars["LYCHEE_LOAD_TIMEOUT"],
			})
		default:
			appendEnvDocs(cmd, envs)
		}
	}

	rootCmd.AddCommand(
		serveCmd,
		createCmd,
		showCmd,
		runCmd,
		stopCmd,
		pullCmd,
		pushCmd,
		signinCmd,
		loginCmd,
		signoutCmd,
		logoutCmd,
		listCmd,
		psCmd,
		copyCmd,
		deleteCmd,
		runnerCmd,
		gpuDiscoverCmd,
		NewHFCmd(),
		NewAgentCmd(),
		NewEmbedCmd(),
		NewScanCmd(),
		NewCompareCmd(),
		NewStatsCmd(),
		NewSearchCmd(),
		NewBenchCmd(),
		NewBenchmarkCmd(),
		NewModelfileCmd(),
		NewQuantizeCmd(),
		NewInspectCmd(),
		NewCatalogCmd(),
		NewComposeCmd(),
		NewExportCmd(),
		NewImportCmd(),
		NewGenerateClientCmd(),
		NewCommunityCmd(),
		NewCompletionCmd(),
		NewInteractiveCmd(),
		NewUpdateCmd(),
		NewVersionCmd(),
		launch.LaunchCmd(checkServerHeartbeat, runInteractiveTUI),
	)

	return rootCmd
}

func autocompleteInstalledModels(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	models, err := client.List(cmd.Context())
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, m := range models.Models {
		if strings.HasPrefix(strings.ToLower(m.Name), strings.ToLower(toComplete)) {
			names = append(names, m.Name)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func autocompleteRunningModels(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	models, err := client.ListRunning(cmd.Context())
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, m := range models.Models {
		if strings.HasPrefix(strings.ToLower(m.Name), strings.ToLower(toComplete)) {
			names = append(names, m.Name)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
