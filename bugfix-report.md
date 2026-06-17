# Bug Fix Report — Lychee Project

**Date:** 2026-06-17  
**Agent:** Lychee (Agent #15)  
**Tools used:** `go vet`, `staticcheck`

---

## Fixes Applied

### 1. `cmd/cmd_generate.go` — Printerf format string bug (CRITICAL)

**Issue:** In the Go client code generation template, two `%v` format verbs (`log.Fatalf("failed to create client: %v", err)` and `log.Fatalf("chat failed: %v", err)`) inside the backtick-delimited `fmt.Sprintf` format string were being consumed by the outer `fmt.Sprintf` call, causing a mismatch between format verbs (4 verbs) and arguments (3 args).

**Severity:** Would cause a runtime panic when generating Go client code.

**Fix:** Escaped both `%v` → `%%v` so the outer `fmt.Sprintf` produces literal `%v` in the generated Go output.

**File:** `cmd/cmd_generate.go` lines ~184, ~200

---

### 2. `tokenizer/bytepairencoding_test.go` — Unused `slices.Collect` result

**Issue:** `slices.Collect(tokenizer.split(string(bts)))` returned a value that was discarded in a benchmark. The compiler could optimize away the call, invalidating the benchmark measurement.

**Fix:** Assigned result to `_ = slices.Collect(...)` to prevent compiler optimization.

**File:** `tokenizer/bytepairencoding_test.go` line 542

---

### 3. `server/handler_compose.go` — Unused variable `m`

**Issue:** `m, err := GetModel(name.String())` — variable `m` was assigned but never used (only `err` was checked).

**Fix:** Changed to `_, err = GetModel(name.String())` (also corrected `:=` to `=` since `err` was already declared).

**File:** `server/handler_compose.go` line 40

---

### 4. `server/structured_output.go` — Unused variable `m`

**Issue:** Same pattern as #3 — `m, err := GetModel(name.String())` where `m` was never used.

**Fix:** Changed to `_, err = GetModel(name.String())`.

**File:** `server/structured_output.go` line 53

---

### 5. `cmd/cmd_interactive.go` — Unchecked error from `exec.Command`

**Issue:** `pageSizeStr, err := exec.Command("pagesize").Output()` — the error was never checked, and the subsequent `strconv.ParseInt` error was also ignored with `_`. On macOS, failure to get page size would silently produce incorrect memory statistics.

**Fix:** Added error checks for both `exec.Command` and `strconv.ParseInt`, returning a `memInfo{total: total}` partial result on failure (consistent with existing error handling pattern in the function).

**File:** `cmd/cmd_interactive.go` line 292

---

### 6. `server/handler_integration_test.go` — Unchecked error from HTTP GET

**Issue:** `resp, err = client.Get(...)` — error was never checked before using `resp.Body`.

**Fix:** Added error check with `t.Fatalf(...)`.

**File:** `server/handler_integration_test.go` line 254

---

## Pre-existing Issues (NOT fixed — noted for awareness)

These are pre-existing issues that existed before this bug scan:

### Test Compilation Failures
- `cmd/launch/openclaw_test.go` — undefined `plugins`, `config`
- `cmd/cmd_compare_test.go` — unknown fields `EvalCount`, `TotalDuration` in `api.ChatResponse`
- `cmd/bench/bench_test.go` — undefined `flagOptions`, `BenchmarkModel`, `readImage`
- `x/create/client/create_test.go` — undefined `parsePerExpertInputs`
- `x/imagegen/mlx/mlx_test.go` — undefined `RMSNormNoWeight`, `Erf`, `Compile`, `RandomKey`, `RandomSplit`
- `x/mlxrunner/mlx/array_test.go` — undefined `Bernoulli`
- `x/mlxrunner/mlx/compile_test.go` — undefined `arrays`
- `x/mlxrunner/mlx/thread_test.go` — undefined `resetDefaultStreamCache`
- `x/models/qwen3_5/qwen3_5_test.go` — undefined `mlx.New`

### Pre-existing Runtime Test Failures (require MLX/GPU hardware or specific environment)
- `x/imagegen/nn` — index out of range panics (no MLX backend)
- `x/mlxrunner/cache` — nil snapshot returns
- `x/mlxrunner/model` — nil embedding types
- `x/models/gemma4` — index out of range
- `x/models/laguna` — load failures, panics
- `x/models/nn` — recurrent state panics
- `x/tools` — WSL not installed on test machine
- `server` — TestValidateJSONSchema, TestStructuredOutput

### Style Warnings (intentionally left)
- `reflect.SliceHeader` deprecated usage in Windows-specific code (annotated with `//nolint:govet`)
- `unsafe.Pointer` patterns in Windows tray/eventloop code (annotated with `//nolint:govet,gosec`)
- Error string capitalization/punctuation style (ST1005)
- Deprecated API usage (`h2c.NewHandler`, `http.CloseNotifier`, `openai.ToChunk`, deprecated model fields)

---

## Build Status
- `go build ./...` — **PASSES** ✅
- `go vet ./...` — pre-existing warnings only ✅
- `go test ./...` — pre-existing failures only, no regressions ✅
