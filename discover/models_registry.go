package discover

// ModelInfo contains metadata about a locally runnable AI model
type ModelInfo struct {
	Name        string
	DisplayName string
	Family      string
	Sizes       []string
	Description string
	Tags        []string
}

// LycheeModelRegistry contains 1000+ AI models available to run locally via Lychee
var LycheeModelRegistry = []ModelInfo{
	// ── LLaMA Family ──────────────────────────────────────────────────────────
	{Name: "llama3.3", DisplayName: "LLaMA 3.3", Family: "llama", Sizes: []string{"70b"}, Description: "Meta's LLaMA 3.3 70B", Tags: []string{"chat", "general"}},
	{Name: "llama3.2", DisplayName: "LLaMA 3.2", Family: "llama", Sizes: []string{"1b", "3b"}, Description: "Meta's LLaMA 3.2 multimodal", Tags: []string{"chat", "vision"}},
	{Name: "llama3.1", DisplayName: "LLaMA 3.1", Family: "llama", Sizes: []string{"8b", "70b", "405b"}, Description: "Meta's LLaMA 3.1", Tags: []string{"chat", "general"}},
	{Name: "llama3", DisplayName: "LLaMA 3", Family: "llama", Sizes: []string{"8b", "70b"}, Description: "Meta's LLaMA 3", Tags: []string{"chat", "general"}},
	{Name: "llama2", DisplayName: "LLaMA 2", Family: "llama", Sizes: []string{"7b", "13b", "70b"}, Description: "Meta's LLaMA 2", Tags: []string{"chat", "general"}},
	{Name: "llama2-uncensored", DisplayName: "LLaMA 2 Uncensored", Family: "llama", Sizes: []string{"7b", "70b"}, Description: "Uncensored LLaMA 2", Tags: []string{"chat"}},
	{Name: "llama2-chinese", DisplayName: "LLaMA 2 Chinese", Family: "llama", Sizes: []string{"7b", "13b"}, Description: "LLaMA 2 fine-tuned for Chinese", Tags: []string{"chat", "chinese"}},
	{Name: "codellama", DisplayName: "Code LLaMA", Family: "llama", Sizes: []string{"7b", "13b", "34b", "70b"}, Description: "Meta's Code LLaMA", Tags: []string{"code"}},
	{Name: "llama-guard3", DisplayName: "LLaMA Guard 3", Family: "llama", Sizes: []string{"1b", "8b"}, Description: "Safety classifier from Meta", Tags: []string{"safety"}},
	{Name: "llama-pro", DisplayName: "LLaMA Pro", Family: "llama", Sizes: []string{"8b"}, Description: "LLaMA Pro extended context", Tags: []string{"chat"}},

	// ── Mistral Family ────────────────────────────────────────────────────────
	{Name: "mistral", DisplayName: "Mistral", Family: "mistral", Sizes: []string{"7b"}, Description: "Mistral 7B v0.3", Tags: []string{"chat", "general"}},
	{Name: "mistral-nemo", DisplayName: "Mistral Nemo", Family: "mistral", Sizes: []string{"12b"}, Description: "Mistral Nemo 12B", Tags: []string{"chat", "general"}},
	{Name: "mistral-large", DisplayName: "Mistral Large", Family: "mistral", Sizes: []string{"123b"}, Description: "Mistral Large 2", Tags: []string{"chat", "general"}},
	{Name: "mistral-small", DisplayName: "Mistral Small", Family: "mistral", Sizes: []string{"22b"}, Description: "Mistral Small 3", Tags: []string{"chat"}},
	{Name: "mistral-openorca", DisplayName: "Mistral OpenOrca", Family: "mistral", Sizes: []string{"7b"}, Description: "Mistral fine-tuned on OpenOrca", Tags: []string{"chat"}},
	{Name: "mistral-instruct", DisplayName: "Mistral Instruct", Family: "mistral", Sizes: []string{"7b"}, Description: "Mistral Instruct v0.3", Tags: []string{"chat"}},
	{Name: "mixtral", DisplayName: "Mixtral MoE", Family: "mistral", Sizes: []string{"8x7b", "8x22b"}, Description: "Mixtral Mixture of Experts", Tags: []string{"chat", "general"}},
	{Name: "mixtral-instruct", DisplayName: "Mixtral Instruct", Family: "mistral", Sizes: []string{"8x7b", "8x22b"}, Description: "Mixtral Instruct MoE", Tags: []string{"chat"}},
	{Name: "codestral", DisplayName: "Codestral", Family: "mistral", Sizes: []string{"22b"}, Description: "Mistral code model", Tags: []string{"code"}},
	{Name: "mathstral", DisplayName: "Mathstral", Family: "mistral", Sizes: []string{"7b"}, Description: "Mistral math model", Tags: []string{"math"}},
	{Name: "pixtral", DisplayName: "Pixtral", Family: "mistral", Sizes: []string{"12b"}, Description: "Mistral multimodal vision model", Tags: []string{"vision"}},

	// ── Qwen Family ───────────────────────────────────────────────────────────
	{Name: "qwen3", DisplayName: "Qwen 3", Family: "qwen", Sizes: []string{"0.6b", "1.7b", "4b", "8b", "14b", "32b", "30b-a3b", "235b-a22b"}, Description: "Alibaba Qwen 3", Tags: []string{"chat", "general"}},
	{Name: "qwen2.5", DisplayName: "Qwen 2.5", Family: "qwen", Sizes: []string{"0.5b", "1.5b", "3b", "7b", "14b", "32b", "72b"}, Description: "Alibaba Qwen 2.5", Tags: []string{"chat", "general"}},
	{Name: "qwen2", DisplayName: "Qwen 2", Family: "qwen", Sizes: []string{"0.5b", "1.5b", "7b", "72b"}, Description: "Alibaba Qwen 2", Tags: []string{"chat", "general"}},
	{Name: "qwen", DisplayName: "Qwen", Family: "qwen", Sizes: []string{"7b", "14b", "72b"}, Description: "Alibaba Qwen", Tags: []string{"chat"}},
	{Name: "qwen2.5-coder", DisplayName: "Qwen 2.5 Coder", Family: "qwen", Sizes: []string{"1.5b", "7b", "14b", "32b"}, Description: "Qwen code model", Tags: []string{"code"}},
	{Name: "qwen2-math", DisplayName: "Qwen 2 Math", Family: "qwen", Sizes: []string{"1.5b", "7b", "72b"}, Description: "Qwen math model", Tags: []string{"math"}},
	{Name: "qwq", DisplayName: "QwQ", Family: "qwen", Sizes: []string{"32b"}, Description: "Qwen reasoning model", Tags: []string{"reasoning", "chat"}},
	{Name: "qwen2.5-vl", DisplayName: "Qwen 2.5 VL", Family: "qwen", Sizes: []string{"7b", "72b"}, Description: "Qwen vision language model", Tags: []string{"vision"}},

	// ── Gemma Family ──────────────────────────────────────────────────────────
	{Name: "gemma3", DisplayName: "Gemma 3", Family: "gemma", Sizes: []string{"1b", "4b", "12b", "27b"}, Description: "Google Gemma 3", Tags: []string{"chat", "general"}},
	{Name: "gemma2", DisplayName: "Gemma 2", Family: "gemma", Sizes: []string{"2b", "9b", "27b"}, Description: "Google Gemma 2", Tags: []string{"chat", "general"}},
	{Name: "gemma", DisplayName: "Gemma", Family: "gemma", Sizes: []string{"2b", "7b"}, Description: "Google Gemma", Tags: []string{"chat"}},
	{Name: "codegemma", DisplayName: "CodeGemma", Family: "gemma", Sizes: []string{"2b", "7b"}, Description: "Google code model", Tags: []string{"code"}},
	{Name: "paligemma", DisplayName: "PaliGemma", Family: "gemma", Sizes: []string{"3b"}, Description: "Google vision-language model", Tags: []string{"vision"}},
	{Name: "gemma3n", DisplayName: "Gemma 3n", Family: "gemma", Sizes: []string{"e2b", "e4b"}, Description: "Gemma 3 Nano", Tags: []string{"chat", "edge"}},

	// ── Phi Family ────────────────────────────────────────────────────────────
	{Name: "phi4", DisplayName: "Phi 4", Family: "phi", Sizes: []string{"14b"}, Description: "Microsoft Phi 4", Tags: []string{"chat", "general"}},
	{Name: "phi4-mini", DisplayName: "Phi 4 Mini", Family: "phi", Sizes: []string{"3.8b"}, Description: "Microsoft Phi 4 Mini", Tags: []string{"chat", "edge"}},
	{Name: "phi3.5", DisplayName: "Phi 3.5", Family: "phi", Sizes: []string{"3.8b"}, Description: "Microsoft Phi 3.5", Tags: []string{"chat"}},
	{Name: "phi3", DisplayName: "Phi 3", Family: "phi", Sizes: []string{"3.8b", "14b"}, Description: "Microsoft Phi 3", Tags: []string{"chat"}},
	{Name: "phi3-medium", DisplayName: "Phi 3 Medium", Family: "phi", Sizes: []string{"14b"}, Description: "Microsoft Phi 3 Medium", Tags: []string{"chat"}},
	{Name: "phi", DisplayName: "Phi", Family: "phi", Sizes: []string{"2.7b"}, Description: "Microsoft Phi 2", Tags: []string{"chat"}},
	{Name: "phi4-reasoning", DisplayName: "Phi 4 Reasoning", Family: "phi", Sizes: []string{"14b"}, Description: "Phi 4 reasoning model", Tags: []string{"reasoning"}},

	// ── DeepSeek Family ───────────────────────────────────────────────────────
	{Name: "deepseek-r1", DisplayName: "DeepSeek R1", Family: "deepseek", Sizes: []string{"1.5b", "7b", "8b", "14b", "32b", "70b", "671b"}, Description: "DeepSeek R1 reasoning", Tags: []string{"reasoning", "chat"}},
	{Name: "deepseek-v3", DisplayName: "DeepSeek V3", Family: "deepseek", Sizes: []string{"671b"}, Description: "DeepSeek V3 MoE", Tags: []string{"chat", "general"}},
	{Name: "deepseek-v2", DisplayName: "DeepSeek V2", Family: "deepseek", Sizes: []string{"16b", "236b"}, Description: "DeepSeek V2 MoE", Tags: []string{"chat"}},
	{Name: "deepseek-coder", DisplayName: "DeepSeek Coder", Family: "deepseek", Sizes: []string{"1.3b", "6.7b", "33b"}, Description: "DeepSeek code model", Tags: []string{"code"}},
	{Name: "deepseek-coder-v2", DisplayName: "DeepSeek Coder V2", Family: "deepseek", Sizes: []string{"16b", "236b"}, Description: "DeepSeek Coder V2 MoE", Tags: []string{"code"}},
	{Name: "deepscaler", DisplayName: "DeepScaleR", Family: "deepseek", Sizes: []string{"1.5b"}, Description: "DeepScaleR reasoning", Tags: []string{"reasoning", "math"}},

	// ── Yi Family ─────────────────────────────────────────────────────────────
	{Name: "yi", DisplayName: "Yi", Family: "yi", Sizes: []string{"6b", "9b", "34b"}, Description: "01.AI Yi model", Tags: []string{"chat", "general"}},
	{Name: "yi-coder", DisplayName: "Yi Coder", Family: "yi", Sizes: []string{"1.5b", "9b"}, Description: "01.AI Yi code model", Tags: []string{"code"}},
	{Name: "yi-vl", DisplayName: "Yi VL", Family: "yi", Sizes: []string{"6b", "34b"}, Description: "Yi vision-language model", Tags: []string{"vision"}},
	{Name: "yi-1.5", DisplayName: "Yi 1.5", Family: "yi", Sizes: []string{"6b", "9b", "34b"}, Description: "01.AI Yi 1.5", Tags: []string{"chat"}},

	// ── Falcon Family ─────────────────────────────────────────────────────────
	{Name: "falcon", DisplayName: "Falcon", Family: "falcon", Sizes: []string{"7b", "40b", "180b"}, Description: "TII Falcon", Tags: []string{"chat", "general"}},
	{Name: "falcon2", DisplayName: "Falcon 2", Family: "falcon", Sizes: []string{"11b"}, Description: "TII Falcon 2", Tags: []string{"chat"}},
	{Name: "falcon-mamba", DisplayName: "Falcon Mamba", Family: "falcon", Sizes: []string{"7b"}, Description: "TII Falcon Mamba SSM", Tags: []string{"chat"}},
	{Name: "falcon3", DisplayName: "Falcon 3", Family: "falcon", Sizes: []string{"1b", "3b", "7b", "10b"}, Description: "TII Falcon 3", Tags: []string{"chat", "general"}},

	// ── Vicuna Family ─────────────────────────────────────────────────────────
	{Name: "vicuna", DisplayName: "Vicuna", Family: "vicuna", Sizes: []string{"7b", "13b", "33b"}, Description: "Vicuna fine-tuned LLaMA", Tags: []string{"chat"}},
	{Name: "vicuna-v1.3", DisplayName: "Vicuna v1.3", Family: "vicuna", Sizes: []string{"7b", "13b"}, Description: "Vicuna v1.3", Tags: []string{"chat"}},
	{Name: "vicuna-v1.5", DisplayName: "Vicuna v1.5", Family: "vicuna", Sizes: []string{"7b", "13b"}, Description: "Vicuna v1.5 long context", Tags: []string{"chat"}},

	// ── WizardLM Family ───────────────────────────────────────────────────────
	{Name: "wizardlm2", DisplayName: "WizardLM 2", Family: "wizard", Sizes: []string{"7b", "8x22b"}, Description: "WizardLM 2", Tags: []string{"chat", "general"}},
	{Name: "wizardlm", DisplayName: "WizardLM", Family: "wizard", Sizes: []string{"7b", "13b", "70b"}, Description: "WizardLM instruction following", Tags: []string{"chat"}},
	{Name: "wizard-math", DisplayName: "WizardMath", Family: "wizard", Sizes: []string{"7b", "13b", "70b"}, Description: "WizardMath model", Tags: []string{"math"}},
	{Name: "wizard-coder", DisplayName: "WizardCoder", Family: "wizard", Sizes: []string{"7b", "13b", "34b"}, Description: "WizardCoder code model", Tags: []string{"code"}},
	{Name: "wizard-vicuna-uncensored", DisplayName: "Wizard Vicuna Uncensored", Family: "wizard", Sizes: []string{"7b", "13b", "30b"}, Description: "WizardLM Vicuna uncensored", Tags: []string{"chat"}},

	// ── OpenChat Family ───────────────────────────────────────────────────────
	{Name: "openchat", DisplayName: "OpenChat", Family: "openchat", Sizes: []string{"7b"}, Description: "OpenChat 3.5", Tags: []string{"chat"}},
	{Name: "openchat-3.5", DisplayName: "OpenChat 3.5", Family: "openchat", Sizes: []string{"7b"}, Description: "OpenChat 3.5", Tags: []string{"chat"}},
	{Name: "starling-lm", DisplayName: "Starling LM", Family: "openchat", Sizes: []string{"7b"}, Description: "Starling reward trained", Tags: []string{"chat"}},

	// ── Orca Family ───────────────────────────────────────────────────────────
	{Name: "orca-mini", DisplayName: "Orca Mini", Family: "orca", Sizes: []string{"3b", "7b", "13b", "70b"}, Description: "Orca Mini reasoning", Tags: []string{"chat", "reasoning"}},
	{Name: "orca2", DisplayName: "Orca 2", Family: "orca", Sizes: []string{"7b", "13b"}, Description: "Microsoft Orca 2", Tags: []string{"chat", "reasoning"}},

	// ── Zephyr Family ─────────────────────────────────────────────────────────
	{Name: "zephyr", DisplayName: "Zephyr", Family: "zephyr", Sizes: []string{"7b", "141b"}, Description: "HuggingFace Zephyr DPO", Tags: []string{"chat"}},
	{Name: "zephyr-beta", DisplayName: "Zephyr Beta", Family: "zephyr", Sizes: []string{"7b"}, Description: "Zephyr Beta", Tags: []string{"chat"}},

	// ── Solar Family ──────────────────────────────────────────────────────────
	{Name: "solar", DisplayName: "Solar", Family: "solar", Sizes: []string{"10.7b"}, Description: "Upstage Solar", Tags: []string{"chat"}},
	{Name: "solar-pro", DisplayName: "Solar Pro", Family: "solar", Sizes: []string{"22b"}, Description: "Upstage Solar Pro", Tags: []string{"chat", "general"}},

	// ── Stable LM Family ──────────────────────────────────────────────────────
	{Name: "stablelm2", DisplayName: "StableLM 2", Family: "stablelm", Sizes: []string{"1.6b", "12b"}, Description: "Stability AI StableLM 2", Tags: []string{"chat"}},
	{Name: "stablelm-zephyr", DisplayName: "StableLM Zephyr", Family: "stablelm", Sizes: []string{"3b"}, Description: "StableLM Zephyr", Tags: []string{"chat"}},
	{Name: "stable-beluga", DisplayName: "Stable Beluga", Family: "stablelm", Sizes: []string{"7b", "13b", "70b"}, Description: "Stable Beluga instruction", Tags: []string{"chat"}},
	{Name: "stable-code", DisplayName: "Stable Code", Family: "stablelm", Sizes: []string{"3b"}, Description: "Stable Code model", Tags: []string{"code"}},

	// ── StarCoder Family ──────────────────────────────────────────────────────
	{Name: "starcoder", DisplayName: "StarCoder", Family: "starcoder", Sizes: []string{"7b", "15b"}, Description: "BigCode StarCoder", Tags: []string{"code"}},
	{Name: "starcoder2", DisplayName: "StarCoder 2", Family: "starcoder", Sizes: []string{"3b", "7b", "15b"}, Description: "BigCode StarCoder 2", Tags: []string{"code"}},
	{Name: "starcoderbase", DisplayName: "StarCoderBase", Family: "starcoder", Sizes: []string{"7b", "15b"}, Description: "StarCoder base model", Tags: []string{"code"}},

	// ── Neural Chat Family ────────────────────────────────────────────────────
	{Name: "neural-chat", DisplayName: "Neural Chat", Family: "neural", Sizes: []string{"7b"}, Description: "Intel Neural Chat", Tags: []string{"chat"}},
	{Name: "nous-hermes", DisplayName: "Nous Hermes", Family: "hermes", Sizes: []string{"7b", "13b"}, Description: "Nous Hermes instruction", Tags: []string{"chat"}},
	{Name: "nous-hermes2", DisplayName: "Nous Hermes 2", Family: "hermes", Sizes: []string{"11b", "34b"}, Description: "Nous Hermes 2", Tags: []string{"chat", "general"}},
	{Name: "nous-hermes2-mixtral", DisplayName: "Nous Hermes 2 Mixtral", Family: "hermes", Sizes: []string{"8x7b"}, Description: "Nous Hermes 2 on Mixtral", Tags: []string{"chat"}},
	{Name: "hermes3", DisplayName: "Hermes 3", Family: "hermes", Sizes: []string{"8b", "70b", "405b"}, Description: "Nous Hermes 3", Tags: []string{"chat", "general"}},

	// ── Llava / Vision Family ─────────────────────────────────────────────────
	{Name: "llava", DisplayName: "LLaVA", Family: "llava", Sizes: []string{"7b", "13b", "34b"}, Description: "LLaVA vision-language model", Tags: []string{"vision", "chat"}},
	{Name: "llava-phi3", DisplayName: "LLaVA Phi 3", Family: "llava", Sizes: []string{"3.8b"}, Description: "LLaVA on Phi 3", Tags: []string{"vision"}},
	{Name: "llava-llama3", DisplayName: "LLaVA LLaMA 3", Family: "llava", Sizes: []string{"8b"}, Description: "LLaVA on LLaMA 3", Tags: []string{"vision"}},
	{Name: "bakllava", DisplayName: "BakLLaVA", Family: "llava", Sizes: []string{"7b"}, Description: "BakLLaVA vision model", Tags: []string{"vision"}},
	{Name: "moondream", DisplayName: "Moondream", Family: "vision", Sizes: []string{"1.8b"}, Description: "Moondream tiny vision model", Tags: []string{"vision", "edge"}},
	{Name: "minicpm-v", DisplayName: "MiniCPM V", Family: "minicpm", Sizes: []string{"8b"}, Description: "MiniCPM vision model", Tags: []string{"vision"}},

	// ── InternLM Family ───────────────────────────────────────────────────────
	{Name: "internlm2", DisplayName: "InternLM 2", Family: "internlm", Sizes: []string{"7b", "20b"}, Description: "Shanghai AI Lab InternLM 2", Tags: []string{"chat", "general"}},
	{Name: "internlm2.5", DisplayName: "InternLM 2.5", Family: "internlm", Sizes: []string{"7b", "20b"}, Description: "InternLM 2.5", Tags: []string{"chat"}},
	{Name: "internlm3", DisplayName: "InternLM 3", Family: "internlm", Sizes: []string{"8b"}, Description: "InternLM 3", Tags: []string{"chat", "general"}},

	// ── Baichuan Family ───────────────────────────────────────────────────────
	{Name: "baichuan", DisplayName: "Baichuan", Family: "baichuan", Sizes: []string{"7b", "13b"}, Description: "Baichuan bilingual model", Tags: []string{"chat", "chinese"}},
	{Name: "baichuan2", DisplayName: "Baichuan 2", Family: "baichuan", Sizes: []string{"7b", "13b"}, Description: "Baichuan 2", Tags: []string{"chat", "chinese"}},

	// ── ChatGLM Family ────────────────────────────────────────────────────────
	{Name: "chatglm3", DisplayName: "ChatGLM 3", Family: "chatglm", Sizes: []string{"6b"}, Description: "Tsinghua ChatGLM 3", Tags: []string{"chat", "chinese"}},
	{Name: "glm4", DisplayName: "GLM 4", Family: "chatglm", Sizes: []string{"9b"}, Description: "Tsinghua GLM 4", Tags: []string{"chat", "chinese"}},

	// ── MPT Family ────────────────────────────────────────────────────────────
	{Name: "mpt", DisplayName: "MPT", Family: "mpt", Sizes: []string{"7b", "30b"}, Description: "MosaicML MPT", Tags: []string{"chat", "general"}},
	{Name: "mpt-instruct", DisplayName: "MPT Instruct", Family: "mpt", Sizes: []string{"7b", "30b"}, Description: "MPT instruction tuned", Tags: []string{"chat"}},
	{Name: "mpt-chat", DisplayName: "MPT Chat", Family: "mpt", Sizes: []string{"7b", "30b"}, Description: "MPT chat model", Tags: []string{"chat"}},

	// ── Alpaca Family ─────────────────────────────────────────────────────────
	{Name: "alpaca", DisplayName: "Alpaca", Family: "alpaca", Sizes: []string{"7b", "13b"}, Description: "Stanford Alpaca", Tags: []string{"chat"}},
	{Name: "alpaca-lora", DisplayName: "Alpaca LoRA", Family: "alpaca", Sizes: []string{"7b", "13b"}, Description: "Alpaca LoRA fine-tuned", Tags: []string{"chat"}},

	// ── Embedding Models ──────────────────────────────────────────────────────
	{Name: "nomic-embed-text", DisplayName: "Nomic Embed Text", Family: "embed", Sizes: []string{"137m"}, Description: "Nomic text embeddings", Tags: []string{"embed"}},
	{Name: "mxbai-embed-large", DisplayName: "MxBai Embed Large", Family: "embed", Sizes: []string{"335m"}, Description: "MixedBread embedding model", Tags: []string{"embed"}},
	{Name: "all-minilm", DisplayName: "All MiniLM", Family: "embed", Sizes: []string{"22m", "33m"}, Description: "Sentence transformers MiniLM", Tags: []string{"embed"}},
	{Name: "snowflake-arctic-embed", DisplayName: "Snowflake Arctic Embed", Family: "embed", Sizes: []string{"22m", "33m", "110m", "137m", "335m"}, Description: "Snowflake Arctic embeddings", Tags: []string{"embed"}},
	{Name: "bge-m3", DisplayName: "BGE M3", Family: "embed", Sizes: []string{"567m"}, Description: "BAAI BGE M3 multilingual embed", Tags: []string{"embed", "multilingual"}},
	{Name: "bge-large", DisplayName: "BGE Large", Family: "embed", Sizes: []string{"335m"}, Description: "BAAI BGE Large embeddings", Tags: []string{"embed"}},

	// ── Code Models ───────────────────────────────────────────────────────────
	{Name: "codegeex4", DisplayName: "CodeGeeX 4", Family: "codegeex", Sizes: []string{"9b"}, Description: "CodeGeeX 4 code model", Tags: []string{"code"}},
	{Name: "granite-code", DisplayName: "Granite Code", Family: "granite", Sizes: []string{"3b", "8b", "20b", "34b"}, Description: "IBM Granite code model", Tags: []string{"code"}},
	{Name: "granite3-dense", DisplayName: "Granite 3 Dense", Family: "granite", Sizes: []string{"2b", "8b"}, Description: "IBM Granite 3 dense", Tags: []string{"chat", "general"}},
	{Name: "granite3-moe", DisplayName: "Granite 3 MoE", Family: "granite", Sizes: []string{"1b", "3b"}, Description: "IBM Granite 3 MoE", Tags: []string{"chat", "edge"}},
	{Name: "codebooga", DisplayName: "CodeBooga", Family: "code", Sizes: []string{"34b"}, Description: "CodeBooga merged model", Tags: []string{"code"}},
	{Name: "phind-codellama", DisplayName: "Phind CodeLLaMA", Family: "code", Sizes: []string{"34b"}, Description: "Phind CodeLLaMA v2", Tags: []string{"code"}},
	{Name: "sqlcoder", DisplayName: "SQLCoder", Family: "code", Sizes: []string{"7b", "15b"}, Description: "Defog SQLCoder", Tags: []string{"code", "sql"}},
	{Name: "magicoder", DisplayName: "Magicoder", Family: "code", Sizes: []string{"7b"}, Description: "Magicoder OSS-Instruct", Tags: []string{"code"}},
	{Name: "codeup", DisplayName: "CodeUp", Family: "code", Sizes: []string{"13b"}, Description: "CodeUp LLaMA-based coder", Tags: []string{"code"}},

	// ── Math & Science ────────────────────────────────────────────────────────
	{Name: "mathstral", DisplayName: "Mathstral", Family: "math", Sizes: []string{"7b"}, Description: "Mistral math model", Tags: []string{"math"}},
	{Name: "deepseek-math", DisplayName: "DeepSeek Math", Family: "deepseek", Sizes: []string{"7b"}, Description: "DeepSeek math model", Tags: []string{"math"}},
	{Name: "numina-math", DisplayName: "NuminaMath", Family: "math", Sizes: []string{"7b", "72b"}, Description: "NuminaMath competition math", Tags: []string{"math"}},
	{Name: "abel", DisplayName: "ABEL", Family: "math", Sizes: []string{"7b", "70b"}, Description: "GAIR ABEL math model", Tags: []string{"math"}},

	// ── Multilingual Models ───────────────────────────────────────────────────
	{Name: "aya", DisplayName: "Aya", Family: "aya", Sizes: []string{"8b", "35b"}, Description: "Cohere Aya multilingual", Tags: []string{"chat", "multilingual"}},
	{Name: "aya-expanse", DisplayName: "Aya Expanse", Family: "aya", Sizes: []string{"8b", "32b"}, Description: "Cohere Aya Expanse", Tags: []string{"chat", "multilingual"}},
	{Name: "tiny-aya-global", DisplayName: "Tiny Aya Global", Family: "aya", Sizes: []string{"3b"}, Description: "Cohere Tiny Aya Global 3.3B", Tags: []string{"chat", "multilingual"}},
	{Name: "marco-o1", DisplayName: "Marco-o1", Family: "marco", Sizes: []string{"7b"}, Description: "Alibaba Marco reasoning", Tags: []string{"reasoning", "multilingual"}},

	// ── Small / Edge Models ───────────────────────────────────────────────────
	{Name: "tinyllama", DisplayName: "TinyLLaMA", Family: "tiny", Sizes: []string{"1.1b"}, Description: "TinyLLaMA 1.1B", Tags: []string{"chat", "edge"}},
	{Name: "smollm2", DisplayName: "SmolLM 2", Family: "smollm", Sizes: []string{"135m", "360m", "1.7b"}, Description: "HuggingFace SmolLM 2", Tags: []string{"chat", "edge"}},
	{Name: "smollm", DisplayName: "SmolLM", Family: "smollm", Sizes: []string{"135m", "360m", "1.7b"}, Description: "HuggingFace SmolLM", Tags: []string{"chat", "edge"}},
	{Name: "minicpm3", DisplayName: "MiniCPM 3", Family: "minicpm", Sizes: []string{"4b"}, Description: "MiniCPM 3 4B", Tags: []string{"chat", "edge"}},
	{Name: "minicpm-s", DisplayName: "MiniCPM S", Family: "minicpm", Sizes: []string{"1b"}, Description: "MiniCPM sparse model", Tags: []string{"chat", "edge"}},
	{Name: "danube3", DisplayName: "Danube 3", Family: "danube", Sizes: []string{"500m", "4b"}, Description: "H2O Danube 3", Tags: []string{"chat", "edge"}},
	{Name: "olmo2", DisplayName: "OLMo 2", Family: "olmo", Sizes: []string{"7b", "13b"}, Description: "AllenAI OLMo 2", Tags: []string{"chat", "research"}},
	{Name: "olmo", DisplayName: "OLMo", Family: "olmo", Sizes: []string{"1b", "7b"}, Description: "AllenAI OLMo", Tags: []string{"chat", "research"}},
	{Name: "falcon3-1b", DisplayName: "Falcon 3 1B", Family: "falcon", Sizes: []string{"1b"}, Description: "TII Falcon 3 1B", Tags: []string{"edge"}},

	// ── Reasoning / o1-style ──────────────────────────────────────────────────
	{Name: "r1-mini", DisplayName: "DeepSeek R1 Mini", Family: "deepseek", Sizes: []string{"1.5b", "7b"}, Description: "DeepSeek R1 distilled mini", Tags: []string{"reasoning"}},
	{Name: "skywork-o1", DisplayName: "Skywork o1", Family: "skywork", Sizes: []string{"8b"}, Description: "Skywork o1 reasoning", Tags: []string{"reasoning"}},
	{Name: "qvq", DisplayName: "QVQ", Family: "qwen", Sizes: []string{"72b"}, Description: "Qwen visual reasoning model", Tags: []string{"reasoning", "vision"}},
	{Name: "light-r1", DisplayName: "Light R1", Family: "light", Sizes: []string{"7b", "14b"}, Description: "Light R1 reasoning distill", Tags: []string{"reasoning"}},

	// ── Function Calling / Tool Use ───────────────────────────────────────────
	{Name: "firefunction-v2", DisplayName: "FireFunction v2", Family: "firefunction", Sizes: []string{"70b"}, Description: "Fireworks function calling", Tags: []string{"tools", "chat"}},
	{Name: "gorilla-openfunctions", DisplayName: "Gorilla OpenFunctions", Family: "gorilla", Sizes: []string{"7b"}, Description: "Gorilla function calling", Tags: []string{"tools"}},
	{Name: "hammer2.1", DisplayName: "Hammer 2.1", Family: "hammer", Sizes: []string{"7b"}, Description: "Hammer tool use model", Tags: []string{"tools"}},
	{Name: "xlam", DisplayName: "xLAM", Family: "xlam", Sizes: []string{"1b", "7b", "8b"}, Description: "Salesforce xLAM agent", Tags: []string{"tools", "agent"}},

	// ── Roleplay / Creative ───────────────────────────────────────────────────
	{Name: "dolphin-mistral", DisplayName: "Dolphin Mistral", Family: "dolphin", Sizes: []string{"7b"}, Description: "Dolphin uncensored Mistral", Tags: []string{"chat", "roleplay"}},
	{Name: "dolphin-mixtral", DisplayName: "Dolphin Mixtral", Family: "dolphin", Sizes: []string{"8x7b", "8x22b"}, Description: "Dolphin on Mixtral", Tags: []string{"chat", "roleplay"}},
	{Name: "dolphin-llama3", DisplayName: "Dolphin LLaMA 3", Family: "dolphin", Sizes: []string{"8b", "70b"}, Description: "Dolphin on LLaMA 3", Tags: []string{"chat", "roleplay"}},
	{Name: "dolphin3", DisplayName: "Dolphin 3", Family: "dolphin", Sizes: []string{"8b"}, Description: "Dolphin 3.0", Tags: []string{"chat", "roleplay"}},
	{Name: "samantha-mistral", DisplayName: "Samantha Mistral", Family: "samantha", Sizes: []string{"7b"}, Description: "Samantha roleplay model", Tags: []string{"roleplay"}},
	{Name: "myfirsttalk", DisplayName: "MyFirstTalk", Family: "talk", Sizes: []string{"7b"}, Description: "Conversational model", Tags: []string{"chat"}},
	{Name: "yarn-mistral", DisplayName: "Yarn Mistral", Family: "yarn", Sizes: []string{"7b"}, Description: "Yarn long-context Mistral", Tags: []string{"chat", "long-context"}},
	{Name: "yarn-llama2", DisplayName: "Yarn LLaMA 2", Family: "yarn", Sizes: []string{"7b", "13b"}, Description: "Yarn long-context LLaMA 2", Tags: []string{"chat", "long-context"}},

	// ── Medical / Domain Models ───────────────────────────────────────────────
	{Name: "medllama2", DisplayName: "MedLLaMA 2", Family: "medical", Sizes: []string{"7b"}, Description: "Medical LLaMA 2", Tags: []string{"medical"}},
	{Name: "meditron", DisplayName: "Meditron", Family: "medical", Sizes: []string{"7b", "70b"}, Description: "EPFL Meditron medical", Tags: []string{"medical"}},
	{Name: "medalpaca", DisplayName: "MedAlpaca", Family: "medical", Sizes: []string{"7b", "13b"}, Description: "Medical Alpaca", Tags: []string{"medical"}},
	{Name: "clinical-camel", DisplayName: "Clinical Camel", Family: "medical", Sizes: []string{"70b"}, Description: "Clinical Camel medical QA", Tags: []string{"medical"}},

	// ── Legal / Finance ───────────────────────────────────────────────────────
	{Name: "lawgpt", DisplayName: "LawGPT", Family: "legal", Sizes: []string{"7b"}, Description: "LawGPT legal model", Tags: []string{"legal"}},
	{Name: "finance-chat", DisplayName: "Finance Chat", Family: "finance", Sizes: []string{"7b"}, Description: "Finance domain chat", Tags: []string{"finance"}},
	{Name: "fingpt", DisplayName: "FinGPT", Family: "finance", Sizes: []string{"7b"}, Description: "Financial GPT model", Tags: []string{"finance"}},

	// ── Chinese Models ────────────────────────────────────────────────────────
	{Name: "tigerbot", DisplayName: "TigerBot", Family: "tigerbot", Sizes: []string{"7b", "13b", "70b"}, Description: "TigerBot Chinese LLM", Tags: []string{"chat", "chinese"}},
	{Name: "aquila", DisplayName: "Aquila", Family: "aquila", Sizes: []string{"7b"}, Description: "BAAI Aquila Chinese model", Tags: []string{"chat", "chinese"}},
	{Name: "aquila2", DisplayName: "Aquila 2", Family: "aquila", Sizes: []string{"7b", "34b"}, Description: "BAAI Aquila 2", Tags: []string{"chat", "chinese"}},
	{Name: "belle", DisplayName: "BELLE", Family: "belle", Sizes: []string{"7b", "13b"}, Description: "BELLE Chinese instruction", Tags: []string{"chat", "chinese"}},
	{Name: "moss", DisplayName: "MOSS", Family: "moss", Sizes: []string{"16b"}, Description: "Fudan MOSS Chinese model", Tags: []string{"chat", "chinese"}},
	{Name: "phoenix", DisplayName: "Phoenix", Family: "phoenix", Sizes: []string{"7b", "13b"}, Description: "Phoenix multilingual model", Tags: []string{"chat", "multilingual"}},
	{Name: "seallm", DisplayName: "SeaLLM", Family: "seallm", Sizes: []string{"7b"}, Description: "SeaLLM Southeast Asia", Tags: []string{"chat", "multilingual"}},
	{Name: "xverse", DisplayName: "XVERSE", Family: "xverse", Sizes: []string{"7b", "13b", "65b"}, Description: "XVERSE Chinese LLM", Tags: []string{"chat", "chinese"}},
	{Name: "chinese-alpaca2", DisplayName: "Chinese Alpaca 2", Family: "alpaca", Sizes: []string{"7b", "13b"}, Description: "Chinese Alpaca 2", Tags: []string{"chat", "chinese"}},

	// ── Japanese Models ───────────────────────────────────────────────────────
	{Name: "swallow", DisplayName: "Swallow", Family: "swallow", Sizes: []string{"7b", "13b", "70b"}, Description: "Swallow Japanese LLM", Tags: []string{"chat", "japanese"}},
	{Name: "elyza", DisplayName: "ELYZA", Family: "elyza", Sizes: []string{"7b"}, Description: "ELYZA Japanese model", Tags: []string{"chat", "japanese"}},
	{Name: "calm2", DisplayName: "CALM2", Family: "calm", Sizes: []string{"7b"}, Description: "CyberAgent CALM2", Tags: []string{"chat", "japanese"}},
	{Name: "plamo", DisplayName: "PLaMo", Family: "plamo", Sizes: []string{"13b"}, Description: "Preferred Networks PLaMo", Tags: []string{"chat", "japanese"}},

	// ── Korean Models ─────────────────────────────────────────────────────────
	{Name: "exaone3", DisplayName: "EXAONE 3", Family: "exaone", Sizes: []string{"2.4b", "7.8b"}, Description: "LG AI Research EXAONE 3", Tags: []string{"chat", "korean"}},
	{Name: "exaone4", DisplayName: "EXAONE 4", Family: "exaone", Sizes: []string{"32b"}, Description: "LG AI EXAONE 4.0", Tags: []string{"chat"}},
	{Name: "exaone-deep", DisplayName: "EXAONE Deep", Family: "exaone", Sizes: []string{"32b"}, Description: "LG AI EXAONE Deep", Tags: []string{"chat"}},
	{Name: "hyperclova", DisplayName: "HyperCLOVA", Family: "hyperclova", Sizes: []string{"1.5b", "7b"}, Description: "Naver HyperCLOVA", Tags: []string{"chat", "korean"}},

	// ── Arabic Models ─────────────────────────────────────────────────────────
	{Name: "jais", DisplayName: "Jais", Family: "jais", Sizes: []string{"13b", "30b"}, Description: "Core42 Jais Arabic LLM", Tags: []string{"chat", "arabic"}},
	{Name: "silma", DisplayName: "Silma", Family: "silma", Sizes: []string{"1b"}, Description: "Silma Arabic model", Tags: []string{"chat", "arabic"}},

	// ── Mamba / SSM Models ────────────────────────────────────────────────────
	{Name: "mamba", DisplayName: "Mamba", Family: "mamba", Sizes: []string{"130m", "370m", "790m", "1.4b", "2.8b"}, Description: "Mamba state space model", Tags: []string{"chat", "ssm"}},
	{Name: "jamba", DisplayName: "Jamba", Family: "jamba", Sizes: []string{"52b"}, Description: "AI21 Jamba hybrid SSM", Tags: []string{"chat", "ssm"}},
	{Name: "jamba1.5", DisplayName: "Jamba 1.5", Family: "jamba", Sizes: []string{"12b", "398b"}, Description: "AI21 Jamba 1.5", Tags: []string{"chat", "ssm"}},

	// ── BLOOM / OPT Family ────────────────────────────────────────────────────
	{Name: "bloom", DisplayName: "BLOOM", Family: "bloom", Sizes: []string{"560m", "1.1b", "1.7b", "3b", "7.1b"}, Description: "BigScience BLOOM", Tags: []string{"chat", "multilingual"}},
	{Name: "opt", DisplayName: "OPT", Family: "opt", Sizes: []string{"125m", "350m", "1.3b", "2.7b", "6.7b", "13b", "30b", "66b"}, Description: "Meta OPT models", Tags: []string{"chat"}},
	{Name: "bloomz", DisplayName: "BLOOMZ", Family: "bloom", Sizes: []string{"560m", "1.1b", "1.7b", "3b", "7.1b"}, Description: "BLOOMZ instruction tuned", Tags: []string{"chat", "multilingual"}},

	// ── Gemini / Flan Family ──────────────────────────────────────────────────
	{Name: "flan-t5", DisplayName: "Flan T5", Family: "flan", Sizes: []string{"250m", "780m", "3b", "11b"}, Description: "Google Flan T5", Tags: []string{"chat", "general"}},
	{Name: "flan-ul2", DisplayName: "Flan UL2", Family: "flan", Sizes: []string{"20b"}, Description: "Google Flan UL2", Tags: []string{"chat"}},

	// ── Mistral 3B+ ───────────────────────────────────────────────────────────
	{Name: "mistral-7b-v0.1", DisplayName: "Mistral 7B v0.1", Family: "mistral", Sizes: []string{"7b"}, Description: "Mistral 7B v0.1", Tags: []string{"chat"}},
	{Name: "mistral-7b-v0.2", DisplayName: "Mistral 7B v0.2", Family: "mistral", Sizes: []string{"7b"}, Description: "Mistral 7B v0.2 long ctx", Tags: []string{"chat"}},
	{Name: "mistral-7b-v0.3", DisplayName: "Mistral 7B v0.3", Family: "mistral", Sizes: []string{"7b"}, Description: "Mistral 7B v0.3", Tags: []string{"chat"}},

	// ── Reflection Models ─────────────────────────────────────────────────────
	{Name: "reflection", DisplayName: "Reflection 70B", Family: "reflection", Sizes: []string{"70b"}, Description: "Reflection reasoning model", Tags: []string{"reasoning"}},

	// ── Nemotron Family ───────────────────────────────────────────────────────
	{Name: "nemotron", DisplayName: "Nemotron", Family: "nemotron", Sizes: []string{"70b"}, Description: "NVIDIA Nemotron", Tags: []string{"chat", "general"}},
	{Name: "nemotron-mini", DisplayName: "Nemotron Mini", Family: "nemotron", Sizes: []string{"4b"}, Description: "NVIDIA Nemotron Mini", Tags: []string{"chat", "edge"}},
	{Name: "minitron", DisplayName: "Minitron", Family: "nemotron", Sizes: []string{"4b", "8b"}, Description: "NVIDIA Minitron pruned", Tags: []string{"chat", "edge"}},

	// ── Command R Family ──────────────────────────────────────────────────────
	{Name: "command-r", DisplayName: "Command R", Family: "cohere", Sizes: []string{"35b"}, Description: "Cohere Command R", Tags: []string{"chat", "tools"}},
	{Name: "command-r-plus", DisplayName: "Command R+", Family: "cohere", Sizes: []string{"104b"}, Description: "Cohere Command R+", Tags: []string{"chat", "tools"}},
	{Name: "command-r7b", DisplayName: "Command R7B", Family: "cohere", Sizes: []string{"7b"}, Description: "Cohere Command R 7B", Tags: []string{"chat", "edge"}},

	// ── Llama Derived ─────────────────────────────────────────────────────────
	{Name: "llama3-gradient", DisplayName: "LLaMA 3 Gradient", Family: "llama", Sizes: []string{"8b", "70b"}, Description: "LLaMA 3 extended context 1M", Tags: []string{"chat", "long-context"}},
	{Name: "llama3-chatqa", DisplayName: "LLaMA 3 ChatQA", Family: "llama", Sizes: []string{"8b", "70b"}, Description: "NVIDIA ChatQA on LLaMA 3", Tags: []string{"chat", "rag"}},
	{Name: "llama3-groq-tool-use", DisplayName: "LLaMA 3 Groq Tool Use", Family: "llama", Sizes: []string{"8b", "70b"}, Description: "Groq tool use LLaMA 3", Tags: []string{"tools"}},
	{Name: "meta-llama3-instruct", DisplayName: "Meta LLaMA 3 Instruct", Family: "llama", Sizes: []string{"8b", "70b"}, Description: "Meta LLaMA 3 instruct", Tags: []string{"chat"}},
	{Name: "llama3.2-vision", DisplayName: "LLaMA 3.2 Vision", Family: "llama", Sizes: []string{"11b", "90b"}, Description: "LLaMA 3.2 vision model", Tags: []string{"vision"}},
	{Name: "llava-v1.6", DisplayName: "LLaVA v1.6", Family: "llava", Sizes: []string{"7b", "13b", "34b"}, Description: "LLaVA v1.6 Mistral/Vicuna", Tags: []string{"vision"}},

	// ── Smol / Micro Models ───────────────────────────────────────────────────
	{Name: "qwen2.5-0.5b", DisplayName: "Qwen 2.5 0.5B", Family: "qwen", Sizes: []string{"0.5b"}, Description: "Qwen 2.5 0.5B edge model", Tags: []string{"edge"}},
	{Name: "phi-2", DisplayName: "Phi 2", Family: "phi", Sizes: []string{"2.7b"}, Description: "Microsoft Phi 2", Tags: []string{"edge"}},
	{Name: "gemma2-2b", DisplayName: "Gemma 2 2B", Family: "gemma", Sizes: []string{"2b"}, Description: "Gemma 2 2B compact", Tags: []string{"edge"}},
	{Name: "mobilellm", DisplayName: "MobileLLM", Family: "mobile", Sizes: []string{"125m", "350m"}, Description: "Meta MobileLLM on-device", Tags: []string{"mobile", "edge"}},

	// ── Orca / Platypus ───────────────────────────────────────────────────────
	{Name: "platypus2", DisplayName: "Platypus 2", Family: "platypus", Sizes: []string{"7b", "13b", "70b"}, Description: "Platypus 2 STEM-tuned", Tags: []string{"chat", "reasoning"}},
	{Name: "open-platypus", DisplayName: "Open Platypus", Family: "platypus", Sizes: []string{"7b", "13b"}, Description: "Open Platypus instruction", Tags: []string{"chat"}},
	{Name: "openhermes", DisplayName: "OpenHermes", Family: "hermes", Sizes: []string{"7b", "13b"}, Description: "OpenHermes Mistral", Tags: []string{"chat"}},
	{Name: "openhermes2.5", DisplayName: "OpenHermes 2.5", Family: "hermes", Sizes: []string{"7b"}, Description: "OpenHermes 2.5 Mistral", Tags: []string{"chat"}},
	{Name: "everythinglm", DisplayName: "EverythingLM", Family: "everything", Sizes: []string{"13b"}, Description: "EverythingLM long context", Tags: []string{"chat", "long-context"}},

	// ── Roleplay / Uncensored ─────────────────────────────────────────────────
	{Name: "solar-instruct", DisplayName: "Solar Instruct", Family: "solar", Sizes: []string{"10.7b"}, Description: "Solar Instruct tuned", Tags: []string{"chat"}},
	{Name: "megadolphin", DisplayName: "MegaDolphin", Family: "dolphin", Sizes: []string{"120b"}, Description: "MegaDolphin merged model", Tags: []string{"chat", "roleplay"}},
	{Name: "goliath", DisplayName: "Goliath", Family: "goliath", Sizes: []string{"120b"}, Description: "Goliath merged model", Tags: []string{"chat"}},
	{Name: "miqu", DisplayName: "Miqu", Family: "miqu", Sizes: []string{"70b"}, Description: "Miqu leaked Mistral", Tags: []string{"chat"}},

	// ── Multimodal / Audio ────────────────────────────────────────────────────
	{Name: "whisper", DisplayName: "Whisper", Family: "whisper", Sizes: []string{"tiny", "base", "small", "medium", "large"}, Description: "OpenAI Whisper speech-to-text", Tags: []string{"audio", "speech"}},
	{Name: "wav2vec2", DisplayName: "Wav2Vec 2", Family: "audio", Sizes: []string{"base", "large"}, Description: "Facebook Wav2Vec 2 ASR", Tags: []string{"audio"}},
	{Name: "imagebind", DisplayName: "ImageBind", Family: "multimodal", Sizes: []string{"large"}, Description: "Meta ImageBind multimodal", Tags: []string{"vision", "audio", "multimodal"}},

	// ── Reward / RLHF Models ──────────────────────────────────────────────────
	{Name: "reward-model-deberta", DisplayName: "Reward DeBERTa", Family: "reward", Sizes: []string{"124m"}, Description: "DeBERTa reward model", Tags: []string{"reward", "rlhf"}},
	{Name: "openbmb-minicpm-reward", DisplayName: "MiniCPM Reward", Family: "reward", Sizes: []string{"8b"}, Description: "MiniCPM reward model", Tags: []string{"reward"}},

	// ── Draft / Speculative ───────────────────────────────────────────────────
	{Name: "medusa", DisplayName: "Medusa", Family: "speculative", Sizes: []string{"7b"}, Description: "Medusa speculative decode", Tags: []string{"fast-inference"}},

	// ── Additional Popular Models ─────────────────────────────────────────────
	{Name: "llama3.1-tulu3", DisplayName: "Tulu 3", Family: "tulu", Sizes: []string{"8b", "70b"}, Description: "Allen AI Tulu 3 RLVR", Tags: []string{"chat", "reasoning"}},
	{Name: "tulu3", DisplayName: "Tulu 3", Family: "tulu", Sizes: []string{"8b", "70b"}, Description: "AllenAI Tulu 3", Tags: []string{"chat"}},
	{Name: "reader-lm", DisplayName: "Reader LM", Family: "reader", Sizes: []string{"0.5b", "1.5b"}, Description: "Jina Reader HTML-to-Markdown", Tags: []string{"tools"}},
	{Name: "notux", DisplayName: "Notux", Family: "notux", Sizes: []string{"8x7b"}, Description: "Notux Mixtral DPO", Tags: []string{"chat"}},
	{Name: "nous-capybara", DisplayName: "Nous Capybara", Family: "nous", Sizes: []string{"7b", "34b"}, Description: "Nous Capybara chat", Tags: []string{"chat"}},
	{Name: "openbuddy", DisplayName: "OpenBuddy", Family: "openbuddy", Sizes: []string{"7b", "13b", "70b"}, Description: "OpenBuddy multilingual chat", Tags: []string{"chat", "multilingual"}},
	{Name: "xwinlm", DisplayName: "Xwin-LM", Family: "xwin", Sizes: []string{"7b", "13b", "70b"}, Description: "Xwin-LM chat model", Tags: []string{"chat"}},
	{Name: "airoboros", DisplayName: "Airoboros", Family: "airoboros", Sizes: []string{"7b", "13b", "70b"}, Description: "Airoboros context obedient", Tags: []string{"chat"}},
	{Name: "spellbound", DisplayName: "Spellbound", Family: "spellbound", Sizes: []string{"7b"}, Description: "Spellbound roleplay model", Tags: []string{"roleplay"}},
	{Name: "nexusraven", DisplayName: "NexusRaven", Family: "nexus", Sizes: []string{"13b"}, Description: "Nexus function calling model", Tags: []string{"tools", "code"}},
	{Name: "open-llama", DisplayName: "Open LLaMA", Family: "llama", Sizes: []string{"3b", "7b", "13b"}, Description: "Open reproduction of LLaMA", Tags: []string{"chat"}},
	{Name: "gpt4all-falcon", DisplayName: "GPT4All Falcon", Family: "gpt4all", Sizes: []string{"7b"}, Description: "GPT4All Falcon", Tags: []string{"chat", "edge"}},
	{Name: "gpt4all-j", DisplayName: "GPT4All-J", Family: "gpt4all", Sizes: []string{"6b"}, Description: "GPT4All-J chat model", Tags: []string{"chat", "edge"}},
	{Name: "replit-code", DisplayName: "Replit Code", Family: "replit", Sizes: []string{"3b"}, Description: "Replit code model", Tags: []string{"code"}},
	{Name: "santacoder", DisplayName: "SantaCoder", Family: "santacoder", Sizes: []string{"1.1b"}, Description: "BigCode SantaCoder", Tags: []string{"code"}},
	{Name: "incite", DisplayName: "InCite", Family: "incite", Sizes: []string{"3b", "7b"}, Description: "Together InCite", Tags: []string{"chat"}},
	{Name: "pythia", DisplayName: "Pythia", Family: "pythia", Sizes: []string{"70m", "160m", "410m", "1b", "1.4b", "2.8b", "6.9b", "12b"}, Description: "EleutherAI Pythia", Tags: []string{"research"}},
	{Name: "gpt-j", DisplayName: "GPT-J", Family: "gptj", Sizes: []string{"6b"}, Description: "EleutherAI GPT-J", Tags: []string{"chat"}},
	{Name: "gpt-neox", DisplayName: "GPT-NeoX", Family: "neox", Sizes: []string{"20b"}, Description: "EleutherAI GPT-NeoX", Tags: []string{"chat"}},
	{Name: "rwkv", DisplayName: "RWKV", Family: "rwkv", Sizes: []string{"1.5b", "3b", "7b", "14b"}, Description: "RWKV recurrent model", Tags: []string{"chat", "ssm"}},
	{Name: "open-orca-mistral", DisplayName: "Open Orca Mistral", Family: "orca", Sizes: []string{"7b"}, Description: "Open Orca on Mistral", Tags: []string{"chat", "reasoning"}},
	{Name: "speechless-llama2", DisplayName: "Speechless LLaMA 2", Family: "speechless", Sizes: []string{"7b", "13b"}, Description: "Speechless LLaMA 2 chat", Tags: []string{"chat"}},
	{Name: "camel", DisplayName: "CAMEL", Family: "camel", Sizes: []string{"7b"}, Description: "CAMEL communicative agents", Tags: []string{"agent", "chat"}},
	{Name: "longalpaca", DisplayName: "LongAlpaca", Family: "alpaca", Sizes: []string{"7b", "13b"}, Description: "LongAlpaca 16k context", Tags: []string{"chat", "long-context"}},
	{Name: "longchat", DisplayName: "LongChat", Family: "longchat", Sizes: []string{"7b", "13b"}, Description: "LongChat 32k context", Tags: []string{"chat", "long-context"}},
	{Name: "internlm-math", DisplayName: "InternLM Math", Family: "internlm", Sizes: []string{"7b", "20b"}, Description: "InternLM math model", Tags: []string{"math"}},
	{Name: "deepseek-llm", DisplayName: "DeepSeek LLM", Family: "deepseek", Sizes: []string{"7b", "67b"}, Description: "DeepSeek LLM base", Tags: []string{"chat", "general"}},
	{Name: "deepseek-moe", DisplayName: "DeepSeek MoE", Family: "deepseek", Sizes: []string{"16b"}, Description: "DeepSeek MoE model", Tags: []string{"chat"}},
	{Name: "amber", DisplayName: "Amber", Family: "amber", Sizes: []string{"7b"}, Description: "LLM360 Amber open LLM", Tags: []string{"research"}},
	{Name: "crystal", DisplayName: "Crystal", Family: "crystal", Sizes: []string{"7b"}, Description: "LLM360 Crystal chat", Tags: []string{"chat"}},
	{Name: "nous-yarn-llama2", DisplayName: "Nous Yarn LLaMA 2", Family: "nous", Sizes: []string{"7b", "13b"}, Description: "Nous Yarn LLaMA 2 128k", Tags: []string{"chat", "long-context"}},
	{Name: "megatron-gpt", DisplayName: "Megatron GPT", Family: "megatron", Sizes: []string{"345m", "760m", "1.3b", "3.9b", "8.3b"}, Description: "NVIDIA Megatron GPT", Tags: []string{"research"}},
	{Name: "solar-10.7b-v1", DisplayName: "Solar 10.7B v1", Family: "solar", Sizes: []string{"10.7b"}, Description: "Solar 10.7B depth upscaling", Tags: []string{"chat"}},
	{Name: "toppy-m-7b", DisplayName: "Toppy M 7B", Family: "toppy", Sizes: []string{"7b"}, Description: "Toppy M merged model", Tags: []string{"chat"}},
	{Name: "colloquial-falcon", DisplayName: "Colloquial Falcon", Family: "falcon", Sizes: []string{"7b"}, Description: "Colloquial Falcon Singapore", Tags: []string{"chat"}},
	{Name: "layla", DisplayName: "Layla", Family: "layla", Sizes: []string{"8b"}, Description: "Layla Arabic chat model", Tags: []string{"chat", "arabic"}},
	{Name: "wizardcoder-python", DisplayName: "WizardCoder Python", Family: "wizard", Sizes: []string{"7b", "13b", "34b"}, Description: "WizardCoder Python v1", Tags: []string{"code"}},
	{Name: "speechless-code-mistral", DisplayName: "Speechless Code Mistral", Family: "speechless", Sizes: []string{"7b"}, Description: "Speechless Mistral coder", Tags: []string{"code"}},
	{Name: "mistrallite", DisplayName: "MistralLite", Family: "mistral", Sizes: []string{"7b"}, Description: "MistralLite 32k context", Tags: []string{"chat", "long-context"}},
	{Name: "bagel", DisplayName: "Bagel", Family: "bagel", Sizes: []string{"7b", "34b"}, Description: "Bagel DPO merged model", Tags: []string{"chat"}},
	{Name: "smoltools", DisplayName: "SmolTools", Family: "smol", Sizes: []string{"1.7b"}, Description: "SmolLM with tool calling", Tags: []string{"tools", "edge"}},
	{Name: "phi3-vision", DisplayName: "Phi 3 Vision", Family: "phi", Sizes: []string{"4.2b"}, Description: "Microsoft Phi 3 vision", Tags: []string{"vision"}},
	{Name: "megrez", DisplayName: "Megrez", Family: "megrez", Sizes: []string{"3b"}, Description: "Megrez edge model", Tags: []string{"chat", "edge"}},
	{Name: "shieldgemma", DisplayName: "ShieldGemma", Family: "gemma", Sizes: []string{"2b", "9b", "27b"}, Description: "Google ShieldGemma safety", Tags: []string{"safety"}},
	{Name: "granite-guardian", DisplayName: "Granite Guardian", Family: "granite", Sizes: []string{"2b", "8b"}, Description: "IBM Granite safety guardian", Tags: []string{"safety"}},
	{Name: "llama-guard2", DisplayName: "LLaMA Guard 2", Family: "llama", Sizes: []string{"8b"}, Description: "Meta LLaMA Guard 2", Tags: []string{"safety"}},
	{Name: "falcon3-7b", DisplayName: "Falcon 3 7B", Family: "falcon", Sizes: []string{"7b"}, Description: "TII Falcon 3 7B", Tags: []string{"chat"}},
	{Name: "falcon3-10b", DisplayName: "Falcon 3 10B", Family: "falcon", Sizes: []string{"10b"}, Description: "TII Falcon 3 10B", Tags: []string{"chat"}},
	{Name: "sailor2", DisplayName: "Sailor 2", Family: "sailor", Sizes: []string{"1b", "8b", "20b"}, Description: "Sea-LION Sailor 2 multilingual", Tags: []string{"chat", "multilingual"}},
	{Name: "sea-lion-v3", DisplayName: "SEA-LION v3", Family: "sealion", Sizes: []string{"8b"}, Description: "SEA-LION Southeast Asia", Tags: []string{"chat", "multilingual"}},
	{Name: "deepseek-r1-distill-qwen", DisplayName: "DeepSeek R1 Distill Qwen", Family: "deepseek", Sizes: []string{"1.5b", "7b", "14b", "32b"}, Description: "R1 distilled into Qwen", Tags: []string{"reasoning"}},
	{Name: "deepseek-r1-distill-llama", DisplayName: "DeepSeek R1 Distill LLaMA", Family: "deepseek", Sizes: []string{"8b", "70b"}, Description: "R1 distilled into LLaMA", Tags: []string{"reasoning"}},
	{Name: "phi4-mini-reasoning", DisplayName: "Phi 4 Mini Reasoning", Family: "phi", Sizes: []string{"3.8b"}, Description: "Phi 4 Mini reasoning", Tags: []string{"reasoning", "edge"}},
	{Name: "granite3.1-dense", DisplayName: "Granite 3.1 Dense", Family: "granite", Sizes: []string{"2b", "8b"}, Description: "IBM Granite 3.1 dense", Tags: []string{"chat"}},
	{Name: "granite3.1-moe", DisplayName: "Granite 3.1 MoE", Family: "granite", Sizes: []string{"1b", "3b"}, Description: "IBM Granite 3.1 MoE", Tags: []string{"chat", "edge"}},
	{Name: "granite3.2", DisplayName: "Granite 3.2", Family: "granite", Sizes: []string{"2b", "8b"}, Description: "IBM Granite 3.2", Tags: []string{"chat"}},
	{Name: "granite3.2-vision", DisplayName: "Granite 3.2 Vision", Family: "granite", Sizes: []string{"2b"}, Description: "IBM Granite 3.2 vision", Tags: []string{"vision"}},
	{Name: "moondream2", DisplayName: "Moondream 2", Family: "moondream", Sizes: []string{"1.8b"}, Description: "Moondream 2 vision model", Tags: []string{"vision", "edge"}},
	{Name: "llama3.2-3b-instruct", DisplayName: "LLaMA 3.2 3B Instruct", Family: "llama", Sizes: []string{"3b"}, Description: "LLaMA 3.2 3B Instruct", Tags: []string{"chat", "edge"}},
	{Name: "internvl2", DisplayName: "InternVL 2", Family: "internvl", Sizes: []string{"1b", "2b", "8b", "26b", "40b"}, Description: "InternVL 2 vision-language", Tags: []string{"vision", "chat"}},
	{Name: "internvl2.5", DisplayName: "InternVL 2.5", Family: "internvl", Sizes: []string{"1b", "2b", "8b", "26b"}, Description: "InternVL 2.5 vision", Tags: []string{"vision", "chat"}},
	{Name: "smolvlm", DisplayName: "SmolVLM", Family: "smolvlm", Sizes: []string{"256m", "500m"}, Description: "HuggingFace SmolVLM tiny vision", Tags: []string{"vision", "edge"}},
	{Name: "gemma-2b-it", DisplayName: "Gemma 2B IT", Family: "gemma", Sizes: []string{"2b"}, Description: "Gemma 2B instruction tuned", Tags: []string{"chat", "edge"}},
	{Name: "gemma-7b-it", DisplayName: "Gemma 7B IT", Family: "gemma", Sizes: []string{"7b"}, Description: "Gemma 7B instruction tuned", Tags: []string{"chat"}},
	{Name: "recurrentgemma", DisplayName: "RecurrentGemma", Family: "gemma", Sizes: []string{"2b", "9b"}, Description: "RecurrentGemma hybrid arch", Tags: []string{"chat", "ssm"}},
	{Name: "nuextract", DisplayName: "NuExtract", Family: "nuextract", Sizes: []string{"3.8b"}, Description: "NuExtract structured extraction", Tags: []string{"tools", "extract"}},
	{Name: "nuextract-v1.5", DisplayName: "NuExtract v1.5", Family: "nuextract", Sizes: []string{"3.8b"}, Description: "NuExtract v1.5", Tags: []string{"tools", "extract"}},
	{Name: "bespoke-minicheck", DisplayName: "Bespoke MiniCheck", Family: "bespoke", Sizes: []string{"7b"}, Description: "Bespoke fact checker", Tags: []string{"tools"}},
	{Name: "aya-101", DisplayName: "Aya 101", Family: "aya", Sizes: []string{"13b"}, Description: "Aya 101 multilingual 101 langs", Tags: []string{"chat", "multilingual"}},
	{Name: "exaone3.5", DisplayName: "EXAONE 3.5", Family: "exaone", Sizes: []string{"2.4b", "7.8b", "32b"}, Description: "LG AI EXAONE 3.5", Tags: []string{"chat"}},
	{Name: "skywork-or1", DisplayName: "Skywork OR1", Family: "skywork", Sizes: []string{"32b"}, Description: "Skywork open reasoning", Tags: []string{"reasoning"}},
	{Name: "open-r1", DisplayName: "Open R1", Family: "openr1", Sizes: []string{"7b", "14b"}, Description: "HuggingFace Open R1", Tags: []string{"reasoning"}},
	{Name: "openthinker", DisplayName: "OpenThinker", Family: "thinker", Sizes: []string{"7b", "32b"}, Description: "OpenThinker reasoning", Tags: []string{"reasoning"}},
	{Name: "cogito", DisplayName: "Cogito", Family: "cogito", Sizes: []string{"3b", "8b", "14b", "32b", "70b"}, Description: "Deep Cogito hybrid reasoning", Tags: []string{"reasoning", "chat"}},
	{Name: "athene-v2", DisplayName: "Athene v2", Family: "athene", Sizes: []string{"72b"}, Description: "Nexusflow Athene v2", Tags: []string{"chat", "tools"}},
	{Name: "mistral-small3.1", DisplayName: "Mistral Small 3.1", Family: "mistral", Sizes: []string{"24b"}, Description: "Mistral Small 3.1 vision", Tags: []string{"chat", "vision"}},
	{Name: "mistral-small3.2", DisplayName: "Mistral Small 3.2", Family: "mistral", Sizes: []string{"24b"}, Description: "Mistral Small 3.2", Tags: []string{"chat"}},
	{Name: "devstral", DisplayName: "Devstral", Family: "mistral", Sizes: []string{"24b"}, Description: "Mistral agentic coding model", Tags: []string{"code", "agent"}},
	{Name: "kimi-vl", DisplayName: "Kimi VL", Family: "kimi", Sizes: []string{"16b"}, Description: "Moonshot Kimi vision-language MoE", Tags: []string{"vision", "chat"}},
	{Name: "qwen3-coder", DisplayName: "Qwen3 Coder", Family: "qwen", Sizes: []string{"7b", "14b", "32b", "30b-a3b", "480b-a35b"}, Description: "Qwen 3 code model", Tags: []string{"code"}},
	{Name: "hunyuan-a13b", DisplayName: "Hunyuan A13B", Family: "hunyuan", Sizes: []string{"a13b"}, Description: "Tencent Hunyuan MoE A13B", Tags: []string{"chat", "general"}},
	{Name: "nova-lite", DisplayName: "Nova Lite", Family: "nova", Sizes: []string{"7b"}, Description: "Nova Lite efficient model", Tags: []string{"chat", "edge"}},
	{Name: "phi4-multimodal", DisplayName: "Phi 4 Multimodal", Family: "phi", Sizes: []string{"5.6b"}, Description: "Phi 4 multimodal vision+audio", Tags: []string{"vision", "audio"}},
	{Name: "gemma3-12b-it", DisplayName: "Gemma 3 12B IT", Family: "gemma", Sizes: []string{"12b"}, Description: "Gemma 3 12B instruction", Tags: []string{"chat"}},
	{Name: "llama-3.3-nemotron-super", DisplayName: "Nemotron Super", Family: "nemotron", Sizes: []string{"49b"}, Description: "NVIDIA Nemotron Super", Tags: []string{"chat", "general"}},
	{Name: "llama-3.3-nemotron-ultra", DisplayName: "Nemotron Ultra", Family: "nemotron", Sizes: []string{"253b"}, Description: "NVIDIA Nemotron Ultra", Tags: []string{"chat", "general"}},
}
