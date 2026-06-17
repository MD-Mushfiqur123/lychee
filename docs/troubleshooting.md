# Troubleshooting FAQ

Quick-reference guide for the most common Lychee issues and their fixes.

---

## 1. "Connection refused"

**Symptoms:**
```
curl: (7) Failed to connect to localhost port 11434: Connection refused
```
Any client or API call to Lychee fails immediately with a connection error.

**Cause:** The Lychee server is not running, or it is running on a different port/host.

**Solution:**
```shell
# Start the Lychee server
lychee serve

# Verify it is listening
curl http://localhost:11434/api/tags
```

If the server is running but on a different host, set the `LYCHEE_HOST` environment variable:
```shell
LYCHEE_HOST=0.0.0.0:11434 lychee serve
```

For Docker:
```shell
docker run -d -p 11434:11434 --name lychee lychee/lychee
docker ps | grep lychee   # confirm it is running
```

---

## 2. "Model not found"

**Symptoms:**
```
Error: model "llama3" not found
```
Trying to run or pull a model name that does not exist locally.

**Cause:** The model has not been downloaded (pulled) yet, or the name is misspelled.

**Solution:**
```shell
# List all models you have pulled
lychee list

# Pull the model you need
lychee pull llama3

# Verify it is now present
lychee list | grep llama3
```

If the model name is correct but still not found, check the exact name on [Lychee Hub](https://lychee.ai/search):
```shell
lychee pull llama3:8b            # include tag
lychee pull lychee.com/library/llama3:8b   # fully qualified name
```

---

## 3. "Out of memory"

**Symptoms:**
```
CUDA out of memory. Tried to allocate 2.00 GiB (GPU 0; 8.00 GiB total capacity; 6.50 GiB already allocated)
```
```
runtime error: cannot allocate memory
```
Model loading fails, or inference crashes mid-response.

**Cause:** The model is too large for your available RAM (system or VRAM).

**Solution:**
```shell
# Use a smaller model
lychee pull phi3:mini          # ~2.4 GB
lychee pull gemma2:2b           # ~1.6 GB

# Use a quantized (smaller) version of your model
lychee pull llama3:8b-q4_0     # 4-bit quantization (~4.7 GB vs ~16 GB)
lychee pull llama3:8b-q5_K_M   # 5-bit medium quant (~5.8 GB)

# Check available models and sizes
lychee list
```

In Docker, increase memory limits:
```shell
docker run -d --memory="16g" --memory-swap="20g" -p 11434:11434 lychee/lychee
```

---

## 4. "CUDA not available"

**Symptoms:**
```
Warning: CUDA not available, falling back to CPU
```
Inference is slow, and logs show CPU fallback despite having an NVIDIA GPU.

**Cause:** NVIDIA drivers or CUDA toolkit are not installed, or the version is incompatible.

**Solution:**
```shell
# Check if NVIDIA driver is installed
nvidia-smi

# If not installed, download the latest driver from:
# https://www.nvidia.com/download/

# Check CUDA version
nvcc --version

# On Linux, verify the container runtime (if using Docker)
docker run --rm --gpus all nvidia/cuda:12.1-base nvidia-smi
```

On Linux, if `nvidia-smi` works but Lychee still falls back to CPU:
```shell
# Reload the nvidia_uvm kernel module
sudo rmmod nvidia_uvm && sudo modprobe nvidia_uvm

# Force CUDA library
LYCHEE_LLM_LIBRARY="cuda_v12" lychee serve
```

---

## 5. "Permission denied" (Linux)

**Symptoms:**
```
Error: permission denied: /usr/share/lychee/.lychee/models/...
```
```
EACCES: permission denied, open '/var/log/lychee/server.log'
```

**Cause:** The user running Lychee does not have read/write access to the required directories.

**Solution:**
```shell
# Check ownership and permissions
ls -la ~/.lychee/
ls -la ~/.lychee/models/

# Fix ownership (replace "user" with your username)
sudo chown -R $USER:$USER ~/.lychee/

# Or set appropriate permissions
chmod -R 755 ~/.lychee/
```

For systemd installations, ensure the service runs under the correct user:
```shell
sudo systemctl edit lychee
# Add:
# [Service]
# User=yourusername
# Group=yourusername

sudo systemctl daemon-reload
sudo systemctl restart lychee
```

For `/dev/kfd` or `/dev/dri` (AMD GPU):
```shell
sudo usermod -aG video $USER
sudo usermod -aG render $USER
# Log out and back in for changes to take effect
```

---

## 6. "Bind: address already in use" (port 11434)

**Symptoms:**
```
Error: listen tcp 127.0.0.1:11434: bind: address already in use
```
Lychee refuses to start because port 11434 is occupied.

**Cause:** Another instance of Lychee (or another application) is already using port 11434.

**Solution:**
```shell
# Find which process is using port 11434

# Linux / macOS:
sudo lsof -i :11434
sudo ss -tlnp | grep 11434

# Windows (PowerShell):
netstat -ano | findstr :11434

# Kill the existing process (replace PID)
kill <PID>                    # Linux/macOS
taskkill /PID <PID> /F        # Windows

# Or use a different port
lychee serve --port 11435
LYCHEE_HOST=127.0.0.1:11435 lychee serve
```

For Docker:
```shell
# Stop an existing container
docker stop lychee && docker rm lychee

# Or map to a different port
docker run -d -p 11435:11434 --name lychee lychee/lychee
```

---

## 7. Slow inference

**Symptoms:**
Responses take a long time (>10 seconds per token), or throughput is unexpectedly low.

**Cause:** Running on CPU when GPU is available, using a large unquantized model, or insufficient system resources.

**Solution:**
```shell
# 1. Confirm whether GPU is being used
#    Check server logs — look for "cuda" or "rocm" in the startup messages
#    If it says "cpu" or "cpu_avx2", GPU acceleration is disabled

# 2. Force GPU if available but not auto-detected
LYCHEE_LLM_LIBRARY="cuda_v12" lychee serve       # NVIDIA
LYCHEE_LLM_LIBRARY="rocm_v6" lychee serve         # AMD

# 3. Switch to a lower quantization for speed
lychee pull llama3:8b-q4_0       # faster, smaller

# 4. Reduce context length (less memory = faster)
LYCHEE_CONTEXT_LENGTH=2048 lychee serve

# 5. Enable debug logging to see what library is loaded
LYCHEE_DEBUG=1 lychee serve
```

On CPU-only machines, enable the fastest available instruction set:
```shell
# Linux — check CPU features
cat /proc/cpuinfo | grep flags | head -1

# Force AVX2 if supported
LYCHEE_LLM_LIBRARY="cpu_avx2" lychee serve
```

---

## 8. Docker: container exits immediately

**Symptoms:**
```
docker ps            # no lychee container listed
docker ps -a         # container status: Exited (1) X seconds ago
```
The container starts and then dies within seconds.

**Cause:** A startup error (missing models, permission issues, port conflict, or GPU access failure).

**Solution:**
```shell
# Check the container logs
docker logs lychee

# If the container name is unknown, list all containers (including stopped)
docker ps -a

# Common fixes:

# 1. Mount a persistent models volume (avoid re-downloading)
docker run -d -v lychee-models:/root/.lychee -p 11434:11434 --name lychee lychee/lychee

# 2. Enable GPU access (NVIDIA)
docker run -d --gpus all -v lychee-models:/root/.lychee -p 11434:11434 --name lychee lychee/lychee

# 3. Check the container runtime supports GPU
docker run --rm --gpus all ubuntu nvidia-smi

# 4. Run interactively to see the error
docker run --rm -it -p 11434:11434 lychee/lychee
```

---

## 9. Windows: antivirus blocking Lychee

**Symptoms:**
- Lychee fails to start with no clear error.
- `lychee serve` hangs or exits immediately.
- Models fail to download or load.
- The `server.log` shows permission errors or "Access denied".

**Cause:** Windows Defender or third-party antivirus software is blocking Lychee's executable or network access.

**Solution:**

**Add an exclusion in Windows Security:**
1. Open **Windows Security** → **Virus & threat protection**
2. Click **Manage settings** under "Virus & threat protection settings"
3. Scroll to **Exclusions** and click **Add or remove exclusions**
4. Add these folders:
   - `%LOCALAPPDATA%\Programs\Lychee`
   - `%HOMEPATH%\.lychee`
   - `%TEMP%` (Lychee writes temporary executables here)

**Or via PowerShell (as Administrator):**
```powershell
Add-MpPreference -ExclusionPath "$env:LOCALAPPDATA\Programs\Lychee"
Add-MpPreference -ExclusionPath "$env:USERPROFILE\.lychee"
Add-MpPreference -ExclusionProcess "lychee.exe"
Add-MpPreference -ExclusionProcess "lychee app.exe"
```

**Allow through the firewall:**
```powershell
New-NetFirewallRule -DisplayName "Lychee" -Direction Inbound -Program "$env:LOCALAPPDATA\Programs\Lychee\lychee.exe" -Action Allow
```

**Restart after adding exclusions:**
```powershell
lychee serve
```

If using third-party antivirus (McAfee, Norton, Bitdefender, etc.), consult its documentation to add folder and process exclusions for the same paths.

---

## Still stuck?

- **Logs:** See the full [troubleshooting guide](./troubleshooting.mdx) for log locations and debug logging.
- **GPU issues:** Check the [GPU documentation](./gpu.mdx).
- **Discord:** Join the [Lychee Discord](https://discord.gg/lychee) for community help.
- **GitHub Issues:** Search or file an issue at [github.com/lychee/lychee](https://github.com/lychee/lychee/issues).
