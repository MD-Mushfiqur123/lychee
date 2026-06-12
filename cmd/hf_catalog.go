package cmd

// hfCatalogEntries returns [name, hfrepo, tags, sizeGB] for all catalog models
func hfCatalogEntries() [][4]string {
	return [][4]string{
		// ── LLaMA ─────────────────────────────────────────────────────────────
		{"llama3.1:8b", "bartowski/Meta-Llama-3.1-8B-Instruct-GGUF", "chat,general", "4.9"},
		{"llama3.1:70b", "bartowski/Meta-Llama-3.1-70B-Instruct-GGUF", "chat,general", "42.5"},
		{"llama3.1:405b", "bartowski/Meta-Llama-3.1-405B-Instruct-GGUF", "chat,general", "230.0"},
		{"llama3.2:1b", "bartowski/Llama-3.2-1B-Instruct-GGUF", "chat,edge", "0.8"},
		{"llama3.2:3b", "bartowski/Llama-3.2-3B-Instruct-GGUF", "chat,edge", "2.0"},
		{"llama3.3:70b", "bartowski/Llama-3.3-70B-Instruct-GGUF", "chat,general", "43.0"},
		{"llama3:8b", "bartowski/Meta-Llama-3-8B-Instruct-GGUF", "chat", "5.0"},
		{"llama3:70b", "bartowski/Meta-Llama-3-70B-Instruct-GGUF", "chat,general", "43.0"},
		{"llama2:7b", "TheBloke/Llama-2-7B-Chat-GGUF", "chat", "4.1"},
		{"llama2:13b", "TheBloke/Llama-2-13B-Chat-GGUF", "chat", "7.9"},
		{"llama2:70b", "TheBloke/Llama-2-70B-Chat-GGUF", "chat", "41.4"},
		{"llama3.2-vision:11b", "bartowski/Llama-3.2-11B-Vision-Instruct-GGUF", "vision,chat", "6.8"},
		{"llama3.2-vision:90b", "bartowski/Llama-3.2-90B-Vision-Instruct-GGUF", "vision,chat", "55.4"},
		{"llama3-chatqa:8b", "bartowski/Llama3-ChatQA-1.5-8B-GGUF", "chat,rag", "5.0"},
		{"llama3-chatqa:70b", "bartowski/Llama3-ChatQA-1.5-70B-GGUF", "chat,rag", "43.0"},
		{"codellama:7b", "TheBloke/CodeLlama-7B-Instruct-GGUF", "code", "4.1"},
		{"codellama:13b", "TheBloke/CodeLlama-13B-Instruct-GGUF", "code", "7.9"},
		{"codellama:34b", "TheBloke/CodeLlama-34B-Instruct-GGUF", "code", "20.7"},
		{"codellama:70b", "TheBloke/CodeLlama-70B-Instruct-hf-GGUF", "code", "41.4"},
		// ── Mistral ───────────────────────────────────────────────────────────
		{"mistral:7b", "bartowski/Mistral-7B-Instruct-v0.3-GGUF", "chat", "4.4"},
		{"mistral-nemo:12b", "bartowski/Mistral-Nemo-Instruct-2407-GGUF", "chat", "7.3"},
		{"mistral-small:22b", "bartowski/Mistral-Small-Instruct-2409-GGUF", "chat", "13.5"},
		{"mistral-large:123b", "bartowski/Mistral-Large-Instruct-2407-GGUF", "chat,general", "75.5"},
		{"mixtral:8x7b", "TheBloke/Mixtral-8x7B-Instruct-v0.1-GGUF", "chat,general", "26.4"},
		{"mixtral:8x22b", "bartowski/Mixtral-8x22B-Instruct-v0.1-GGUF", "chat,general", "79.9"},
		{"codestral:22b", "bartowski/Codestral-22B-v0.1-GGUF", "code", "13.5"},
		{"mistral-small3.1:24b", "bartowski/Mistral-Small-3.1-24B-Instruct-2503-GGUF", "chat,vision", "14.8"},
		{"mistral-openorca:7b", "TheBloke/Mistral-7B-OpenOrca-GGUF", "chat", "4.4"},
		// ── Qwen ──────────────────────────────────────────────────────────────
		{"qwen2.5:0.5b", "Qwen/Qwen2.5-0.5B-Instruct-GGUF", "chat,edge", "0.4"},
		{"qwen2.5:1.5b", "Qwen/Qwen2.5-1.5B-Instruct-GGUF", "chat,edge", "1.0"},
		{"qwen2.5:3b", "Qwen/Qwen2.5-3B-Instruct-GGUF", "chat,edge", "2.0"},
		{"qwen2.5:7b", "Qwen/Qwen2.5-7B-Instruct-GGUF", "chat", "4.7"},
		{"qwen2.5:14b", "bartowski/Qwen2.5-14B-Instruct-GGUF", "chat", "8.6"},
		{"qwen2.5:32b", "bartowski/Qwen2.5-32B-Instruct-GGUF", "chat,general", "19.8"},
		{"qwen2.5:72b", "bartowski/Qwen2.5-72B-Instruct-GGUF", "chat,general", "44.5"},
		{"qwen3:0.6b", "bartowski/Qwen3-0.6B-GGUF", "chat,edge", "0.4"},
		{"qwen3:1.7b", "bartowski/Qwen3-1.7B-GGUF", "chat,edge", "1.1"},
		{"qwen3:4b", "bartowski/Qwen3-4B-GGUF", "chat,edge", "2.6"},
		{"qwen3:8b", "bartowski/Qwen3-8B-GGUF", "chat,reasoning", "5.2"},
		{"qwen3:14b", "bartowski/Qwen3-14B-GGUF", "chat,reasoning", "9.3"},
		{"qwen3:32b", "bartowski/Qwen3-32B-GGUF", "chat,reasoning", "20.4"},
		{"qwen3:30b-a3b", "bartowski/Qwen3-30B-A3B-GGUF", "chat,general", "18.0"},
		{"qwen3:235b-a22b", "bartowski/Qwen3-235B-A22B-GGUF", "chat,general", "140.0"},
		{"qwen3-coder:7b", "bartowski/Qwen3-Coder-7B-Instruct-GGUF", "code", "4.8"},
		{"qwen3-coder:14b", "bartowski/Qwen3-Coder-14B-Instruct-GGUF", "code", "9.3"},
		{"qwen3-coder:32b", "bartowski/Qwen3-Coder-32B-Instruct-GGUF", "code", "20.4"},
		{"qwen3-coder:30b-a3b", "bartowski/Qwen3-Coder-30B-A3B-Instruct-GGUF", "code", "18.0"},
		{"qwen3-coder:480b-a35b", "bartowski/Qwen3-Coder-480B-A35B-Instruct-GGUF", "code", "270.0"},
		{"qwq:32b", "bartowski/QwQ-32B-GGUF", "reasoning", "20.4"},
		{"qwen2.5-coder:1.5b", "Qwen/Qwen2.5-Coder-1.5B-Instruct-GGUF", "code,edge", "1.0"},
		{"qwen2.5-coder:7b", "Qwen/Qwen2.5-Coder-7B-Instruct-GGUF", "code", "4.7"},
		{"qwen2.5-coder:14b", "bartowski/Qwen2.5-Coder-14B-Instruct-GGUF", "code", "8.6"},
		{"qwen2.5-coder:32b", "bartowski/Qwen2.5-Coder-32B-Instruct-GGUF", "code", "19.8"},
		{"qwen2:0.5b", "Qwen/Qwen2-0.5B-Instruct-GGUF", "chat,edge", "0.4"},
		{"qwen2:1.5b", "Qwen/Qwen2-1.5B-Instruct-GGUF", "chat,edge", "1.0"},
		{"qwen2:7b", "Qwen/Qwen2-7B-Instruct-GGUF", "chat", "4.7"},
		{"qwen2:72b", "bartowski/Qwen2-72B-Instruct-GGUF", "chat,general", "44.5"},
		{"qvq:72b", "bartowski/QVQ-72B-Preview-GGUF", "vision,reasoning", "44.5"},
		// ── Gemma ─────────────────────────────────────────────────────────────
		{"gemma3:1b", "bartowski/gemma-3-1b-it-GGUF", "chat,edge", "0.8"},
		{"gemma3:4b", "bartowski/gemma-3-4b-it-GGUF", "chat", "2.5"},
		{"gemma3:12b", "bartowski/gemma-3-12b-it-GGUF", "chat", "7.5"},
		{"gemma3:27b", "bartowski/gemma-3-27b-it-GGUF", "chat,general", "17.0"},
		{"gemma2:2b", "bartowski/gemma-2-2b-it-GGUF", "chat,edge", "1.6"},
		{"gemma2:9b", "bartowski/gemma-2-9b-it-GGUF", "chat", "5.6"},
		{"gemma2:27b", "bartowski/gemma-2-27b-it-GGUF", "chat,general", "17.0"},
		{"codegemma:2b", "bartowski/codegemma-2b-GGUF", "code,edge", "1.4"},
		{"codegemma:7b", "bartowski/codegemma-7b-it-GGUF", "code", "4.6"},
		// ── Phi ───────────────────────────────────────────────────────────────
		{"phi4:14b", "bartowski/phi-4-GGUF", "chat", "8.6"},
		{"phi4-mini:3.8b", "bartowski/Phi-4-mini-instruct-GGUF", "chat,edge", "2.5"},
		{"phi4-reasoning:14b", "bartowski/Phi-4-reasoning-GGUF", "reasoning", "8.6"},
		{"phi4-mini-reasoning:3.8b", "bartowski/Phi-4-mini-reasoning-GGUF", "reasoning,edge", "2.5"},
		{"phi4-multimodal:5.6b", "bartowski/Phi-4-multimodal-instruct-GGUF", "vision,audio", "3.5"},
		{"phi3.5:3.8b", "bartowski/Phi-3.5-mini-instruct-GGUF", "chat,edge", "2.4"},
		{"phi3:3.8b", "bartowski/Phi-3-mini-4k-instruct-GGUF", "chat,edge", "2.4"},
		{"phi3:14b", "bartowski/Phi-3-medium-4k-instruct-GGUF", "chat", "8.6"},
		{"phi3-vision:4.2b", "bartowski/Phi-3-vision-128k-instruct-GGUF", "vision,edge", "2.8"},
		// ── DeepSeek ──────────────────────────────────────────────────────────
		{"deepseek-r1:1.5b", "bartowski/DeepSeek-R1-Distill-Qwen-1.5B-GGUF", "reasoning,edge", "1.0"},
		{"deepseek-r1:7b", "bartowski/DeepSeek-R1-Distill-Qwen-7B-GGUF", "reasoning", "4.7"},
		{"deepseek-r1:8b", "bartowski/DeepSeek-R1-Distill-Llama-8B-GGUF", "reasoning", "5.0"},
		{"deepseek-r1:14b", "bartowski/DeepSeek-R1-Distill-Qwen-14B-GGUF", "reasoning", "9.0"},
		{"deepseek-r1:32b", "bartowski/DeepSeek-R1-Distill-Qwen-32B-GGUF", "reasoning", "20.0"},
		{"deepseek-r1:70b", "bartowski/DeepSeek-R1-Distill-Llama-70B-GGUF", "reasoning", "43.5"},
		{"deepseek-coder:1.3b", "TheBloke/deepseek-coder-1.3b-instruct-GGUF", "code,edge", "0.8"},
		{"deepseek-coder:6.7b", "TheBloke/deepseek-coder-6.7B-instruct-GGUF", "code", "4.1"},
		{"deepseek-coder:33b", "TheBloke/deepseek-coder-33B-instruct-GGUF", "code", "20.1"},
		{"deepseek-coder-v2:16b", "bartowski/DeepSeek-Coder-V2-Lite-Instruct-GGUF", "code", "9.7"},
		// ── Yi ────────────────────────────────────────────────────────────────
		{"yi:6b", "bartowski/Yi-6B-Chat-GGUF", "chat", "3.7"},
		{"yi:9b", "bartowski/Yi-1.5-9B-Chat-GGUF", "chat", "5.6"},
		{"yi:34b", "TheBloke/Yi-34B-Chat-GGUF", "chat,general", "20.7"},
		{"yi-coder:1.5b", "bartowski/Yi-Coder-1.5B-Chat-GGUF", "code,edge", "1.0"},
		{"yi-coder:9b", "bartowski/Yi-Coder-9B-Chat-GGUF", "code", "5.6"},
		{"yi-vl:6b", "bartowski/Yi-VL-6B-GGUF", "vision", "3.7"},
		{"yi-vl:34b", "bartowski/Yi-VL-34B-GGUF", "vision,general", "20.7"},
		// ── Falcon ────────────────────────────────────────────────────────────
		{"falcon:7b", "TheBloke/falcon-7b-instruct-GGUF", "chat", "4.3"},
		{"falcon2:11b", "bartowski/falcon-11b-GGUF", "chat", "6.8"},
		{"falcon3:1b", "bartowski/Falcon3-1B-Instruct-GGUF", "chat,edge", "0.7"},
		{"falcon3:3b", "bartowski/Falcon3-3B-Instruct-GGUF", "chat,edge", "2.0"},
		{"falcon3:7b", "bartowski/Falcon3-7B-Instruct-GGUF", "chat", "4.5"},
		{"falcon3:10b", "bartowski/Falcon3-10B-Instruct-GGUF", "chat", "6.2"},
		// ── StarCoder ─────────────────────────────────────────────────────────
		{"starcoder2:3b", "bartowski/starcoder2-3b-GGUF", "code,edge", "1.9"},
		{"starcoder2:7b", "bartowski/starcoder2-7b-GGUF", "code", "4.4"},
		{"starcoder2:15b", "bartowski/starcoder2-15b-GGUF", "code", "9.2"},
		// ── WizardLM ──────────────────────────────────────────────────────────
		{"wizardlm2:7b", "bartowski/WizardLM-2-7B-GGUF", "chat", "4.4"},
		{"wizardlm2:8x22b", "bartowski/WizardLM-2-8x22B-GGUF", "chat,general", "79.9"},
		{"wizard-coder:7b", "TheBloke/WizardCoder-Python-7B-V1.0-GGUF", "code", "4.1"},
		{"wizard-coder:13b", "TheBloke/WizardCoder-Python-13B-V1.0-GGUF", "code", "7.9"},
		{"wizard-coder:33b", "TheBloke/WizardCoder-33B-V1.1-GGUF", "code", "20.1"},
		{"wizard-math:7b", "TheBloke/WizardMath-7B-V1.1-GGUF", "math", "4.1"},
		{"wizard-math:70b", "TheBloke/WizardMath-70B-V1.0-GGUF", "math", "41.4"},
		// ── Hermes ────────────────────────────────────────────────────────────
		{"hermes3:8b", "bartowski/Hermes-3-Llama-3.1-8B-GGUF", "chat", "5.0"},
		{"hermes3:70b", "bartowski/Hermes-3-Llama-3.1-70B-GGUF", "chat,general", "43.0"},
		{"hermes3:405b", "bartowski/Hermes-3-Llama-3.1-405B-GGUF", "chat,general", "230.0"},
		{"openhermes2.5:7b", "TheBloke/OpenHermes-2.5-Mistral-7B-GGUF", "chat", "4.4"},
		{"nous-hermes2:11b", "bartowski/Nous-Hermes-2-Pro-Llama-3-8B-GGUF", "chat", "5.0"},
		{"nous-hermes2:34b", "TheBloke/Nous-Hermes-2-SOLAR-10.7B-GGUF", "chat", "6.5"},
		// ── Dolphin ───────────────────────────────────────────────────────────
		{"dolphin3:8b", "bartowski/dolphin3.0-llama3.1-8b-GGUF", "chat,roleplay", "5.0"},
		{"dolphin-mistral:7b", "TheBloke/dolphin-2.6-mistral-7B-GGUF", "chat,roleplay", "4.4"},
		{"dolphin-mixtral:8x7b", "bartowski/dolphin-2.7-mixtral-8x7b-GGUF", "chat,roleplay", "26.4"},
		{"dolphin-llama3:8b", "bartowski/dolphin-2.9.1-llama-3-8b-GGUF", "chat,roleplay", "5.0"},
		{"dolphin-llama3:70b", "bartowski/dolphin-2.9-llama3-70b-GGUF", "chat,roleplay", "43.0"},
		// ── Orca ──────────────────────────────────────────────────────────────
		{"orca-mini:3b", "TheBloke/orca_mini_3B-GGUF", "chat,edge", "1.9"},
		{"orca-mini:7b", "TheBloke/orca_mini_7B-GGUF", "chat", "4.1"},
		{"orca-mini:13b", "TheBloke/orca_mini_13b-GGUF", "chat", "7.9"},
		{"orca-mini:70b", "TheBloke/orca_mini_70b-GGUF", "chat,general", "41.4"},
		{"orca2:7b", "TheBloke/Orca-2-7B-GGUF", "chat,reasoning", "4.1"},
		{"orca2:13b", "TheBloke/Orca-2-13B-GGUF", "chat,reasoning", "7.9"},
		// ── Vicuna ────────────────────────────────────────────────────────────
		{"vicuna:7b", "TheBloke/vicuna-7B-v1.5-GGUF", "chat", "4.1"},
		{"vicuna:13b", "TheBloke/vicuna-13B-v1.5-GGUF", "chat", "7.9"},
		{"vicuna:33b", "TheBloke/vicuna-33B-GGUF", "chat,general", "20.1"},
		// ── OpenChat ──────────────────────────────────────────────────────────
		{"openchat:7b", "bartowski/openchat-3.6-8b-20240522-GGUF", "chat", "5.0"},
		{"openhermes:7b", "TheBloke/OpenHermes-2-Mistral-7B-GGUF", "chat", "4.4"},
		// ── Zephyr ────────────────────────────────────────────────────────────
		{"zephyr:7b", "TheBloke/zephyr-7B-beta-GGUF", "chat", "4.4"},
		{"zephyr:141b", "bartowski/Zephyr-141B-A39B-GGUF", "chat,general", "85.2"},
		{"stablelm-zephyr:3b", "bartowski/stablelm-zephyr-3b-GGUF", "chat,edge", "1.9"},
		// ── Solar ─────────────────────────────────────────────────────────────
		{"solar:10.7b", "bartowski/SOLAR-10.7B-Instruct-v1.0-GGUF", "chat", "6.5"},
		{"solar-pro:22b", "bartowski/solar-pro-preview-instruct-GGUF", "chat,general", "13.5"},
		// ── Command R ─────────────────────────────────────────────────────────
		{"command-r:35b", "bartowski/c4ai-command-r-v01-GGUF", "chat,tools", "21.4"},
		{"command-r-plus:104b", "bartowski/c4ai-command-r-plus-GGUF", "chat,tools", "63.7"},
		{"command-r7b:7b", "bartowski/c4ai-command-r7b-12-2024-GGUF", "chat,edge", "4.4"},
		// ── Stable LM ─────────────────────────────────────────────────────────
		{"stablelm2:1.6b", "bartowski/stablelm-2-zephyr-1_6b-GGUF", "chat,edge", "1.0"},
		{"stablelm2:12b", "bartowski/stablelm-2-12b-chat-GGUF", "chat", "7.5"},
		{"stable-beluga:7b", "TheBloke/StableBeluga-7B-GGUF", "chat", "4.1"},
		{"stable-beluga:13b", "TheBloke/StableBeluga-13B-GGUF", "chat", "7.9"},
		{"stable-beluga:70b", "TheBloke/StableBeluga-70B-GGUF", "chat", "41.4"},
		// ── Nemotron ──────────────────────────────────────────────────────────
		{"nemotron:70b", "bartowski/Llama-3.1-Nemotron-70B-Instruct-HF-GGUF", "chat,general", "43.0"},
		{"nemotron-mini:4b", "bartowski/Nemotron-Mini-4B-Instruct-GGUF", "chat,edge", "2.5"},
		{"nemotron-super:49b", "bartowski/Llama-3.3-Nemotron-Super-49B-v1-GGUF", "chat,general", "30.3"},
		// ── Embedding ─────────────────────────────────────────────────────────
		{"nomic-embed-text", "nomic-ai/nomic-embed-text-v1.5-GGUF", "embed", "0.1"},
		{"mxbai-embed-large", "mixedbread-ai/mxbai-embed-large-v1-GGUF", "embed", "0.2"},
		{"bge-m3", "gpustack/bge-m3-GGUF", "embed,multilingual", "0.3"},
		{"bge-large", "gpustack/bge-large-en-v1.5-GGUF", "embed", "0.2"},
		{"snowflake-arctic-embed:335m", "gpustack/snowflake-arctic-embed-m-v2.0-GGUF", "embed", "0.2"},
		{"all-minilm:22m", "second-state/All-MiniLM-L6-v2-Embedding-GGUF", "embed", "0.05"},
		// ── Tiny / Edge ───────────────────────────────────────────────────────
		{"tinyllama:1.1b", "TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF", "chat,edge", "0.7"},
		{"smollm2:135m", "bartowski/SmolLM2-135M-Instruct-GGUF", "chat,edge", "0.1"},
		{"smollm2:360m", "bartowski/SmolLM2-360M-Instruct-GGUF", "chat,edge", "0.2"},
		{"smollm2:1.7b", "bartowski/SmolLM2-1.7B-Instruct-GGUF", "chat,edge", "1.1"},
		{"minicpm3:4b", "bartowski/MiniCPM3-4B-GGUF", "chat,edge", "2.5"},
		{"megrez:3b", "bartowski/Megrez-3B-Instruct-GGUF", "chat,edge", "2.0"},
		{"mobilellm:125m", "bartowski/MobileLLM-125M-GGUF", "edge,mobile", "0.1"},
		{"mobilellm:350m", "bartowski/MobileLLM-350M-GGUF", "edge,mobile", "0.2"},
		// ── Vision ────────────────────────────────────────────────────────────
		{"llava:7b", "mys/ggml_llava-v1.5-7b", "vision,chat", "4.1"},
		{"llava:13b", "mys/ggml_llava-v1.5-13b", "vision,chat", "7.9"},
		{"llava:34b", "cjpais/llava-v1.6-34b-gguf", "vision,chat", "20.7"},
		{"llava-phi3:3.8b", "xtuner/llava-phi-3-mini-gguf", "vision,edge", "2.4"},
		{"llava-llama3:8b", "xtuner/llava-llama-3-8b-v1_1-gguf", "vision", "5.0"},
		{"bakllava:7b", "mys/ggml_bakllava-1", "vision,chat", "4.1"},
		{"moondream2:1.8b", "vikhyatk/moondream2", "vision,edge", "1.1"},
		{"moondream:1.8b", "vikhyatk/moondream2", "vision,edge", "1.1"},
		{"minicpm-v:8b", "bartowski/MiniCPM-V-2_6-GGUF", "vision", "5.0"},
		{"internvl2:1b", "bartowski/InternVL2-1B-GGUF", "vision,edge", "0.9"},
		{"internvl2:2b", "bartowski/InternVL2-2B-GGUF", "vision,edge", "1.6"},
		{"internvl2:8b", "bartowski/InternVL2-8B-GGUF", "vision,chat", "5.0"},
		{"internvl2:26b", "bartowski/InternVL2-26B-GGUF", "vision,chat", "16.0"},
		{"smolvlm:256m", "ggml-org/SmolVLM-256M-Instruct-GGUF", "vision,edge", "0.2"},
		{"smolvlm:500m", "ggml-org/SmolVLM-500M-Instruct-GGUF", "vision,edge", "0.3"},
		{"kimi-vl:16b", "bartowski/Kimi-VL-A3B-Instruct-GGUF", "vision,chat", "9.9"},
		// ── Chinese / Bilingual ───────────────────────────────────────────────
		{"chatglm3:6b", "THUDM/chatglm3-6b-gguf", "chat,chinese", "3.7"},
		{"glm4:9b", "bartowski/glm-4-9b-chat-GGUF", "chat,chinese", "5.6"},
		{"baichuan2:7b", "TheBloke/Baichuan2-7B-Chat-GGUF", "chat,chinese", "4.1"},
		{"baichuan2:13b", "TheBloke/Baichuan2-13B-Chat-GGUF", "chat,chinese", "7.9"},
		{"xverse:7b", "xverse/XVERSE-7B-Chat-GGUF", "chat,chinese", "4.3"},
		// ── Arabic ────────────────────────────────────────────────────────────
		{"jais:13b", "inceptionai/jais-13b-chat-GGUF", "chat,arabic", "7.9"},
		{"jais:30b", "inceptionai/jais-30b-chat-GGUF", "chat,arabic", "18.3"},
		{"layla:8b", "bartowski/Llama-3.1-8B-Lexi-Uncensored-V2-GGUF", "chat,arabic", "5.0"},
		// ── Multilingual ──────────────────────────────────────────────────────
		{"aya-expanse:8b", "bartowski/aya-expanse-8b-GGUF", "chat,multilingual", "5.0"},
		{"aya-expanse:32b", "bartowski/aya-expanse-32b-GGUF", "chat,multilingual", "19.8"},
		{"tiny-aya-global:3b", "bartowski/tiny-aya-global-GGUF", "chat,multilingual", "2.1"},
		{"openbuddy:70b", "TheBloke/OpenBuddy-LLaMA2-70B-v13.2-GGUF", "chat,multilingual", "41.4"},
		{"phoenix:7b", "TheBloke/phoenix-7b-GGUF", "chat,multilingual", "4.1"},
		{"sailor2:8b", "bartowski/Sailor2-8B-Chat-GGUF", "chat,multilingual", "5.0"},
		{"sailor2:20b", "bartowski/Sailor2-20B-Chat-GGUF", "chat,multilingual", "12.2"},
		// ── InternLM ──────────────────────────────────────────────────────────
		{"internlm2.5:7b", "bartowski/internlm2_5-7b-chat-GGUF", "chat", "4.4"},
		{"internlm2.5:20b", "bartowski/internlm2_5-20b-chat-GGUF", "chat", "12.2"},
		{"internlm3:8b", "bartowski/internlm3-8b-instruct-GGUF", "chat", "5.0"},
		// ── Granite ───────────────────────────────────────────────────────────
		{"granite-code:3b", "bartowski/granite-3b-code-instruct-128k-GGUF", "code,edge", "1.9"},
		{"granite-code:8b", "bartowski/granite-8b-code-instruct-128k-GGUF", "code", "5.0"},
		{"granite-code:20b", "bartowski/granite-20b-code-instruct-8k-GGUF", "code", "12.2"},
		{"granite-code:34b", "bartowski/granite-34b-code-instruct-8k-GGUF", "code", "20.7"},
		{"granite3.1-dense:2b", "bartowski/granite-3.1-2b-instruct-GGUF", "chat,edge", "1.3"},
		{"granite3.1-dense:8b", "bartowski/granite-3.1-8b-instruct-GGUF", "chat", "5.0"},
		{"granite3.2-vision:2b", "bartowski/granite-3.2-2b-instruct-GGUF", "vision,edge", "1.3"},
		// ── Medical ───────────────────────────────────────────────────────────
		{"meditron:7b", "TheBloke/meditron-7B-GGUF", "medical", "4.1"},
		{"meditron:70b", "TheBloke/meditron-70B-GGUF", "medical", "41.4"},
		{"medalpaca:7b", "TheBloke/medalpaca-7B-GGUF", "medical", "4.1"},
		{"clinical-camel:70b", "TheBloke/ClinicalCamel-70B-GGUF", "medical", "41.4"},
		// ── Math ──────────────────────────────────────────────────────────────
		{"deepseek-math:7b", "TheBloke/deepseek-math-7b-instruct-GGUF", "math", "4.1"},
		{"numina-math:7b", "bartowski/NuminaMath-7B-CoT-GGUF", "math", "4.4"},
		// ── Reasoning ─────────────────────────────────────────────────────────
		{"reflection:70b", "bartowski/Reflection-Llama-3.1-70B-GGUF", "reasoning", "43.0"},
		{"skywork-o1:8b", "bartowski/Skywork-o1-Open-Llama-3.1-8B-GGUF", "reasoning", "5.0"},
		{"skywork-or1:32b", "bartowski/Skywork-OR1-32B-Preview-GGUF", "reasoning", "20.0"},
		{"cogito:8b", "bartowski/cogito-v1-preview-llama-3B-GGUF", "reasoning,chat", "5.0"},
		{"cogito:14b", "bartowski/cogito-v1-preview-qwen-14b-GGUF", "reasoning,chat", "9.0"},
		{"cogito:32b", "bartowski/cogito-v1-preview-qwen-32b-GGUF", "reasoning,chat", "20.0"},
		{"openthinker:7b", "bartowski/OpenThinker-7B-GGUF", "reasoning", "4.4"},
		{"openthinker:32b", "bartowski/OpenThinker-32B-GGUF", "reasoning", "20.0"},
		{"tulu3:8b", "bartowski/Llama-3.1-Tulu-3-8B-GGUF", "chat,reasoning", "5.0"},
		{"tulu3:70b", "bartowski/Llama-3.1-Tulu-3-70B-GGUF", "chat,reasoning", "43.0"},
		// ── Tool Use / Agents ─────────────────────────────────────────────────
		{"nexusraven:13b", "bartowski/NexusRaven-V2-13B-GGUF", "tools", "7.9"},
		{"gorilla-openfunctions:7b", "TheBloke/gorilla-openfunctions-v2-GGUF", "tools", "4.1"},
		{"xlam:8b", "bartowski/xLAM-8x7b-r-GGUF", "tools,agent", "26.4"},
		{"firefunction-v2:70b", "bartowski/FireFunction-v2-GGUF", "tools,chat", "43.0"},
		{"llama3-groq-tool-use:8b", "bartowski/Llama-3-Groq-8B-Tool-Use-GGUF", "tools", "5.0"},
		{"llama3-groq-tool-use:70b", "bartowski/Llama-3-Groq-70B-Tool-Use-GGUF", "tools", "43.0"},
		{"devstral:24b", "bartowski/Devstral-Small-2505-GGUF", "code,agent", "14.8"},
		// ── Roleplay ──────────────────────────────────────────────────────────
		{"samantha-mistral:7b", "TheBloke/Samantha-Mistral-7B-GGUF", "roleplay,chat", "4.4"},
		{"airoboros:7b", "TheBloke/Airoboros-L2-7B-2.2-GGUF", "chat,roleplay", "4.1"},
		{"airoboros:70b", "TheBloke/Airoboros-L2-70b-GPT4-2.0-GGUF", "chat,roleplay", "41.4"},
		{"goliath:120b", "bartowski/goliath-120b-GGUF", "chat", "73.4"},
		{"bagel:34b", "bartowski/bagel-dpo-34b-v0.2-GGUF", "chat", "20.7"},
		{"toppy-m:7b", "TheBloke/Toppy-M-7B-GGUF", "chat,roleplay", "4.4"},
		// ── SQL ───────────────────────────────────────────────────────────────
		{"sqlcoder:7b", "bartowski/sqlcoder-7b-2-GGUF", "code,sql", "4.4"},
		{"sqlcoder:15b", "bartowski/sqlcoder-15b-GGUF", "code,sql", "9.2"},
		// ── Code (More) ───────────────────────────────────────────────────────
		{"phind-codellama:34b", "TheBloke/Phind-CodeLlama-34B-v2-GGUF", "code", "20.7"},
		{"magicoder:7b", "TheBloke/Magicoder-S-DS-6.7B-GGUF", "code", "4.1"},
		{"stable-code:3b", "bartowski/stable-code-instruct-3b-GGUF", "code,edge", "1.9"},
		{"codeup:13b", "TheBloke/CodeUp-Alpha-13B-HF-GGUF", "code", "7.9"},
		{"santacoder:1.1b", "bartowski/starcoder-GGUF", "code,edge", "0.7"},
		{"replit-code:3b", "TheBloke/replit-code-v1_5-3b-GGUF", "code,edge", "1.9"},
		// ── Platypus / STEM ───────────────────────────────────────────────────
		{"platypus2:7b", "TheBloke/Platypus2-7B-GGUF", "chat,reasoning", "4.1"},
		{"platypus2:13b", "TheBloke/Platypus2-13B-GGUF", "chat,reasoning", "7.9"},
		{"platypus2:70b", "TheBloke/Platypus2-70B-instruct-GGUF", "chat,reasoning", "41.4"},
		// ── OLMo ──────────────────────────────────────────────────────────────
		{"olmo2:7b", "bartowski/OLMo-2-1124-7B-Instruct-GGUF", "chat,research", "4.4"},
		{"olmo2:13b", "bartowski/OLMo-2-1124-13B-Instruct-GGUF", "chat,research", "7.9"},
		// ── Korean ────────────────────────────────────────────────────────────
		{"exaone3.5:2.4b", "bartowski/EXAONE-3.5-2.4B-Instruct-GGUF", "chat,edge,korean", "1.5"},
		{"exaone3.5:7.8b", "bartowski/EXAONE-3.5-7.8B-Instruct-GGUF", "chat,korean", "4.9"},
		{"exaone3.5:32b", "bartowski/EXAONE-3.5-32B-Instruct-GGUF", "chat,general,korean", "19.8"},
		{"exaone4:32b", "bartowski/EXAONE-4.0-32B-GGUF", "chat", "20.4"},
		{"exaone-deep:32b", "bartowski/EXAONE-Deep-32B-GGUF", "chat", "20.4"},
		{"hyperclova:1.5b", "bartowski/HyperCLOVAX-SEED-Text-Instruct-1.5B-GGUF", "chat,korean", "1.0"},
		// ── Nous Capybara ─────────────────────────────────────────────────────
		{"nous-capybara:7b", "TheBloke/Nous-Capybara-7B-V1.9-GGUF", "chat", "4.1"},
		{"nous-capybara:34b", "TheBloke/Nous-Capybara-34B-GGUF", "chat,general", "20.7"},
		// ── Tools / Extraction ────────────────────────────────────────────────
		{"nuextract:3.8b", "bartowski/NuExtract-v1.5-GGUF", "tools,extract", "2.4"},
		{"reader-lm:0.5b", "bartowski/reader-lm-0.5b-GGUF", "tools,edge", "0.4"},
		{"reader-lm:1.5b", "bartowski/reader-lm-1.5b-GGUF", "tools", "1.0"},
		{"bespoke-minicheck:7b", "bartowski/Bespoke-MiniCheck-7B-GGUF", "tools", "4.4"},
		// ── Athene ────────────────────────────────────────────────────────────
		{"athene-v2:72b", "bartowski/Athene-V2-Chat-GGUF", "chat,tools", "44.5"},
		// ── Hunyuan ───────────────────────────────────────────────────────────
		{"hunyuan-a13b", "bartowski/Hunyuan-A13B-Instruct-GGUF", "chat,general", "8.0"},
		// ── GPT-J / Old Classic ───────────────────────────────────────────────
		{"gpt-j:6b", "nomic-ai/gpt4all-j-groovy-GGUF", "chat", "3.8"},
		{"gpt4all-falcon:7b", "nomic-ai/gpt4all-falcon-newbpe-q4_0", "chat,edge", "4.3"},
		// ── Pythia ────────────────────────────────────────────────────────────
		{"pythia:1b", "bartowski/pythia-1b-GGUF", "research", "0.6"},
		{"pythia:2.8b", "bartowski/pythia-2.8b-GGUF", "research", "1.7"},
		{"pythia:6.9b", "bartowski/pythia-6.9b-GGUF", "research", "4.2"},
		// ── Mamba / SSM ───────────────────────────────────────────────────────
		{"mamba:2.8b", "bartowski/mamba-2.8b-hf-GGUF", "chat,ssm", "1.7"},
		{"jamba1.5:mini:12b", "bartowski/Jamba-1.5-Mini-GGUF", "chat,ssm", "7.4"},
		{"jamba1.5:large:398b", "bartowski/Jamba-1.5-Large-GGUF", "chat,ssm,general", "244.7"},
		// ── Everythinglm / Long context ───────────────────────────────────────
		{"everythinglm:13b", "TheBloke/EverythingLM-13B-V2-GGUF", "chat,long-context", "7.9"},
		{"yarn-mistral:7b", "TheBloke/Yarn-Mistral-7B-128k-GGUF", "chat,long-context", "4.4"},
		{"yarn-llama2:7b", "TheBloke/Yarn-Llama-2-7B-128k-GGUF", "chat,long-context", "4.1"},
		{"mistrallite:7b", "TheBloke/MistralLite-7B-GGUF", "chat,long-context", "4.4"},
		// ── Xwin ──────────────────────────────────────────────────────────────
		{"xwinlm:7b", "TheBloke/Xwin-LM-7B-V0.2-GGUF", "chat", "4.1"},
		{"xwinlm:13b", "TheBloke/Xwin-LM-13B-V0.2-GGUF", "chat", "7.9"},
		{"xwinlm:70b", "TheBloke/Xwin-LM-70B-V0.1-GGUF", "chat,general", "41.4"},
		// ── Notux / MoE ───────────────────────────────────────────────────────
		{"notux:8x7b", "bartowski/notux-8x7b-v1-GGUF", "chat,general", "26.4"},
		// ── Long context LLaMA 3 ──────────────────────────────────────────────
		{"llama3-gradient:8b", "bartowski/Llama-3-8B-Gradient-1048k-GGUF", "chat,long-context", "5.0"},
		{"llama3-gradient:70b", "bartowski/Llama-3-70B-Gradient-1048k-GGUF", "chat,long-context", "43.0"},
		// ── Incite ────────────────────────────────────────────────────────────
		{"incite-chat:3b", "TheBloke/RedPajama-INCITE-Chat-3B-v1-GGUF", "chat,edge", "1.9"},
		{"incite-chat:7b", "TheBloke/RedPajama-INCITE-Chat-7B-v0.1-GGUF", "chat", "4.3"},
		// ── OpenBuddy ─────────────────────────────────────────────────────────
		{"openbuddy:13b", "TheBloke/OpenBuddy-LLaMA-13B-v11.1-GGUF", "chat,multilingual", "7.9"},
		{"openbuddy:34b", "TheBloke/OpenBuddy-Orca-LLaMA2-34B-v11.1-GGUF", "chat,multilingual", "20.7"},
		// ── Amber / Crystal (Open LLM360) ─────────────────────────────────────
		{"amber:7b", "LLM360/Amber-GGUF", "research", "4.1"},
		// ── Speechless ────────────────────────────────────────────────────────
		{"speechless-llama2:13b", "TheBloke/Speechless-Llama2-13B-GGUF", "chat", "7.9"},
		{"speechless-code-mistral:7b", "TheBloke/Speechless-Code-Mistral-7B-v1.0-GGUF", "code", "4.4"},
		// ── Camel ─────────────────────────────────────────────────────────────
		{"camel:7b", "TheBloke/camel-5b-hermes-slerp-GGUF", "chat,agent", "3.1"},
		// ── Nous Yi ───────────────────────────────────────────────────────────
		{"nous-yarn-llama2:7b", "TheBloke/Nous-Yarn-Llama-2-7B-128k-GGUF", "chat,long-context", "4.1"},
		{"nous-yarn-llama2:13b", "TheBloke/Nous-Yarn-Llama-2-13B-128k-GGUF", "chat,long-context", "7.9"},

		// ── DeepSeek V2 / V3 / V3.5 ──────────────────────────────────────────
		{"deepseek-v2:16b", "bartowski/DeepSeek-V2-Lite-Chat-GGUF", "chat,moe", "9.7"},
		{"deepseek-v2:236b", "bartowski/DeepSeek-V2-Chat-GGUF", "chat,moe,general", "144.7"},
		{"deepseek-v3:671b", "bartowski/DeepSeek-V3-GGUF", "chat,moe,general", "410.0"},
		{"deepseek-v3-0324:671b", "bartowski/DeepSeek-V3-0324-GGUF", "chat,moe,general", "410.0"},
		{"deepseek-r1:671b", "bartowski/DeepSeek-R1-GGUF", "reasoning,moe,general", "410.0"},
		{"deepseek-r1-0528:671b", "bartowski/DeepSeek-R1-0528-GGUF", "reasoning,moe,general", "410.0"},
		{"deepseek-prover:7b", "bartowski/DeepSeek-Prover-V1.5-Instruct-GGUF", "math,reasoning", "4.4"},
		{"deepseek-prover:671b", "bartowski/DeepSeek-Prover-V2-GGUF", "math,reasoning,moe", "410.0"},
		{"deepseek-r1t:671b", "bartowski/DeepSeek-R1T-Chimera-GGUF", "reasoning,moe", "410.0"},
		{"janus-pro:7b", "bartowski/Janus-Pro-7B-GGUF", "vision,chat", "4.4"},

		// ── Qwen3 MoE ─────────────────────────────────────────────────────────
		{"qwen3:30b-a3b", "bartowski/Qwen3-30B-A3B-GGUF", "chat,reasoning,moe,edge", "18.6"},
		{"qwen3:235b-a22b", "bartowski/Qwen3-235B-A22B-GGUF", "chat,reasoning,moe,general", "144.0"},
		{"qwen3-coder:8b", "bartowski/Qwen3-8B-Coder-GGUF", "code", "5.2"},
		{"qwen3-coder:32b", "bartowski/Qwen3-32B-Coder-GGUF", "code", "20.4"},
		{"qwen3-embedding:0.6b", "Qwen/Qwen3-Embedding-0.6B-GGUF", "embed,edge", "0.4"},
		{"qwen3-embedding:4b", "Qwen/Qwen3-Embedding-4B-GGUF", "embed", "2.6"},
		{"qwen3-embedding:8b", "Qwen/Qwen3-Embedding-8B-GGUF", "embed", "5.2"},
		{"qwq:72b", "bartowski/QwQ-72B-GGUF", "reasoning,general", "44.5"},
		{"qvq:32b", "bartowski/QvQ-32B-Preview-GGUF", "vision,reasoning", "20.4"},
		{"qwen2.5-vl:3b", "bartowski/Qwen2.5-VL-3B-Instruct-GGUF", "vision,edge", "2.0"},
		{"qwen2.5-vl:7b", "bartowski/Qwen2.5-VL-7B-Instruct-GGUF", "vision,chat", "4.7"},
		{"qwen2.5-vl:32b", "bartowski/Qwen2.5-VL-32B-Instruct-GGUF", "vision,general", "19.8"},
		{"qwen2.5-vl:72b", "bartowski/Qwen2.5-VL-72B-Instruct-GGUF", "vision,general", "44.5"},
		{"qwen2-vl:7b", "bartowski/Qwen2-VL-7B-Instruct-GGUF", "vision,chat", "4.7"},
		{"qwen2-vl:72b", "bartowski/Qwen2-VL-72B-Instruct-GGUF", "vision,general", "44.5"},
		{"qwen2-audio:7b", "bartowski/Qwen2-Audio-7B-Instruct-GGUF", "audio,chat", "4.7"},

		// ── Mistral / Mixtral (more) ──────────────────────────────────────────
		{"mistral:7b-v0.2", "bartowski/Mistral-7B-Instruct-v0.2-GGUF", "chat", "4.4"},
		{"mistral:7b-v0.1", "TheBloke/Mistral-7B-Instruct-v0.1-GGUF", "chat", "4.4"},
		{"mistral-small3:24b", "bartowski/Mistral-Small-3-24B-Instruct-2503-GGUF", "chat", "14.8"},
		{"mistral-medium3:123b", "bartowski/Mistral-Medium-3-GGUF", "chat,general", "75.5"},
		{"ministral:3b", "bartowski/Ministral-3B-Instruct-GGUF", "chat,edge", "2.0"},
		{"ministral:8b", "bartowski/Ministral-8B-Instruct-2410-GGUF", "chat", "5.0"},
		{"mixtral-nemo:12b", "bartowski/Mistral-Nemo-Instruct-2407-GGUF", "chat", "7.3"},
		{"mathstral:7b", "bartowski/mathstral-7B-v0.1-GGUF", "math", "4.4"},
		{"pixtral:12b", "bartowski/Pixtral-12B-2409-GGUF", "vision,chat", "7.5"},
		{"pixtral-large:124b", "bartowski/Pixtral-Large-Instruct-2411-GGUF", "vision,general", "75.5"},
		{"mistral-small3.1:24b-vision", "bartowski/Mistral-Small-3.1-24B-Instruct-2503-GGUF", "vision,chat", "14.8"},

		// ── Gemma 3n / SmolLM3 (very new) ────────────────────────────────────
		{"gemma3n:e2b", "bartowski/gemma-3n-E2B-it-GGUF", "chat,multimodal,edge", "2.0"},
		{"gemma3n:e4b", "bartowski/gemma-3n-E4B-it-GGUF", "chat,multimodal,edge", "4.0"},
		{"smollm3:3b", "bartowski/SmolLM3-3B-GGUF", "chat,edge", "2.0"},

		// ── Phi 4 extra sizes ─────────────────────────────────────────────────
		{"phi4:3.8b", "bartowski/Phi-4-mini-instruct-GGUF", "chat,edge", "2.5"},
		{"phi3.5-moe:41b", "bartowski/Phi-3.5-MoE-instruct-GGUF", "chat,moe", "25.3"},
		{"phi2:2.7b", "bartowski/phi-2-GGUF", "chat,edge", "1.7"},
		{"phi1.5:1.3b", "bartowski/phi-1_5-GGUF", "chat,edge", "0.8"},

		// ── Llama 3.x (more variants) ─────────────────────────────────────────
		{"llama3.1:8b-text", "bartowski/Meta-Llama-3.1-8B-GGUF", "chat", "4.9"},
		{"llama3.1:70b-text", "bartowski/Meta-Llama-3.1-70B-GGUF", "chat,general", "42.5"},
		{"llama-guard3:8b", "bartowski/Llama-Guard-3-8B-GGUF", "safety", "5.0"},
		{"llama-guard3:11b-vision", "bartowski/Llama-Guard-3-11B-Vision-GGUF", "safety,vision", "6.8"},
		{"meta-llama3.3:70b-nemo", "bartowski/Llama-3.3-70B-Instruct-evals-GGUF", "chat,general", "43.0"},
		{"llama3.1-storm:8b", "bartowski/Llama-3.1-Storm-8B-GGUF", "chat,roleplay", "5.0"},
		{"llama3.1-supernova:70b", "bartowski/Llama-3.1-Supernova-70B-GGUF", "chat,general", "43.0"},
		{"llama3.1-intuition:8b", "bartowski/Llama-3.1-8B-Instruct-abliterated-GGUF", "chat", "5.0"},
		{"magnum:72b", "bartowski/Magnum-72B-v4-GGUF", "chat,roleplay", "44.5"},
		{"magnum:v4:7b", "bartowski/Magnum-v4-7B-GGUF", "chat,roleplay", "4.4"},
		{"calme3.3:8b", "bartowski/calme-3.3-llamaloi-8b-GGUF", "chat,tools", "5.0"},
		{"calme3:78b", "bartowski/calme-3.1-qwen2.5-78b-GGUF", "chat,tools,general", "44.5"},
		{"skyfall:7b", "bartowski/Skyfall-36v-v2-GGUF", "chat,roleplay", "4.4"},
		{"llama3.2:3b-tools", "bartowski/Llama-3.2-3B-Instruct-uncensored-GGUF", "chat,tools,edge", "2.0"},

		// ── Gemma 3 (more) ────────────────────────────────────────────────────
		{"gemma3:4b-vision", "bartowski/gemma-3-4b-it-GGUF", "vision,edge", "2.5"},
		{"gemma3:27b-vision", "bartowski/gemma-3-27b-it-GGUF", "vision,general", "17.0"},
		{"paligemma2:3b", "bartowski/paligemma2-3b-ft-docci-448-GGUF", "vision,edge", "2.0"},
		{"paligemma2:10b", "bartowski/paligemma2-10b-ft-docci-448-GGUF", "vision", "6.2"},
		{"paligemma2:28b", "bartowski/paligemma2-28b-ft-docci-448-GGUF", "vision,general", "17.0"},

		// ── GLM / ChatGLM (more) ──────────────────────────────────────────────
		{"glm4-9b-long:9b", "bartowski/glm-4-9b-chat-1m-GGUF", "chat,chinese,long-context", "5.6"},
		{"glm4v:9b", "bartowski/GLM-4V-9B-GGUF", "vision,chinese", "5.6"},
		{"glm-z1:9b", "bartowski/GLM-Z1-9B-0414-GGUF", "reasoning,chinese", "5.6"},
		{"glm-z1:32b", "bartowski/GLM-Z1-32B-0414-GGUF", "reasoning,chinese,general", "20.0"},
		{"glm4-flash:9b", "bartowski/glm-4-9b-chat-GGUF", "chat,chinese", "5.6"},
		{"codegeex4:9b", "bartowski/codegeex4-all-9b-GGUF", "code,chinese", "5.6"},

		// ── Nemotron (more) ───────────────────────────────────────────────────
		{"nemotron:340b", "bartowski/Nemotron-340B-Instruct-GGUF", "chat,general", "207.6"},
		{"llama-3.1-nemotron-nano:8b", "bartowski/Llama-3.1-Nemotron-Nano-8B-v1-GGUF", "chat,edge", "5.0"},
		{"minitron:4b", "bartowski/Minitron-4B-Base-GGUF", "chat,edge", "2.5"},

		// ── InternLM (more) ───────────────────────────────────────────────────
		{"internlm2:1.8b", "bartowski/internlm2-chat-1_8b-GGUF", "chat,edge", "1.1"},
		{"internlm2.5:1.8b", "bartowski/internlm2_5-1_8b-chat-GGUF", "chat,edge", "1.1"},
		{"internlm-xcomposer2.5:7b", "bartowski/InternLM-XComposer2d5-7b-GGUF", "vision,chinese", "4.4"},
		{"internvl2.5:4b", "bartowski/InternVL2_5-4B-GGUF", "vision,edge", "2.5"},
		{"internvl2.5:8b", "bartowski/InternVL2_5-8B-GGUF", "vision,chat", "5.0"},
		{"internvl2.5:38b", "bartowski/InternVL2_5-38B-GGUF", "vision,general", "23.4"},
		{"internvl2.5:78b", "bartowski/InternVL2_5-78B-GGUF", "vision,general", "48.2"},

		// ── OLMo (more) ───────────────────────────────────────────────────────
		{"olmo:7b", "bartowski/OLMo-7B-0424-Instruct-GGUF", "chat,research", "4.4"},
		{"olmo-bitnet:7b", "bartowski/OLMo-Bitnet-1B-hf-GGUF", "research,edge", "0.6"},
		{"olmoe:1b-7b", "bartowski/OLMoE-1B-7B-0924-Instruct-GGUF", "chat,moe,edge", "4.4"},

		// ── Bloom / BigScience ────────────────────────────────────────────────
		{"bloomz:7b", "TheBloke/bloomz-7b1-GGUF", "chat,multilingual", "4.3"},

		// ── Granite (more) ────────────────────────────────────────────────────
		{"granite3.3:2b", "bartowski/granite-3.3-2b-instruct-GGUF", "chat,tools,edge", "1.3"},
		{"granite3.3:8b", "bartowski/granite-3.3-8b-instruct-GGUF", "chat,tools", "5.0"},
		{"granite3.2:8b", "bartowski/granite-3.2-8b-instruct-GGUF", "chat", "5.0"},
		{"granite3.1-moe:1b", "bartowski/granite-3.1-1b-a400m-instruct-GGUF", "chat,moe,edge", "0.8"},
		{"granite3.1-moe:3b", "bartowski/granite-3.1-3b-a800m-instruct-GGUF", "chat,moe,edge", "1.9"},

		// ── Microsoft (more) ──────────────────────────────────────────────────
		{"phi-silica:3.3b", "bartowski/Phi-Silica-GGUF", "chat,edge", "2.0"},
		{"bitnet-b1.58:3b", "bartowski/BitNet-b1.58-2B-4T-GGUF", "chat,edge", "0.4"},
		{"wizard-vicuna-uncensored:7b", "TheBloke/Wizard-Vicuna-7B-Uncensored-GGUF", "chat", "4.1"},
		{"wizard-vicuna-uncensored:13b", "TheBloke/Wizard-Vicuna-13B-Uncensored-GGUF", "chat", "7.9"},
		{"wizard-vicuna-uncensored:30b", "TheBloke/Wizard-Vicuna-30B-Uncensored-GGUF", "chat,general", "18.3"},

		// ── Cohere / Command (more) ───────────────────────────────────────────
		{"command-r:08-2024", "bartowski/c4ai-command-r-08-2024-GGUF", "chat,tools", "21.4"},
		{"command-r-plus:08-2024", "bartowski/c4ai-command-r-plus-08-2024-GGUF", "chat,tools,general", "63.7"},
		{"aya23:8b", "bartowski/aya-23-8B-GGUF", "chat,multilingual", "5.0"},
		{"aya23:35b", "bartowski/aya-23-35B-GGUF", "chat,multilingual,general", "21.4"},

		// ── xAI Grok (open weights) ───────────────────────────────────────────
		{"grok-2-27b", "bartowski/grok-2-27b-GGUF", "chat,general", "16.6"},

		// ── 01.AI / Skywork (more) ────────────────────────────────────────────
		{"skywork:9b", "bartowski/Skywork-9B-Chat-GGUF", "chat", "5.5"},
		{"skywork-reward:8b", "bartowski/Skywork-Reward-Llama-3.1-8B-GGUF", "tools", "5.0"},

		// ── Conan / Korean specialized ────────────────────────────────────────
		{"conan-v2:7b", "bartowski/Conan-v2-GGUF", "chat,korean", "4.4"},
		{"kanana:2.1b", "bartowski/kanana-nano-2.1b-instruct-GGUF", "chat,korean,edge", "1.3"},

		// ── Arabic / Jais (more) ──────────────────────────────────────────────
		{"jais:6.7b", "inceptionai/jais-adapted-7b-chat-GGUF", "chat,arabic", "4.1"},

		// ── Falcon (more) ─────────────────────────────────────────────────────
		{"falcon3:mamba:7b", "bartowski/Falcon-Mamba-7B-Instruct-GGUF", "chat,ssm", "4.4"},

		// ── LFM / Liquid AI ───────────────────────────────────────────────────
		{"lfm:3b", "bartowski/LFM-3B-GGUF", "chat,edge", "1.9"},
		{"lfm:7b", "bartowski/LFM-7B-GGUF", "chat", "4.3"},
		{"lfm:40b", "bartowski/LFM-40B-MoE-GGUF", "chat,moe,general", "24.6"},

		// ── Minimax ───────────────────────────────────────────────────────────
		{"minimax-text01:456b", "bartowski/MiniMax-Text-01-GGUF", "chat,moe,general", "280.0"},

		// ── Seed ──────────────────────────────────────────────────────────────
		{"seed-thinking:20b", "bartowski/Seed-Thinking-v1.5-GGUF", "reasoning", "12.4"},

		// ── Perplexity ────────────────────────────────────────────────────────
		{"r1-1776:70b", "bartowski/r1-1776-GGUF", "reasoning,general", "43.0"},

		// ── Alan / Alana ──────────────────────────────────────────────────────
		{"exaone4:7.8b", "bartowski/EXAONE-4.0-7.8B-GGUF", "chat,korean", "4.9"},

		// ── Phi-4 variants ────────────────────────────────────────────────────
		{"phi4-vision:14b", "bartowski/phi-4-vision-instruct-GGUF", "vision,chat", "8.6"},

		// ── Kimi (more) ───────────────────────────────────────────────────────
		{"kimi-vl-thinking:16b", "bartowski/Kimi-VL-A3B-Thinking-GGUF", "vision,reasoning", "9.9"},

		// ── MoE Variants ──────────────────────────────────────────────────────
		{"nous-mamba:2.8b", "bartowski/mamba-nous-2.8b-GGUF", "chat,ssm,edge", "1.7"},
		{"megalodon:7b", "bartowski/megalodon-7b-chat-GGUF", "chat,ssm", "4.3"},
		{"rwkv6:7b", "bartowski/RWKV-6-World-7B-GGUF", "chat", "4.3"},
		{"rwkv6:14b", "bartowski/RWKV-6-World-14B-GGUF", "chat,general", "8.6"},

		// ── Extra Embedding ───────────────────────────────────────────────────
		{"text-embedding-ada-compatible", "nomic-ai/nomic-embed-text-v1.5-GGUF", "embed", "0.1"},
		{"bge-small-en:v1.5", "gpustack/bge-small-en-v1.5-GGUF", "embed,edge", "0.1"},
		{"bge-base-en:v1.5", "gpustack/bge-base-en-v1.5-GGUF", "embed", "0.2"},
		{"e5-mistral-7b", "bartowski/e5-mistral-7b-instruct-GGUF", "embed", "4.4"},
		{"jina-embeddings-v3:570m", "jina-ai/jina-embeddings-v3-GGUF", "embed", "0.6"},
		{"qwen2.5-embedding:7b", "Qwen/Qwen2.5-7B-Instruct-GGUF", "embed", "4.7"},
		{"gte-qwen2:7b", "bartowski/gte-Qwen2-7B-instruct-GGUF", "embed", "4.7"},
		{"gte-qwen2.5:1.5b", "bartowski/gte-Qwen2.5-1.5B-instruct-GGUF", "embed,edge", "1.0"},
		{"voyage-embed:2b", "bartowski/voyage-code-3-GGUF", "embed,code", "1.4"},
		{"nomic-embed-v2:137m", "nomic-ai/nomic-embed-text-v2-moe-GGUF", "embed,moe", "0.1"},
		{"paraphrase-minilm:22m", "second-state/all-MiniLM-L12-v2-Embedding-GGUF", "embed,edge", "0.05"},

		// ── Rerankers ─────────────────────────────────────────────────────────
		{"bge-reranker-v2:238m", "gpustack/bge-reranker-v2-m3-GGUF", "rerank", "0.5"},
		{"jina-reranker-v2:278m", "bartowski/jina-reranker-v2-base-multilingual-GGUF", "rerank,multilingual", "0.6"},

		// ── Llava / Vision (more) ──────────────────────────────────────────────
		{"llava1.6-mistral:7b", "bartowski/llava-v1.6-mistral-7b-GGUF", "vision,chat", "4.4"},
		{"llava1.6-vicuna:13b", "bartowski/llava-v1.6-vicuna-13b-GGUF", "vision,chat", "7.9"},
		{"llava1.6-34b", "bartowski/llava-v1.6-34b-GGUF", "vision,general", "20.7"},
		{"idefics2:8b", "bartowski/idefics2-8b-GGUF", "vision,chat", "5.0"},
		{"idefics3:8b", "bartowski/Idefics3-8B-Llama3-GGUF", "vision,chat", "5.0"},
		{"bunny-llama3:8b", "bartowski/Bunny-Llama-3-8B-V-GGUF", "vision,chat", "5.0"},
		{"deepseek-vl2:4.5b", "bartowski/DeepSeek-VL2-Tiny-GGUF", "vision,edge", "2.8"},
		{"deepseek-vl2:16b", "bartowski/DeepSeek-VL2-Small-GGUF", "vision", "9.7"},
		{"deepseek-vl2:28b", "bartowski/DeepSeek-VL2-GGUF", "vision,general", "17.2"},
		{"aria:25b", "bartowski/Aria-GGUF", "vision,moe", "15.5"},
		{"molmo:7b", "bartowski/Molmo-7B-D-0924-GGUF", "vision,chat", "4.6"},
		{"molmo:72b", "bartowski/Molmo-72B-0924-GGUF", "vision,general", "44.5"},
		{"pixtral:124b", "bartowski/Pixtral-Large-Instruct-2411-GGUF", "vision,general", "75.5"},

		// ── Multimodal / Audio ────────────────────────────────────────────────
		{"ultravox:8b", "bartowski/ultravox-v0_5-llama-3_2-1b-GGUF", "audio,edge", "0.8"},
		{"whisper-large-v3:1.5b", "ggml-org/whisper-large-v3-GGUF", "audio,transcription", "1.5"},
		{"whisper-large-v3-turbo:809m", "ggml-org/whisper-large-v3-turbo-GGUF", "audio,transcription", "0.8"},
		{"whisper-medium:769m", "ggml-org/whisper-medium-GGUF", "audio,transcription", "0.8"},
		{"whisper-small:244m", "ggml-org/whisper-small-GGUF", "audio,transcription", "0.3"},
		{"whisper-base:74m", "ggml-org/whisper-base-GGUF", "audio,transcription,edge", "0.1"},
		{"qwen2-audio:72b", "bartowski/Qwen2-Audio-72B-Instruct-GGUF", "audio,general", "44.5"},

		// ── Code (many more) ──────────────────────────────────────────────────
		{"codegemma:9b", "bartowski/codegemma-7b-GGUF", "code", "4.6"},
		{"deepseek-coder-v2:236b", "bartowski/DeepSeek-Coder-V2-Instruct-GGUF", "code,moe,general", "144.7"},
		{"codeqwen:7b", "Qwen/CodeQwen1.5-7B-Chat-GGUF", "code", "4.7"},
		{"code-llama-python:7b", "TheBloke/CodeLlama-7B-Python-GGUF", "code", "4.1"},
		{"code-llama-python:13b", "TheBloke/CodeLlama-13B-Python-GGUF", "code", "7.9"},
		{"code-llama-python:34b", "TheBloke/CodeLlama-34B-Python-GGUF", "code", "20.7"},
		{"code-llama-python:70b", "TheBloke/CodeLlama-70B-Python-hf-GGUF", "code", "41.4"},
		{"octocoder:15b", "bartowski/octocoder-GGUF", "code", "9.2"},
		{"codebooga:34b", "bartowski/CodeBooga-34B-v0.1-GGUF", "code", "20.7"},
		{"deepcoder:14b", "bartowski/DeepCoder-14B-Preview-GGUF", "code,reasoning", "9.0"},
		{"granite-code:7b", "bartowski/granite-7b-instruct-8k-GGUF", "code", "4.3"},
		{"koder:7b", "bartowski/koder-GGUF", "code", "4.3"},
		{"qwen2.5-coder:3b", "Qwen/Qwen2.5-Coder-3B-Instruct-GGUF", "code,edge", "2.0"},
		{"qwen2.5-coder:0.5b", "Qwen/Qwen2.5-Coder-0.5B-Instruct-GGUF", "code,edge", "0.4"},
		{"starcoder:7b", "bartowski/starcoder2-7b-GGUF", "code", "4.4"},
		{"polycoder:2.7b", "bartowski/polyglot-ko-5.8b-GGUF", "code,multilingual", "3.6"},
		{"arctic-code:100b", "bartowski/Snowflake-Arctic-GGUF", "code,moe,general", "61.3"},
		{"devstral-small:24b", "bartowski/Devstral-Small-2505-GGUF", "code,agent", "14.8"},
		{"qwen2.5-coder:7b-instruct-abliterated", "bartowski/Qwen2.5-Coder-7B-Instruct-abliterated-GGUF", "code", "4.7"},

		// ── Math / STEM (more) ────────────────────────────────────────────────
		{"internlm-math:7b", "bartowski/internlm2-math-7b-GGUF", "math", "4.4"},
		{"internlm-math:20b", "bartowski/internlm2-math-20b-GGUF", "math", "12.2"},
		{"metamath:7b", "TheBloke/MetaMath-Mistral-7B-GGUF", "math", "4.4"},
		{"metamath:13b", "TheBloke/MetaMath-Llama-2-13B-GGUF", "math", "7.9"},
		{"metamath:70b", "TheBloke/MetaMath-70B-V1.0-GGUF", "math,general", "41.4"},
		{"abel:7b", "TheBloke/Abel-7B-002-GGUF", "math", "4.1"},
		{"mammoth:7b", "TheBloke/MAmmoTH-7B-Instruct-GGUF", "math", "4.1"},
		{"mammoth:70b", "TheBloke/MAmmoTH-70B-Instruct-GGUF", "math,general", "41.4"},

		// ── Medical (more) ────────────────────────────────────────────────────
		{"aloe:8b", "bartowski/aloe-beta-8b-GGUF", "medical", "5.0"},
		{"biollm:7b", "TheBloke/BioMedGPT-LM-7B-GGUF", "medical", "4.1"},
		{"llama3-med42:70b", "bartowski/Llama3-Med42-70B-GGUF", "medical,general", "43.0"},
		{"med-llama3:8b", "bartowski/Meta-Llama-3-8B-Instruct-Medical-GGUF", "medical", "5.0"},
		{"openbiollm:70b", "bartowski/OpenBioLLM-70B-GGUF", "medical,general", "43.0"},
		{"asclepius:13b", "TheBloke/Asclepius-Llama2-13B-GGUF", "medical", "7.9"},
		{"medfound:7b", "bartowski/MedFound-7B-GGUF", "medical", "4.3"},

		// ── Legal ─────────────────────────────────────────────────────────────
		{"lawlm:13b", "bartowski/law-llm-13b-GGUF", "legal", "7.9"},
		{"layla-legal:8b", "bartowski/Llama-3.1-8B-LegalBench-GGUF", "legal", "5.0"},

		// ── Finance ───────────────────────────────────────────────────────────
		{"fingpt:7b", "TheBloke/FinGPT_v3.3-llama2-7b-GGUF", "finance", "4.1"},
		{"fingpt:13b", "TheBloke/FinGPT_v3.3-llama2-13b-GGUF", "finance", "7.9"},
		{"finance-llm:13b", "bartowski/Finance-LLM-13B-GGUF", "finance", "7.9"},

		// ── Reasoning (more) ──────────────────────────────────────────────────
		{"deepscaler:1.5b", "bartowski/DeepScaleR-1.5B-Preview-GGUF", "reasoning,math,edge", "1.0"},
		{"light-r1:14b", "bartowski/Light-R1-14B-GGUF", "reasoning", "9.0"},
		{"open-thoughts:7b", "bartowski/OpenThoughts-114-7B-GGUF", "reasoning", "4.4"},
		{"open-thoughts:32b", "bartowski/OpenThoughts-114-32B-GGUF", "reasoning", "20.0"},
		{"s1:32b", "bartowski/s1-32B-GGUF", "reasoning", "20.0"},
		{"simpleo1:7b", "bartowski/simpleO1-GGUF", "reasoning", "4.4"},
		{"yet-another-o1:7b", "bartowski/YetiLM-GGUF", "reasoning", "4.4"},
		{"still-3:8b", "bartowski/STILL-3-1.5B-preview-GGUF", "reasoning,edge", "1.0"},
		{"marco-o1:7b", "bartowski/Marco-o1-GGUF", "reasoning", "4.4"},
		{"journey-math:7b", "bartowski/QwQ-LCoT2-7B-GGUF", "reasoning,math", "4.4"},

		// ── Agent / Tool Use (more) ───────────────────────────────────────────
		{"functionary-v3:8b", "meetkai/functionary-small-v3.2-GGUF", "tools,chat,edge", "5.0"},
		{"functionary-v3:70b", "bartowski/functionary-medium-v3.2-GGUF", "tools,general", "43.0"},
		{"toolace:8b", "bartowski/ToolACE-8B-GGUF", "tools", "5.0"},
		{"hammer2.1:7b", "bartowski/hammer2.1-7b-GGUF", "tools", "4.4"},
		{"xlam:1b", "Salesforce/xLAM-1b-fc-r-GGUF", "tools,edge", "0.8"},
		{"xlam:7b", "Salesforce/xLAM-7b-fc-r-GGUF", "tools", "4.4"},
		{"octopus-v2:2b", "bartowski/octopus-v2-GGUF", "tools,edge", "1.4"},
		{"octo:7b", "bartowski/Octo-instruct-GGUF", "tools,chat", "4.4"},
		{"agent-flan:7b", "bartowski/AgentInstruct-7B-GGUF", "tools,agent", "4.3"},

		// ── Long Context (more) ───────────────────────────────────────────────
		{"kimi-moonshot:8b", "bartowski/Kimi-Moonshot-7B-Instruct-GGUF", "chat,long-context", "4.4"},
		{"yi-long:9b", "bartowski/Yi-1.5-9B-32K-GGUF", "chat,long-context", "5.6"},
		{"qwen2.5-72b-long", "bartowski/Qwen2.5-72B-Instruct-GGUF", "chat,long-context,general", "44.5"},
		{"chatglm-long:6b", "bartowski/glm-4-9b-chat-1m-GGUF", "chat,chinese,long-context", "5.6"},

		// ── Compact / Efficient ───────────────────────────────────────────────
		{"h2o-danube3:500m", "bartowski/h2o-danube3-500m-chat-GGUF", "chat,edge", "0.3"},
		{"h2o-danube3:4b", "bartowski/h2o-danube3-4b-chat-GGUF", "chat,edge", "2.5"},
		{"h2o-danube2:1.8b", "bartowski/h2o-danube2-1.8b-chat-GGUF", "chat,edge", "1.1"},
		{"fox:8b", "bartowski/Fox-1-1.6B-instruct-v0.1-GGUF", "chat,edge", "1.0"},
		{"gemma2-it:2b", "bartowski/gemma-2-2b-it-GGUF", "chat,edge", "1.6"},
		{"stablelm2-chat:1.6b", "bartowski/stablelm-2-1_6b-chat-GGUF", "chat,edge", "1.0"},
		{"openchat3.6:8b", "bartowski/openchat-3.6-8b-20240522-GGUF", "chat", "5.0"},
		{"starling-lm:7b", "bartowski/Starling-LM-7B-beta-GGUF", "chat", "4.4"},
		{"starling-lm:alpha:7b", "bartowski/Starling-LM-7B-alpha-GGUF", "chat", "4.4"},

		// ── Multilingual (more) ───────────────────────────────────────────────
		{"euryllm:9b", "bartowski/EurAlex-9B-GGUF", "chat,multilingual", "5.5"},
		{"vikhr-7b", "bartowski/Vikhr-7B-instruct_0.4-GGUF", "chat,russian", "4.4"},
		{"saiga:7b", "bartowski/saiga_llama3_8b-GGUF", "chat,russian", "5.0"},
		{"gigachat:13b", "bartowski/GigaSaiga-13B-GGUF", "chat,russian", "7.9"},
		{"belebele:7b", "bartowski/Belebele-7B-GGUF", "chat,multilingual", "4.4"},
		{"alma-13b", "bartowski/ALMA-13B-GGUF", "chat,translation", "7.9"},
		{"mgsm-llama3:8b", "bartowski/Llama-3-8B-Instruct-mgsm-GGUF", "math,multilingual", "5.0"},

		// ── Summarization / Document ──────────────────────────────────────────
		{"recomp-llm:7b", "bartowski/RECOMP-abstractive-7b-GGUF", "tools,summarize", "4.4"},
		{"longformer-large:435m", "bartowski/longformer-large-4096-GGUF", "tools,long-context", "0.5"},

		// ── More Roleplay / Creative ──────────────────────────────────────────
		{"euryale:70b", "bartowski/Euryale-70B-v2.1-GGUF", "chat,roleplay,general", "43.0"},
		{"euryale:8b", "bartowski/Euryale-L3.3-8B-GGUF", "chat,roleplay", "5.0"},
		{"stheno:12b", "bartowski/Stheno-3.5-L3.1-12B-GGUF", "chat,roleplay", "7.5"},
		{"lumimaid:8b", "bartowski/Lumimaid-v0.2-8B-GGUF", "chat,roleplay", "5.0"},
		{"lumimaid:70b", "bartowski/Lumimaid-v0.2-70B-GGUF", "chat,roleplay,general", "43.0"},
		{"venice-uncensored:7b", "bartowski/Venice-Mistral-7B-Instruct-GGUF", "chat", "4.4"},
		{"midnight-miqu:70b", "bartowski/MidnightMiqu-70B-v1.5-GGUF", "chat,roleplay,general", "43.0"},
		{"weaver-alpha:12b", "bartowski/Weaver-Alpha-GGUF", "chat,roleplay", "7.5"},

		// ── Tiny / Edge (more) ────────────────────────────────────────────────
		{"qwen3:0.6b-instruct", "bartowski/Qwen3-0.6B-GGUF", "chat,edge", "0.4"},
		{"gemma3:1b-it", "bartowski/gemma-3-1b-it-GGUF", "chat,edge", "0.8"},
		{"llama3.2:1b-instruct", "bartowski/Llama-3.2-1B-Instruct-GGUF", "chat,edge", "0.8"},
		{"microsoft-phi-1:1.3b", "bartowski/phi-1-GGUF", "chat,edge", "0.8"},
		{"smollm:1.7b-v0.2", "bartowski/SmolLM-1.7B-Instruct-GGUF", "chat,edge", "1.1"},
		{"smollm:360m-v0.2", "bartowski/SmolLM-360M-Instruct-GGUF", "chat,edge", "0.2"},
		{"smollm:135m-v0.2", "bartowski/SmolLM-135M-Instruct-GGUF", "chat,edge", "0.1"},
		{"qwen2-0.5b-instruct", "Qwen/Qwen2-0.5B-Instruct-GGUF", "chat,edge", "0.4"},
		{"codegeex2:6b", "bartowski/codegeex2-6b-GGUF", "code,chinese,edge", "3.7"},
		{"stablebeluga:7b", "TheBloke/StableBeluga-7B-GGUF", "chat,edge", "4.1"},
		{"incite:3b", "TheBloke/RedPajama-INCITE-Instruct-3B-v1-GGUF", "chat,edge", "1.9"},
		{"opt-6.7b", "bartowski/opt-6.7b-GGUF", "chat,research", "4.1"},
		{"opt-13b", "bartowski/opt-13b-GGUF", "chat,research", "7.9"},
		{"cerebras-gpt:6.7b", "bartowski/Cerebras-GPT-6.7B-GGUF", "research", "4.1"},

		// ── DBRX ──────────────────────────────────────────────────────────────
		{"dbrx:132b", "bartowski/dbrx-instruct-GGUF", "chat,moe,general", "81.1"},

		// ── Arctic / Snowflake ────────────────────────────────────────────────
		{"snowflake-arctic-instruct:480b", "bartowski/snowflake-arctic-instruct-GGUF", "chat,moe,general", "294.4"},

		// ── Reflection / Self-refine ──────────────────────────────────────────
		{"llama3-selfcorrect:8b", "bartowski/self-correct-llama3-8b-GGUF", "reasoning,chat", "5.0"},

		// ── Coding Agent ──────────────────────────────────────────────────────
		{"autocoder:6.7b", "TheBloke/Autocoder-GGUF", "code", "4.1"},
		{"codemonkey:7b", "bartowski/CodeMonkey-7B-GGUF", "code,agent", "4.3"},
		{"starcoder2-instruct:15b", "bartowski/starcoder2-15b-instruct-v0.1-GGUF", "code", "9.2"},
		{"wizardcoder-python:34b", "bartowski/WizardCoder-Python-34B-V1.0-GGUF", "code", "20.7"},
		{"deepseek-coder:7b-base", "bartowski/deepseek-coder-7b-base-v1.5-GGUF", "code", "4.4"},
		{"deepseek-coder:1.3b-instruct", "bartowski/deepseek-coder-1.3b-instruct-GGUF", "code,edge", "0.8"},
		{"opencodeinterpreter:7b", "bartowski/OpenCodeInterpreter-DS-6.7B-GGUF", "code", "4.1"},
		{"opencodeinterpreter:33b", "bartowski/OpenCodeInterpreter-DS-33B-GGUF", "code,general", "20.0"},

		// ── More General Chat ──────────────────────────────────────────────────
		{"nous-hermes2-theta:8b", "bartowski/Hermes-2-Theta-Llama-3-8B-GGUF", "chat", "5.0"},
		{"llama3-euryale:8b", "bartowski/Llama-3-Euryale-8B-v2.3-GGUF", "chat", "5.0"},
		{"mythomax:13b", "bartowski/MythoMax-L2-13b-GGUF", "chat,roleplay", "7.9"},
		{"mythomist:7b", "TheBloke/MythoMist-7B-GGUF", "chat,roleplay", "4.4"},
		{"l3-soliloquy:8b", "bartowski/L3-SoliloQuY-8B-GGUF", "chat,creative", "5.0"},
		{"l3-nitro:8b", "bartowski/L3-Nitro-Tess3-8B-GGUF", "chat", "5.0"},
		{"rocinante:12b", "bartowski/Rocinante-12B-v2.2-GGUF", "chat,roleplay", "7.5"},
		{"qwen3-coder:32b-instruct", "bartowski/Qwen3-32B-GGUF", "code,reasoning", "20.4"},
		{"arliai-rpmax:8b", "bartowski/Arliai-RPMax-L3.3-8B-GGUF", "chat,roleplay", "5.0"},
		{"mistral-openorca-8x7b", "bartowski/MistralOrca-8x7B-GGUF", "chat,general", "26.4"},
		{"dolphin-deepseek-r1:14b", "bartowski/dolphin-r1-qwen-14b-GGUF", "chat,reasoning", "9.0"},
		{"dolphin-deepseek-r1:32b", "bartowski/dolphin-r1-mistral-24b-GGUF", "chat,reasoning", "14.8"},
		{"soliloquy:8b", "bartowski/Soliloquy-L3.2-8B-v2.1-GGUF", "chat,creative", "5.0"},
		{"fimbulvetr:11b", "bartowski/Fimbulvetr-11B-v2-GGUF", "chat,creative", "6.8"},
		{"darkidol:8b", "bartowski/DarkIdol-L3.3-8B-GGUF", "chat,roleplay", "5.0"},
		{"llama3-uncensored:8b", "bartowski/Meta-Llama-3-8B-Instruct-Uncensored-GGUF", "chat", "5.0"},
		{"tiger-gemma:9b", "bartowski/Tiger-Gemma-9B-v3-GGUF", "chat", "5.5"},

		// ── Llama 2 (full family) ─────────────────────────────────────────────
		{"llama2:7b-chat", "TheBloke/Llama-2-7B-Chat-GGUF", "chat", "4.1"},
		{"llama2:13b-chat", "TheBloke/Llama-2-13B-Chat-GGUF", "chat", "7.9"},
		{"llama2:70b-chat", "TheBloke/Llama-2-70B-Chat-GGUF", "chat,general", "41.4"},
		{"llama2:7b", "TheBloke/Llama-2-7B-GGUF", "chat", "4.1"},
		{"llama2:13b", "TheBloke/Llama-2-13B-GGUF", "chat", "7.9"},
		{"llama2:70b", "TheBloke/Llama-2-70B-GGUF", "chat,general", "41.4"},
		{"llama2-uncensored:7b", "TheBloke/llama-2-7b-Uncensored-GGUF", "chat", "4.1"},
		{"llama2-uncensored:13b", "TheBloke/llama-2-13b-Uncensored-GGUF", "chat", "7.9"},

		// ── Alpaca / RLHF classics ────────────────────────────────────────────
		{"alpaca:7b", "TheBloke/gpt4-x-alpaca-GGUF", "chat", "4.1"},
		{"guanaco:7b", "TheBloke/guanaco-7B-GGUF", "chat", "4.1"},
		{"guanaco:13b", "TheBloke/guanaco-13B-GGUF", "chat", "7.9"},
		{"guanaco:33b", "TheBloke/guanaco-33B-GGUF", "chat", "20.0"},
		{"guanaco:65b", "TheBloke/guanaco-65B-GGUF", "chat,general", "41.4"},
		{"airoboros:7b", "TheBloke/Airoboros-L2-7B-GGUF", "chat", "4.1"},
		{"airoboros:13b", "TheBloke/Airoboros-L2-13B-GGUF", "chat", "7.9"},
		{"airoboros:70b", "TheBloke/Airoboros-L2-70B-GGUF", "chat,general", "41.4"},
		{"lazarus:30b", "TheBloke/Lazarus-30B-GGUF", "chat", "18.3"},
		{"upstage-solar:10.7b", "TheBloke/SOLAR-10.7B-Instruct-v1.0-GGUF", "chat", "6.6"},
		{"beluga:7b", "TheBloke/StableBeluga-7B-GGUF", "chat", "4.1"},
		{"beluga:13b", "TheBloke/StableBeluga-13B-GGUF", "chat", "7.9"},
		{"beluga2:70b", "TheBloke/StableBeluga2-GGUF", "chat,general", "41.4"},
		{"chronos:13b", "TheBloke/ChronosHermes-13b-GGUF", "chat,roleplay", "7.9"},
		{"tigerbot:7b", "TheBloke/tigerbot-7b-chat-GGUF", "chat,chinese", "4.4"},
		{"tigerbot:13b", "TheBloke/tigerbot-13b-chat-v4-GGUF", "chat,chinese", "7.9"},

		// ── WizardLM (more) ───────────────────────────────────────────────────
		{"wizardlm:13b-v1.2", "TheBloke/WizardLM-13B-V1.2-GGUF", "chat", "7.9"},
		{"wizardlm:30b-v1", "TheBloke/WizardLM-30B-V1.0-GGUF", "chat,general", "18.3"},
		{"wizardlm2:8x22b", "bartowski/WizardLM-2-8x22B-GGUF", "chat,moe,general", "79.4"},
		{"evol-instruct:70b", "TheBloke/WizardLM-70B-V1.0-GGUF", "chat,general", "41.4"},

		// ── Vicuna (full) ─────────────────────────────────────────────────────
		{"vicuna:7b-v1.5", "TheBloke/vicuna-7B-v1.5-GGUF", "chat", "4.1"},
		{"vicuna:13b-v1.5", "TheBloke/vicuna-13B-v1.5-GGUF", "chat", "7.9"},
		{"vicuna:33b-v1.3", "TheBloke/vicuna-33B-v1.3-GGUF", "chat,general", "20.0"},

		// ── Llama 3 Gradient ──────────────────────────────────────────────────
		{"llama3-gradient:8b", "bartowski/Llama-3-8B-Instruct-Gradient-1048k-GGUF", "chat,long-context", "5.0"},
		{"llama3-gradient:70b", "bartowski/Llama-3-70B-Instruct-Gradient-1048k-GGUF", "chat,long-context,general", "43.0"},

		// ── Mistral community variants ────────────────────────────────────────
		{"mistral-tiny-instruct:7b", "bartowski/Mistral-7B-Instruct-v0.3-GGUF", "chat", "4.4"},
		{"mistral-openhermes:7b", "bartowski/Mistral-7B-OpenOrca-GGUF", "chat", "4.4"},
		{"mistral-text:7b", "bartowski/Mistral-7B-v0.3-GGUF", "chat", "4.4"},
		{"notus:7b", "bartowski/notus-7b-v1-GGUF", "chat", "4.4"},
		{"notux:8x7b", "bartowski/notux-8x7b-v1-GGUF", "chat,moe", "26.4"},

		// ── Intel NeuralChat (more) ───────────────────────────────────────────
		{"neural-chat:7b-v3.3", "bartowski/neural-chat-7b-v3-3-GGUF", "chat", "4.4"},

		// ── Qwen2 (more) ──────────────────────────────────────────────────────
		{"qwen2:72b-instruct", "bartowski/Qwen2-72B-Instruct-GGUF", "chat,general", "44.5"},
		{"qwen2:7b-instruct", "Qwen/Qwen2-7B-Instruct-GGUF", "chat", "4.7"},
		{"qwen2:1.5b-instruct", "Qwen/Qwen2-1.5B-Instruct-GGUF", "chat,edge", "1.0"},

		// ── Llama 4 Scout / Maverick ──────────────────────────────────────────
		{"llama4:scout:17b", "bartowski/Llama-4-Scout-17B-16E-Instruct-GGUF", "chat,moe,vision", "10.6"},
		{"llama4:maverick:17b", "bartowski/Llama-4-Maverick-17B-128E-Instruct-GGUF", "chat,moe,vision", "10.6"},

		// ── TinyLlama ─────────────────────────────────────────────────────────
		{"tinyllama:1.1b", "TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF", "chat,edge", "0.7"},
		{"tinyllama:1.1b-1t", "bartowski/TinyLlama_v1.1-1T-OpenHermes-2.5-GGUF", "chat,edge", "0.7"},

		// ── SQLCoder / Text-to-SQL (more) ─────────────────────────────────────
		{"sqlcoder-2:7b", "bartowski/sqlcoder-7b-2-GGUF", "sql", "4.4"},
		{"sqlcoder:34b", "bartowski/sqlcoder-34b-alpha-GGUF", "sql,general", "20.7"},
		{"nsql-llama:7b", "bartowski/nsql-llama-2-7B-GGUF", "sql,edge", "4.1"},
		{"defog-llama3:8b", "bartowski/defog-llama-3-sqlcoder-8b-GGUF", "sql", "5.0"},

		// ── Falcon2 ───────────────────────────────────────────────────────────
		{"falcon2:11b", "bartowski/falcon-11b-GGUF", "chat", "6.8"},

		// ── Mixtral community variants ────────────────────────────────────────
		{"mixtral-dolphin:8x7b", "bartowski/dolphin-2.7-mixtral-8x7b-GGUF", "chat,moe", "26.4"},
		{"miqu-70b", "bartowski/miqu-1-70b-GGUF", "chat,general", "43.0"},

		// ── Deep Hermes ───────────────────────────────────────────────────────
		{"deephermes-3:8b", "bartowski/DeepHermes-3-Llama-3-8B-Preview-GGUF", "chat,reasoning", "5.0"},
		{"deephermes-3:70b", "bartowski/DeepHermes-3-Llama-3.1-70B-GGUF", "chat,reasoning,general", "43.0"},

		// ── Mistral AI flagship ────────────────────────────────────────────────
		{"mistral-large:123b", "bartowski/Mistral-Large-Instruct-2407-GGUF", "chat,general", "75.5"},
		{"codestral-22b", "bartowski/Codestral-22B-v0.1-GGUF", "code", "13.5"},

		// ── WizardMath ────────────────────────────────────────────────────────
		{"wizardmath:7b", "TheBloke/WizardMath-7B-V1.1-GGUF", "math", "4.4"},
		{"wizardmath:13b", "TheBloke/WizardMath-13B-V1.0-GGUF", "math", "7.9"},
		{"wizardmath:70b", "TheBloke/WizardMath-70B-V1.0-GGUF", "math,general", "41.4"},

		// ── More Medical ──────────────────────────────────────────────────────
		{"meditron:7b", "bartowski/meditron-7b-GGUF", "medical", "4.4"},
		{"meditron:70b", "bartowski/meditron-70b-GGUF", "medical,general", "43.0"},
		{"clinicalcamel:70b", "TheBloke/ClinicalCamel-70B-GGUF", "medical,general", "41.4"},
		{"bioplatypus:70b", "TheBloke/BioPlatypus-70B-GGUF", "medical,general", "41.4"},

		// ── More Code ─────────────────────────────────────────────────────────
		{"phind-codellama:34b-v2", "TheBloke/Phind-CodeLlama-34B-v2-GGUF", "code,general", "20.7"},
		{"refact:1.6b", "TheBloke/refact-1_6B-fim-GGUF", "code,edge", "1.0"},
		{"starcoder2-instruct:15b", "bartowski/starcoder2-15b-instruct-v0.1-GGUF", "code", "9.2"},
		{"wizardcoder-python:34b", "bartowski/WizardCoder-Python-34B-V1.0-GGUF", "code", "20.7"},
		{"deepseek-coder:1.3b-instruct", "bartowski/deepseek-coder-1.3b-instruct-GGUF", "code,edge", "0.8"},
		{"opencodeinterpreter:7b", "bartowski/OpenCodeInterpreter-DS-6.7B-GGUF", "code", "4.1"},
		{"opencodeinterpreter:33b", "bartowski/OpenCodeInterpreter-DS-33B-GGUF", "code,general", "20.0"},
		{"autocoder:6.7b", "TheBloke/Autocoder-GGUF", "code", "4.1"},

		// ── Platypus (more) ───────────────────────────────────────────────────
		{"platypus2:70b", "TheBloke/Platypus2-70B-instruct-GGUF", "chat,general", "41.4"},
		{"open-platypus:13b", "TheBloke/Open-Platypus-13B-GGUF", "chat", "7.9"},

		// ── Tulu / AllenAI ────────────────────────────────────────────────────
		{"tulu3:8b", "bartowski/Llama-3.1-Tulu-3-8B-GGUF", "chat", "5.0"},
		{"tulu3:70b", "bartowski/Llama-3.1-Tulu-3-70B-GGUF", "chat,general", "43.0"},
		{"olmo2:7b", "bartowski/OLMo-2-1124-7B-Instruct-GGUF", "chat,research", "4.4"},
		{"olmo2:13b", "bartowski/OLMo-2-1124-13B-Instruct-GGUF", "chat,research", "8.0"},

		// ── Hermes 3 flagship ─────────────────────────────────────────────────
		{"hermes3:405b", "bartowski/Hermes-3-Llama-3.1-405B-GGUF", "chat,tools,general", "248.0"},

		// ── Nous Research (more) ──────────────────────────────────────────────
		{"nous-capybara:34b", "TheBloke/Nous-Capybara-34B-GGUF", "chat,general", "20.7"},
		{"nous-hermes-2-mixtral:8x7b", "bartowski/Nous-Hermes-2-Mixtral-8x7B-DPO-GGUF", "chat,moe", "26.4"},
		{"nous-hermes-2-mistral:7b-dpo", "bartowski/Nous-Hermes-2-Mistral-7B-DPO-GGUF", "chat", "4.4"},

		// ── Orca (more) ───────────────────────────────────────────────────────
		{"orca-mini:7b", "TheBloke/orca_mini_v3_7B-GGUF", "chat", "4.1"},
		{"orca-mini:13b", "TheBloke/orca_mini_v3_13B-GGUF", "chat", "7.9"},
		{"orca-mini:70b", "TheBloke/orca_mini_v3_70B-GGUF", "chat,general", "41.4"},
		{"orca2:7b", "TheBloke/Orca-2-7B-GGUF", "chat", "4.1"},
		{"orca2:13b", "TheBloke/Orca-2-13B-GGUF", "chat", "7.9"},

		// ── Baichuan ──────────────────────────────────────────────────────────
		{"baichuan2:7b", "TheBloke/Baichuan2-7B-Chat-GGUF", "chat,chinese", "4.3"},
		{"baichuan2:13b", "TheBloke/Baichuan2-13B-Chat-GGUF", "chat,chinese", "7.9"},

		// ── SOLAR variants ────────────────────────────────────────────────────
		{"solar-pro:22b", "bartowski/SOLAR-Pro-Preview-Instruct-GGUF", "chat", "13.5"},

		// ── MPT ───────────────────────────────────────────────────────────────
		{"mpt:7b-chat", "TheBloke/mpt-7b-chat-GGUF", "chat", "4.0"},
		{"mpt:30b-chat", "TheBloke/mpt-30b-chat-GGUF", "chat,general", "18.3"},
		{"mpt:7b-storywriter", "TheBloke/mpt-7b-storywriter-GGUF", "chat,creative", "4.0"},

		// ── Qwen 1 (legacy) ───────────────────────────────────────────────────
		{"qwen1.5:7b-chat", "Qwen/Qwen1.5-7B-Chat-GGUF", "chat", "4.7"},
		{"qwen1.5:14b-chat", "Qwen/Qwen1.5-14B-Chat-GGUF", "chat", "9.0"},
		{"qwen1.5:72b-chat", "Qwen/Qwen1.5-72B-Chat-GGUF", "chat,general", "44.5"},
		{"qwen1.5:110b-chat", "Qwen/Qwen1.5-110B-Chat-GGUF", "chat,general", "67.4"},

		// ── EXAONE (more) ─────────────────────────────────────────────────────
		{"exaone3.5:7.8b", "bartowski/EXAONE-3.5-7.8B-Instruct-GGUF", "chat,korean", "4.9"},
		{"exaone3.5:2.4b", "bartowski/EXAONE-3.5-2.4B-Instruct-GGUF", "chat,korean,edge", "1.5"},
		{"exaone3.5:32b", "bartowski/EXAONE-3.5-32B-Instruct-GGUF", "chat,korean,general", "20.0"},

		// ── Aya Expanse ───────────────────────────────────────────────────────
		{"aya-expanse:8b", "bartowski/aya-expanse-8b-GGUF", "chat,multilingual", "5.0"},
		{"aya-expanse:32b", "bartowski/aya-expanse-32b-GGUF", "chat,multilingual,general", "20.0"},

		// ── Research / Base models ────────────────────────────────────────────
		{"open-llama:3b", "bartowski/open_llama_3b_v2-GGUF", "chat,edge", "1.9"},
		{"open-llama:7b", "bartowski/open_llama_7b_v2-GGUF", "chat", "4.1"},
		{"pythia:12b", "bartowski/pythia-12b-GGUF", "research", "7.3"},
		{"neox:20b", "TheBloke/GPT-NeoX-20B-GGUF", "research", "12.2"},
		{"gpt-j:6b", "bartowski/gpt-j-6b-GGUF", "research", "3.8"},
		{"santacoder:1.1b", "bartowski/santacoder-GGUF", "code,edge", "0.7"},
		{"opt-6.7b", "bartowski/opt-6.7b-GGUF", "research", "4.1"},
		{"opt-13b", "bartowski/opt-13b-GGUF", "research", "7.9"},

		// ── Falcon3 (full) ────────────────────────────────────────────────────
		{"falcon3:1b", "bartowski/Falcon3-1B-Instruct-GGUF", "chat,edge", "0.8"},
		{"falcon3:3b", "bartowski/Falcon3-3B-Instruct-GGUF", "chat,edge", "2.0"},
		{"falcon3:7b", "bartowski/Falcon3-7B-Instruct-GGUF", "chat", "4.4"},
		{"falcon3:10b", "bartowski/Falcon3-10B-Instruct-GGUF", "chat", "6.2"},
		{"falcon2:11b", "bartowski/falcon-11b-GGUF", "chat", "6.8"},

		// ── Phi-3 128k ────────────────────────────────────────────────────────
		{"phi3:mini:128k", "bartowski/Phi-3-mini-128k-instruct-GGUF", "chat,edge", "2.2"},
		{"phi3:medium:128k", "bartowski/Phi-3-medium-128k-instruct-GGUF", "chat", "8.6"},
		{"phi3.5:vision:4.2b", "bartowski/Phi-3.5-vision-instruct-GGUF", "vision,edge", "2.8"},

		// ── Gemma 2 large ─────────────────────────────────────────────────────
		{"gemma2:27b-it", "bartowski/gemma-2-27b-it-GGUF", "chat,general", "16.9"},

		// ── SEA multilingual ──────────────────────────────────────────────────
		{"sailor:7b", "bartowski/Sailor-7B-Chat-GGUF", "chat,multilingual", "4.4"},
		{"seallm:7b", "bartowski/SeaLLM-7B-v2.5-GGUF", "chat,multilingual", "4.4"},

		// ── Chinese specific ──────────────────────────────────────────────────
		{"chinese-alpaca-2:7b", "TheBloke/Chinese-Alpaca-2-7B-GGUF", "chat,chinese", "4.1"},
		{"chinese-alpaca-2:13b", "TheBloke/Chinese-Alpaca-2-13B-GGUF", "chat,chinese", "7.9"},

		// ── Zephyr (more) ─────────────────────────────────────────────────────
		{"zephyr-141b-a39b", "bartowski/Zephyr-141B-A39B-v0.1-GGUF", "chat,moe,general", "85.4"},

		// ── DBRX ──────────────────────────────────────────────────────────────
		{"dbrx:132b", "bartowski/dbrx-instruct-GGUF", "chat,moe,general", "81.1"},

		// ── Snowflake Arctic ──────────────────────────────────────────────────
		{"snowflake-arctic:480b", "bartowski/snowflake-arctic-instruct-GGUF", "chat,moe,general", "294.4"},

		// ── FreeWilly ─────────────────────────────────────────────────────────
		{"freewilly2:70b", "TheBloke/FreeWilly2-GGUF", "chat,general", "41.4"},
		{"freewilly:13b", "TheBloke/FreeWilly1-GGUF", "chat", "7.9"},

		// ── Samantha ──────────────────────────────────────────────────────────
		{"samantha:7b", "TheBloke/samantha-1.11-7b-GGUF", "chat,roleplay", "4.1"},
		{"samantha:13b", "TheBloke/samantha-1.11-13b-GGUF", "chat,roleplay", "7.9"},
		{"samantha:70b", "TheBloke/Samantha-70B-GGUF", "chat,roleplay,general", "41.4"},

		// ── Mistral NeMo (more) ───────────────────────────────────────────────
		{"mistral-nemo:12b-2407", "bartowski/Mistral-Nemo-Instruct-2407-GGUF", "chat", "7.3"},

		// ── Merged / Fused ────────────────────────────────────────────────────
		{"goliath:120b", "bartowski/goliath-120b-GGUF", "chat,general", "73.3"},
		{"bagel:8b", "bartowski/bagel-8b-v1-GGUF", "chat", "5.0"},
		{"llama-pro:8b", "bartowski/llama-pro-8b-instruct-GGUF", "chat", "5.0"},
		{"reflection:70b", "bartowski/Reflection-Llama-3.1-70B-GGUF", "chat,general", "43.0"},

		// ── Instructlab ───────────────────────────────────────────────────────
		{"instructlab:7b", "bartowski/instructlab-merlinite-7b-GGUF", "chat", "4.4"},

		// ── SmolLM2 ───────────────────────────────────────────────────────────
		{"smollm2:135m", "bartowski/SmolLM2-135M-Instruct-GGUF", "chat,edge", "0.1"},
		{"smollm2:360m", "bartowski/SmolLM2-360M-Instruct-GGUF", "chat,edge", "0.2"},
		{"smollm2:1.7b", "bartowski/SmolLM2-1.7B-Instruct-GGUF", "chat,edge", "1.1"},

		// ── Gemma 3 full ──────────────────────────────────────────────────────
		{"gemma3:12b-it", "bartowski/gemma-3-12b-it-GGUF", "chat", "7.5"},

		// ── Phi 4 full ────────────────────────────────────────────────────────
		{"phi4:14b", "bartowski/phi-4-GGUF", "chat", "8.6"},
		{"phi4-vision:14b", "bartowski/phi-4-vision-instruct-GGUF", "vision,chat", "8.6"},

		// ── Qwen 2.5 math ─────────────────────────────────────────────────────
		{"qwen2.5-math:7b", "Qwen/Qwen2.5-Math-7B-Instruct-GGUF", "math", "4.7"},
		{"qwen2.5-math:72b", "Qwen/Qwen2.5-Math-72B-Instruct-GGUF", "math,general", "44.5"},

		// ── Llama 3.3 ─────────────────────────────────────────────────────────
		{"llama3.3:70b", "bartowski/Llama-3.3-70B-Instruct-GGUF", "chat,general", "43.0"},

		// ── Kimi ──────────────────────────────────────────────────────────────
		{"kimi-vl-thinking:16b", "bartowski/Kimi-VL-A3B-Thinking-GGUF", "vision,reasoning", "9.9"},

		// ── Grok ─────────────────────────────────────────────────────────────
		{"grok-2-mini:27b", "bartowski/grok-2-27b-GGUF", "chat,general", "16.6"},

		// ── LFM / Liquid ──────────────────────────────────────────────────────
		{"lfm2:1b", "bartowski/LFM2-1B-GGUF", "chat,edge", "0.8"},

		// ── Biomed (more) ─────────────────────────────────────────────────────
		{"aloe:8b", "bartowski/aloe-beta-8b-GGUF", "medical", "5.0"},
		{"llama3-med42:70b", "bartowski/Llama3-Med42-70B-GGUF", "medical,general", "43.0"},

		// ── Finance ───────────────────────────────────────────────────────────
		{"fingpt:7b", "TheBloke/FinGPT_v3.3-llama2-7b-GGUF", "finance", "4.1"},
		{"fingpt:13b", "TheBloke/FinGPT_v3.3-llama2-13b-GGUF", "finance", "7.9"},

		// ── Legal ─────────────────────────────────────────────────────────────
		{"layla-legal:8b", "bartowski/Llama-3.1-8B-LegalBench-GGUF", "legal", "5.0"},

		// ── Persimmon ─────────────────────────────────────────────────────────
		{"persimmon:8b", "bartowski/persimmon-8b-chat-GGUF", "chat", "5.0"},

		// ── More roleplay ─────────────────────────────────────────────────────
		{"weaver-alpha:12b", "bartowski/Weaver-Alpha-GGUF", "chat,roleplay", "7.5"},
		{"darkidol2:8b", "bartowski/DarkIdol-L3.3-8B-Instruct-GGUF", "chat,roleplay", "5.0"},
		{"erebus:7b", "TheBloke/Erebus-v3-7B-GGUF", "chat,roleplay", "4.4"},
		{"erebus:13b", "TheBloke/Erebus-v3-13B-GGUF", "chat,roleplay", "7.9"},
		{"erebus:70b", "TheBloke/Erebus-v3-70B-GGUF", "chat,roleplay,general", "41.4"},
		{"mythalion:13b", "TheBloke/Mythalion-13B-GGUF", "chat,roleplay", "7.9"},
		{"airochronos:70b", "TheBloke/AiroChronos-70B-GGUF", "chat,roleplay,general", "41.4"},

		// ── Misc popular community ────────────────────────────────────────────
		{"llama3-supernova-lite:8b", "bartowski/Llama-3.1-SuperNova-Lite-GGUF", "chat", "5.0"},
		{"midnight-rose:70b", "bartowski/MidnightRose-70B-GGUF", "chat,general", "43.0"},
		{"mmlutest:7b", "bartowski/NovaSky-7B-GGUF", "chat", "4.4"},
		{"nemomix:12b", "bartowski/NemoMix-Ollama-12B-v1-GGUF", "chat", "7.5"},
		{"tess-v2:7b", "bartowski/Tess-v2.5-Mistral-7B-GGUF", "chat", "4.4"},
		{"tess-v2:34b", "bartowski/Tess-v2.5.3-Llama-3.1-70B-GGUF", "chat,general", "43.0"},
		{"dolphin-llama3:70b", "bartowski/dolphin-2.9.2-llama3-70b-GGUF", "chat,general", "43.0"},
		{"dolphin3:8b", "bartowski/dolphin3.0-llama3.1-8b-GGUF", "chat", "5.0"},
		{"dolphin3:70b", "bartowski/dolphin3.0-llama3.1-70b-GGUF", "chat,general", "43.0"},
		{"capybara:34b", "bartowski/Nous-Capybara-34b-GGUF", "chat,general", "20.7"},
		{"kunoichi-dpo:7b", "bartowski/Kunoichi-DPO-v2-7B-GGUF", "chat,roleplay", "4.4"},
		{"alpacino-mistral:7b", "bartowski/Alpacino-Mistral-7B-GGUF", "chat", "4.4"},
		{"amethyst:7b", "bartowski/Amethyst-13B-Mistral-GGUF", "chat", "7.9"},
		{"theia:34b", "bartowski/theia-34b-GGUF", "chat,general", "20.7"},
	}
}
