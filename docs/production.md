# 🍒 Lychee in Production

A practical guide to deploying and operating Lychee in production environments. Every section includes copy-paste ready examples.

---

## Table of Contents

1. [System Requirements](#1-system-requirements)
2. [Running as a Service](#2-running-as-a-service)
3. [Docker Deployment](#3-docker-deployment)
4. [Security](#4-security)
5. [Monitoring](#5-monitoring)
6. [Scaling](#6-scaling)
7. [Backup](#7-backup)

---

## 1. System Requirements

### Minimum (1× 3B model, CPU-only)

| Resource | Requirement |
|:---|:---|
| **RAM** | 8 GB system + model memory (e.g., ~2 GB for Qwen2.5:3B Q4) |
| **Disk** | 20 GB free (models + logs) |
| **CPU** | 4 cores, x86-64 or ARM64 |
| **GPU** | None required (CPU fallback via llama.cpp) |

### Recommended (1–2× 8B models, GPU)

| Resource | Requirement |
|:---|:---|
| **RAM** | 32 GB system |
| **Disk** | 100 GB+ NVMe SSD |
| **CPU** | 8+ cores |
| **GPU** | NVIDIA RTX 3060+ (12 GB VRAM) or Apple M1+ (16 GB unified) |

### Production (multi-model, high concurrency)

| Resource | Requirement |
|:---|:---|
| **RAM** | 64 GB+ ECC |
| **Disk** | 500 GB+ NVMe SSD (RAID-1 recommended) |
| **CPU** | 16+ cores |
| **GPU** | NVIDIA A10 / A100 / RTX 4090 (24+ GB VRAM), or multi-GPU |
| **Network** | 1 Gbps+ for model downloads; low latency between load-balanced instances |

### GPU Acceleration Matrix

| Backend | NVIDIA | AMD | Apple Silicon | Intel |
|:---|:---:|:---:|:---:|:---:|
| **CUDA** | ✅ | ❌ | ❌ | ❌ |
| **ROCm** | ❌ | ✅ | ❌ | ❌ |
| **Metal (MLX)** | ❌ | ❌ | ✅ | ❌ |
| **Vulkan** | ✅ | ✅ | ❌ | ✅ |
| **CPU (AVX2)** | ✅ | ✅ | ✅ | ✅ |

> 💡 Lychee auto-detects the best available backend. Run `lychee scan` to see what your hardware supports and get model recommendations.

### Disk Space Planning

Model sizes (Q4_K_M quantization, approximate):

| Model Family | 3B | 7B/8B | 13B/14B | 32B/34B | 70B/72B |
|:---|---:|---:|---:|---:|---:|
| Llama 3.x | 2.0 GB | 4.9 GB | — | 20 GB | 40 GB |
| Qwen 2.5 | 1.9 GB | 4.9 GB | 8.9 GB | 19 GB | 41 GB |
| Gemma 3 | 2.2 GB | 5.5 GB | — | — | — |
| Phi-3/4 | 2.0 GB | 4.5 GB | 8.0 GB | — | — |

---

## 2. Running as a Service

### 2.1 Linux (systemd)

Create the service unit file:

```bash
sudo tee /etc/systemd/system/lychee.service << 'EOF'
[Unit]
Description=Lychee LLM Runtime
Documentation=https://github.com/MD-Mushfiqur123/lychee
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=lychee
Group=lychee
ExecStart=/usr/local/bin/lychee serve
Restart=always
RestartSec=5

# Environment
Environment=LYCHEE_HOST=127.0.0.1
Environment=LYCHEE_PORT=11434
Environment=LYCHEE_LOG_LEVEL=info
Environment=LYCHEE_KEEP_ALIVE=300
Environment=LYCHEE_MAX_LOADED_MODELS=3
Environment=LYCHEE_NUM_PARALLEL=4

# Security hardening
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/var/lib/lychee/models /var/log/lychee
PrivateTmp=yes
CapabilityBoundingSet=

# Resource limits
LimitNOFILE=65536
MemoryHigh=80%
CPUQuota=400%

[Install]
WantedBy=multi-user.target
EOF
```

Create the service user and directories:

```bash
sudo useradd --system --home-dir /var/lib/lychee --create-home lychee
sudo mkdir -p /var/lib/lychee/models /var/log/lychee
sudo chown -R lychee:lychee /var/lib/lychee /var/log/lychee
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now lychee
sudo systemctl status lychee
```

**View logs:**

```bash
sudo journalctl -u lychee -f        # follow
sudo journalctl -u lychee -n 100    # last 100 lines
sudo journalctl -u lychee --since "1 hour ago"
```

**Common operations:**

```bash
sudo systemctl stop lychee          # graceful stop
sudo systemctl restart lychee       # restart
sudo systemctl reload lychee        # config reload (if supported)
```

### 2.2 Windows Service

Lychee runs as a Windows Service using `sc.exe` or a wrapper like [NSSM](https://nssm.cc/) (Non-Sucking Service Manager).

#### Option A: Using NSSM (Recommended)

```powershell
# Download and install NSSM
Invoke-WebRequest -Uri "https://nssm.cc/release/nssm-2.24.zip" -OutFile "$env:TEMP\nssm.zip"
Expand-Archive "$env:TEMP\nssm.zip" -DestinationPath "$env:TEMP\nssm"
Copy-Item "$env:TEMP\nssm\nssm-2.24\win64\nssm.exe" "C:\Windows\System32\"

# Create the service
nssm install Lychee "C:\Program Files\lychee\lychee.exe" "serve"

# Configure environment
nssm set Lychee AppEnvironmentExtra "LYCHEE_HOST=127.0.0.1" "LYCHEE_PORT=11434" "LYCHEE_LOG_LEVEL=info"
nssm set Lychee AppStdout "C:\ProgramData\lychee\logs\stdout.log"
nssm set Lychee AppStderr "C:\ProgramData\lychee\logs\stderr.log"
nssm set Lychee AppRotateFiles 1
nssm set Lychee AppRotateSeconds 86400

# Start the service
nssm start Lychee
```

#### Option B: Using built-in `sc.exe`

```powershell
sc.exe create Lychee binPath= "\"C:\Program Files\lychee\lychee.exe\" serve" start= auto
sc.exe start Lychee
sc.exe query Lychee
```

To remove:

```powershell
sc.exe stop Lychee
sc.exe delete Lychee
```

### 2.3 macOS (launchd)

Create the plist file:

```bash
sudo tee /Library/LaunchDaemons/com.lychee.server.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.lychee.server</string>

    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/lychee</string>
        <string>serve</string>
    </array>

    <key>EnvironmentVariables</key>
    <dict>
        <key>LYCHEE_HOST</key>
        <string>127.0.0.1</string>
        <key>LYCHEE_PORT</key>
        <string>11434</string>
        <key>LYCHEE_LOG_LEVEL</key>
        <string>info</string>
    </dict>

    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/var/log/lychee/stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/lychee/stderr.log</string>

    <key>UserName</key>
    <string>_lychee</string>
    <key>GroupName</key>
    <string>_lychee</string>
</dict>
</plist>
EOF
```

Load and start:

```bash
sudo mkdir -p /var/log/lychee
sudo launchctl bootstrap system /Library/LaunchDaemons/com.lychee.server.plist
sudo launchctl kickstart system/com.lychee.server

# Check status
sudo launchctl print system/com.lychee.server

# Stop / unload
sudo launchctl bootout system/com.lychee.server
```

---

## 3. Docker Deployment

### 3.1 Docker Compose (Production)

Create `docker-compose.prod.yml`:

```yaml
services:
  lychee:
    image: ghcr.io/md-mushfiqur123/lychee:latest
    # Or pin a specific version:
    # image: ghcr.io/md-mushfiqur123/lychee:v0.1.0
    container_name: lychee

    ports:
      - "127.0.0.1:11434:11434"  # bind to localhost only; use reverse proxy for external access

    volumes:
      - lychee-models:/root/.lychee/models
      - ./lychee-config.yaml:/root/.lychee/config.yaml:ro  # optional config file

    environment:
      LYCHEE_HOST: "0.0.0.0"
      LYCHEE_PORT: "11434"
      LYCHEE_LOG_LEVEL: "info"
      LYCHEE_MAX_LOADED_MODELS: "3"
      LYCHEE_NUM_PARALLEL: "4"
      LYCHEE_KEEP_ALIVE: "300"
      LYCHEE_MAX_QUEUE: "100"

    restart: unless-stopped

    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:11434/api/tags"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

    # Resource limits
    deploy:
      resources:
        limits:
          memory: 32g
          cpus: "8"
        reservations:
          memory: 8g
          cpus: "2"

    logging:
      driver: "json-file"
      options:
        max-size: "50m"
        max-file: "5"

volumes:
  lychee-models:
    name: lychee-models
    driver: local
```

Deploy:

```bash
# Start
docker compose -f docker-compose.prod.yml up -d

# Check health
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs -f

# Pull a model
docker exec lychee lychee pull microsoft/Phi-3-mini-4k-instruct-gguf

# Stop gracefully (unloads models first)
docker compose -f docker-compose.prod.yml down
```

### 3.2 Docker with NVIDIA GPU

```yaml
services:
  lychee:
    image: ghcr.io/md-mushfiqur123/lychee:latest
    container_name: lychee
    runtime: nvidia
    environment:
      - NVIDIA_VISIBLE_DEVICES=all
      - NVIDIA_DRIVER_CAPABILITIES=compute,utility
    # ... rest of config same as above ...
```

Or using the newer `deploy` syntax:

```yaml
services:
  lychee:
    image: ghcr.io/md-mushfiqur123/lychee:latest
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
```

**Prerequisite — install NVIDIA Container Toolkit:**

```bash
# Ubuntu / Debian
curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey | \
    sudo gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg
curl -fsSL https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list | \
    sed 's#deb https://#deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://#g' | \
    sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list
sudo apt-get update
sudo apt-get install -y nvidia-container-toolkit
sudo nvidia-ctk runtime configure --runtime=docker
sudo systemctl restart docker

# Verify
docker run --rm --gpus all nvidia/cuda:12.0-base nvidia-smi
```

### 3.3 Docker with AMD GPU (ROCm)

```yaml
services:
  lychee:
    image: ghcr.io/md-mushfiqur123/lychee:latest
    devices:
      - /dev/kfd
      - /dev/dri
    group_add:
      - video
      - render
```

### 3.4 Docker with Vulkan (Generic GPU)

```yaml
services:
  lychee:
    image: ghcr.io/md-mushfiqur123/lychee:latest
    devices:
      - /dev/dri:/dev/dri
    environment:
      GGML_VK_VISIBLE_DEVICES: "0"   # select Vulkan device 0
      # LYCHEE_VULKAN: "0"           # disable Vulkan if needed
```

### 3.5 Docker Healthcheck & Readiness

Lychee exposes `/api/tags` for health checks:

```yaml
healthcheck:
  test: ["CMD-SHELL", "curl -f -s http://localhost:11434/api/tags || exit 1"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 15s
```

For orchestrators (Kubernetes, Nomad), use:

```yaml
# Readiness probe — server is up and responding
readinessProbe:
  httpGet:
    path: /api/tags
    port: 11434
  initialDelaySeconds: 10
  periodSeconds: 10

# Liveness probe — restart if unresponsive
livenessProbe:
  httpGet:
    path: /api/tags
    port: 11434
  initialDelaySeconds: 30
  periodSeconds: 30
  failureThreshold: 3
```

### 3.6 Volume Strategy

| Volume | Purpose | Backup? |
|:---|:---|:---:|
| `lychee-models` | Downloaded GGUF model files | ✅ Yes — large but critical |
| `lychee-config` | Configuration file(s) | ✅ Yes — tiny |
| `lychee-logs` | Application logs | ⚠️ Optional — rotate aggressively |

```yaml
volumes:
  lychee-models:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /data/lychee/models  # bind mount to SSD/NVMe
```

---

## 4. Security

### 4.1 Firewall

Limit access to the Lychee port. The server should NOT be exposed directly to the internet.

#### UFW (Ubuntu/Debian)

```bash
# Allow only from localhost and trusted internal IPs
sudo ufw allow from 127.0.0.1 to any port 11434
sudo ufw allow from 10.0.0.0/8 to any port 11434    # internal network
sudo ufw allow from 172.16.0.0/12 to any port 11434
# NEVER: sudo ufw allow 11434  -- this exposes to the world
sudo ufw enable
```

#### firewalld (RHEL/CentOS/Fedora)

```bash
sudo firewall-cmd --permanent --add-rich-rule='rule family="ipv4" source address="127.0.0.1" port port="11434" protocol="tcp" accept'
sudo firewall-cmd --permanent --add-rich-rule='rule family="ipv4" source address="10.0.0.0/8" port port="11434" protocol="tcp" accept'
sudo firewall-cmd --reload
```

#### iptables (bare metal)

```bash
sudo iptables -A INPUT -p tcp --dport 11434 -s 127.0.0.1 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 11434 -s 10.0.0.0/8 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 11434 -j DROP
sudo iptables-save > /etc/iptables/rules.v4
```

### 4.2 API Authentication

Lychee supports token-based authentication via environment variables.

```bash
# Set on server startup
export LYCHEE_API_KEY="sk-your-secret-token-here"
lychee serve
```

Or in docker-compose:

```yaml
environment:
  LYCHEE_API_KEY: "${LYCHEE_API_KEY}"  # inject from .env file
```

Clients must include the token:

```bash
# CLI
lychee run gemma3 "Hello" --api-key "sk-your-secret-token-here"

# OpenAI SDK
from openai import OpenAI
client = OpenAI(
    base_url="http://localhost:11434/v1",
    api_key="sk-your-secret-token-here"
)

# Plain HTTP
curl http://localhost:11434/v1/chat/completions \
  -H "Authorization: Bearer sk-your-secret-token-here" \
  -H "Content-Type: application/json" \
  -d '{"model":"gemma3","messages":[{"role":"user","content":"Hello"}]}'
```

### 4.3 HTTPS Reverse Proxy with nginx

Never expose Lychee directly to the internet. Put nginx in front.

```nginx
# /etc/nginx/sites-available/lychee
upstream lychee_backend {
    server 127.0.0.1:11434;
    # Add more instances for load balancing:
    # server 10.0.1.2:11434 weight=3;
    # server 10.0.1.3:11434 weight=2;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    server_name lychee-api.example.com;

    # SSL (Let's Encrypt)
    ssl_certificate     /etc/letsencrypt/live/lychee-api.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/lychee-api.example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Client upload size (for large prompts, image attachments)
    client_max_body_size 100m;

    # Timeouts — LLM inference can be slow
    proxy_connect_timeout 60s;
    proxy_read_timeout 300s;      # 5 min for long generations
    proxy_send_timeout 300s;

    # Headers
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # Buffering — stream responses for real-time token generation
    proxy_buffering off;
    proxy_cache off;

    # Rate limiting (optional)
    limit_req_zone $binary_remote_addr zone=lychee_api:10m rate=10r/s;
    limit_req zone=lychee_api burst=20 nodelay;

    # Strip /lychee prefix if using path-based routing
    location /lychee/ {
        rewrite ^/lychee/(.*) /$1 break;
        proxy_pass http://lychee_backend;
    }

    # Direct subdomain proxy
    location / {
        proxy_pass http://lychee_backend;
    }

    # Block direct access to internal endpoints (optional)
    location ~ ^/(api/tags|api/ps) {
        allow 10.0.0.0/8;
        allow 127.0.0.1;
        deny all;
        proxy_pass http://lychee_backend;
    }
}

# Redirect HTTP → HTTPS
server {
    listen 80;
    server_name lychee-api.example.com;
    return 301 https://$host$request_uri;
}
```

Enable the site:

```bash
sudo ln -s /etc/nginx/sites-available/lychee /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# Generate Let's Encrypt cert
sudo certbot --nginx -d lychee-api.example.com
```

### 4.4 CORS Configuration

If your frontend calls Lychee directly (dev/staging only):

```bash
export LYCHEE_CORS_ORIGINS="https://your-frontend.example.com,https://localhost:3000"
lychee serve
```

> ⚠️ In production, route all calls through your backend to avoid exposing Lychee's port.

### 4.5 Security Checklist

- [ ] Lychee bound to `127.0.0.1` (not `0.0.0.0`) unless behind a reverse proxy
- [ ] Firewall blocks port 11434 from public internet
- [ ] API key configured (`LYCHEE_API_KEY`)
- [ ] HTTPS enforced at the reverse proxy
- [ ] nginx rate limiting active
- [ ] Regular updates (`docker compose pull` / binary upgrades)
- [ ] File system permissions: `lychee` user owns model directory, no world-writable files
- [ ] Models verified with SHA256 (Lychee does this automatically on pull)

---

## 5. Monitoring

### 5.1 Health Endpoint

```bash
# Basic health — server is up
curl http://localhost:11434/api/tags
# → {"models":[...]}

# Check specific model availability
curl http://localhost:11434/api/show -d '{"name":"gemma3"}'

# List running models
curl http://localhost:11434/api/ps
```

**Prometheus-style health in your monitoring:**

```bash
#!/bin/bash
# check_lychee.sh — Nagios / Sensu / cron health check
HEALTH=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:11434/api/tags)
if [ "$HEALTH" != "200" ]; then
    echo "CRITICAL: Lychee returned HTTP $HEALTH"
    exit 2
fi
echo "OK: Lychee is healthy"
exit 0
```

**Docker healthcheck integration:**

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:11434/api/tags"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 10s
```

### 5.2 Logs

Lychee writes structured logs to stdout/stderr.

**Log levels:**

```bash
export LYCHEE_LOG_LEVEL=debug    # verbose — model loading, request details
export LYCHEE_LOG_LEVEL=info     # default — requests, model lifecycle
export LYCHEE_LOG_LEVEL=warn     # warnings only
export LYCHEE_LOG_LEVEL=error    # errors only
```

**systemd journal (Linux):**

```bash
journalctl -u lychee -f                         # follow
journalctl -u lychee --since "10 minutes ago"   # recent
journalctl -u lychee -p err                     # errors only
journalctl -u lychee -o json                    # machine-readable
```

**Docker logs:**

```bash
docker compose logs -f lychee                   # follow
docker compose logs --tail=200 lychee           # last 200 lines
docker compose logs --since 10m lychee          # last 10 minutes
```

**Log rotation with docker-compose:**

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "50m"
    max-file: "5"
```

### 5.3 Metrics with `lychee stats`

```bash
# Snapshot
lychee stats

# Interactive TUI dashboard (real-time)
lychee stats --tui

# Machine-readable (JSON)
lychee stats --json
```

### 5.4 Prometheus + Grafana (Advanced)

If you want metrics in your existing monitoring stack, wrap the health endpoint in a small exporter:

```python
# lychee_exporter.py — minimal Prometheus exporter
from prometheus_client import start_http_server, Gauge, Counter
import requests
import time
import json

MODEL_COUNT = Gauge('lychee_models_total', 'Number of downloaded models')
REQUEST_COUNT = Counter('lychee_requests_total', 'Total inference requests', ['model'])
UP = Gauge('lychee_up', 'Lychee server health (1=up, 0=down)')

LYCHEE_URL = "http://localhost:11434"

def collect():
    try:
        resp = requests.get(f"{LYCHEE_URL}/api/tags", timeout=5)
        if resp.status_code == 200:
            UP.set(1)
            models = resp.json().get("models", [])
            MODEL_COUNT.set(len(models))
            return models
        else:
            UP.set(0)
    except Exception:
        UP.set(0)
    return []

if __name__ == "__main__":
    start_http_server(9090)
    while True:
        collect()
        time.sleep(30)
```

Run it:

```bash
pip install prometheus-client requests
python lychee_exporter.py &
# Metrics at http://localhost:9090/metrics
```

Then point Prometheus at `localhost:9090` and build dashboards in Grafana.

### 5.5 Alerting Rules

Key things to alert on:

| Condition | Check | Threshold |
|:---|:---|:---|
| Server down | `curl /api/tags` fails | Immediately |
| High latency | Response time > percentile | p95 > 30s |
| Model load failure | Check logs for "failed to load" | Any occurrence |
| Disk space low | `df /var/lib/lychee/models` | < 20% free |
| Memory pressure | Container/system OOM events | Any occurrence |

**Simple cron-based alert:**

```bash
# /etc/cron.d/lychee-health
*/5 * * * * root /usr/local/bin/check_lychee.sh || echo "Lychee health check failed" | mail -s "Lychee Alert" ops@example.com
```

---

## 6. Scaling

### 6.1 Built-in Load Balancing Router

Lychee has a built-in weighted round-robin router. Register multiple backends:

```bash
# Register external lychee instances as routes
curl -X POST http://localhost:11434/api/routes \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production",
    "endpoints": [
      {"url": "http://10.0.1.10:11434", "weight": 3},
      {"url": "http://10.0.1.11:11434", "weight": 3},
      {"url": "http://10.0.1.12:11434", "weight": 2}
    ],
    "strategy": "weighted_round_robin"
  }'
```

Then route inference through the router:

```bash
curl http://localhost:11434/api/chat \
  -H "Content-Type: application/json" \
  -d '{"route":"production","model":"gemma3","messages":[{"role":"user","content":"Hello"}]}'
```

### 6.2 nginx Load Balancing

```nginx
upstream lychee_cluster {
    least_conn;  # send to instance with fewest active connections

    server 10.0.1.10:11434 weight=3 max_fails=3 fail_timeout=30s;
    server 10.0.1.11:11434 weight=3 max_fails=3 fail_timeout=30s;
    server 10.0.1.12:11434 weight=2 max_fails=3 fail_timeout=30s;

    keepalive 64;
}

server {
    listen 443 ssl http2;
    server_name lychee-api.example.com;

    # SSL config ...
    # (see Section 4.3 for full SSL config)

    location / {
        proxy_pass http://lychee_cluster;

        # Critical for long-running generations
        proxy_read_timeout 600s;
        proxy_buffering off;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }
}
```

### 6.3 Model Distribution Strategy

Instead of duplicating models on every node, distribute by capability:

```
┌─────────────────────────────────────────────────────────────┐
│                     nginx / HAProxy                          │
│                  (route by model name)                       │
└───────┬──────────────────┬──────────────────┬───────────────┘
        │                  │                  │
   ┌────▼─────┐      ┌────▼─────┐      ┌────▼─────┐
   │ Node A   │      │ Node B   │      │ Node C   │
   │ GPU: A100│      │ GPU: 4090│      │ GPU: 4090│
   │           │      │           │      │           │
   │ llama3.1  │      │ gemma3   │      │ qwen2.5  │
   │ 70B       │      │ 27B      │      │ 32B       │
   │ phi4      │      │ mistral   │      │ codellama │
   └───────────┘      └───────────┘      └───────────┘
```

**nginx map routing by model:**

```nginx
map $http_x_model $lychee_backend {
    "~*llama-3"     "backend_llama";
    "~*gemma"       "backend_gemma";
    "~*qwen"        "backend_qwen";
    "~*phi"         "backend_llama";
    default         "backend_default";
}

upstream backend_llama { server 10.0.1.10:11434; }
upstream backend_gemma { server 10.0.1.11:11434; }
upstream backend_qwen  { server 10.0.1.12:11434; }
upstream backend_default { server 10.0.1.10:11434; }

server {
    # ...
    location / {
        proxy_pass http://$lychee_backend;
    }
}
```

> Clients set the `X-Model` header. If they do not, requests fall through to the default backend.

### 6.4 Capacity Planning

| Concurrency | Instances | Notes |
|---:|---:|:---|
| 1–5 req/s | 1 | Single instance, 1–2 models in memory |
| 5–20 req/s | 2–3 | Load balance, keep 2+ models warm |
| 20–50 req/s | 4–6 | Each GPU handles 1–2 models, queue overflow handling |
| 50+ req/s | 6+ | Horizontal scaling with model sharding by type |

**Queue depth protection:**

```bash
export LYCHEE_MAX_QUEUE=100   # reject requests beyond this
export LYCHEE_NUM_PARALLEL=4  # max concurrent inferences per model
```

When queue fills, clients get `503 Service Unavailable` — they should retry with backoff.

### 6.5 Kubernetes (Minimal)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lychee
spec:
  replicas: 3
  selector:
    matchLabels:
      app: lychee
  template:
    metadata:
      labels:
        app: lychee
    spec:
      containers:
        - name: lychee
          image: ghcr.io/md-mushfiqur123/lychee:latest
          ports:
            - containerPort: 11434
          env:
            - name: LYCHEE_HOST
              value: "0.0.0.0"
            - name: LYCHEE_PORT
              value: "11434"
          volumeMounts:
            - name: models
              mountPath: /root/.lychee/models
          resources:
            requests:
              memory: "8Gi"
              cpu: "2"
            limits:
              memory: "32Gi"
              cpu: "8"
          readinessProbe:
            httpGet:
              path: /api/tags
              port: 11434
            initialDelaySeconds: 10
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /api/tags
              port: 11434
            initialDelaySeconds: 30
            periodSeconds: 30
      volumes:
        - name: models
          persistentVolumeClaim:
            claimName: lychee-models-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: lychee
spec:
  selector:
    app: lychee
  ports:
    - port: 11434
      targetPort: 11434
  type: ClusterIP
```

---

## 7. Backup

### 7.1 What to Back Up

| Path | Contents | Size | Priority |
|:---|:---|:---|:---|
| `~/.lychee/models/` | Downloaded GGUF files | 2–200+ GB | **Critical** — models can be re-downloaded but slowly |
| `~/.lychee/config.yaml` | Server configuration | < 5 KB | **Critical** — trivial to back up |
| `~/.lychee/conversations/` | Conversation history (SQLite) | < 100 MB | **Important** — user data |
| `/var/log/lychee/` | Application logs | Variable | Optional — rotate and age out |

### 7.2 Backup Script

```bash
#!/bin/bash
# backup-lychee.sh — Rsync-based backup
set -euo pipefail

BACKUP_DIR="/backups/lychee"
LYCHEE_DATA="/var/lib/lychee"
TIMESTAMP=$(date +%Y-%m-%d_%H%M%S)
BACKUP_PATH="${BACKUP_DIR}/lychee-${TIMESTAMP}"

mkdir -p "$BACKUP_DIR"

# Stop lychee briefly for model consistency (optional — models are read-only after download)
# systemctl stop lychee || true

echo "Backing up lychee to $BACKUP_PATH ..."

# Models (snapshot with hard links to save space on incremental backups)
rsync -ah --link-dest="${BACKUP_DIR}/latest" \
    "${LYCHEE_DATA}/models/" \
    "${BACKUP_PATH}/models/"

# Config & conversations
rsync -ah \
    "${LYCHEE_DATA}/config.yaml" \
    "${LYCHEE_DATA}/conversations/" \
    "${BACKUP_PATH}/"

# Update "latest" symlink
rm -f "${BACKUP_DIR}/latest"
ln -s "${BACKUP_PATH}" "${BACKUP_DIR}/latest"

# Start lychee again
# systemctl start lychee || true

# Cleanup: keep last 7 backups
find "$BACKUP_DIR" -maxdepth 1 -type d -name "lychee-*" | sort | head -n -7 | xargs rm -rf

echo "Backup complete: $BACKUP_PATH"
echo "Size: $(du -sh "$BACKUP_PATH" | cut -f1)"
```

**Cron it:**

```bash
# /etc/cron.d/lychee-backup
0 3 * * * lychee /usr/local/bin/backup-lychee.sh >> /var/log/lychee/backup.log 2>&1
```

### 7.3 Docker Volume Backup

```bash
#!/bin/bash
# docker-backup-lychee.sh
set -euo pipefail

BACKUP_DIR="/backups/lychee"
TIMESTAMP=$(date +%Y-%m-%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/lychee-models-${TIMESTAMP}.tar.gz"

mkdir -p "$BACKUP_DIR"

# Create a temporary container that mounts the volume and backs it up
docker run --rm \
    -v lychee-models:/data:ro \
    -v "${BACKUP_DIR}:/backup" \
    alpine:latest \
    tar czf "/backup/$(basename "$BACKUP_FILE")" -C /data .

echo "Backup saved: $BACKUP_FILE ($(du -sh "$BACKUP_FILE" | cut -f1))"

# Keep last 5 backups
ls -t "${BACKUP_DIR}"/lychee-models-*.tar.gz | tail -n +6 | xargs rm -f
```

### 7.4 Restore

```bash
# Filesystem restore
systemctl stop lychee
cp -r /backups/lychee/latest/models/* /var/lib/lychee/models/
cp /backups/lychee/latest/config.yaml /var/lib/lychee/config.yaml
chown -R lychee:lychee /var/lib/lychee
systemctl start lychee
```

```bash
# Docker volume restore
docker run --rm \
    -v lychee-models:/data \
    -v /backups/lychee:/backup:ro \
    alpine:latest \
    tar xzf /backup/lychee-models-2025-01-01_030000.tar.gz -C /data
```

### 7.5 Cloud Backup (S3-compatible)

```bash
#!/bin/bash
# sync-models-to-s3.sh
set -euo pipefail

S3_BUCKET="s3://my-lychee-backups/models"
MODELS_DIR="/var/lib/lychee/models"

# Sync models to S3 (awscli required)
aws s3 sync "$MODELS_DIR" "$S3_BUCKET" \
    --storage-class STANDARD_IA \
    --exclude "*.tmp" \
    --exclude ".locks/*" \
    --delete

echo "Synced to $S3_BUCKET"
```

### 7.6 Disaster Recovery Runbook

1. **Server lost →** Provision new instance, install Lychee (`curl | sh`), restore config and models from backup
2. **Models corrupted →** Verify with `lychee list`, re-pull corrupted models with `lychee pull <model>`
3. **Docker volume lost →** `docker compose down && docker volume rm lychee-models && docker compose up -d`, then re-pull models
4. **Configuration lost →** Restore `config.yaml` from backup or set environment variables
5. **Disk full →** Remove unused models (`lychee remove <model>`), expand volume, restore from backup

---

## Appendix: Full Environment Variable Reference

| Variable | Default | Description |
|:---|:---|:---|
| `LYCHEE_HOST` | `127.0.0.1` | Server bind address |
| `LYCHEE_PORT` | `11434` | Server port |
| `LYCHEE_API_KEY` | (none) | API authentication token |
| `LYCHEE_LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `LYCHEE_MAX_LOADED_MODELS` | `0` (unlimited) | Max models kept in RAM |
| `LYCHEE_NUM_PARALLEL` | `0` (auto) | Max concurrent requests per model |
| `LYCHEE_MAX_QUEUE` | `0` (unlimited) | Max pending request queue size |
| `LYCHEE_KEEP_ALIVE` | `300` | Seconds to keep model in memory after last use |
| `LYCHEE_VULKAN` | `1` | Enable/disable Vulkan backend (`1` or `0`) |
| `LYCHEE_CORS_ORIGINS` | (none) | Comma-separated allowed CORS origins |
| `GGML_VK_VISIBLE_DEVICES` | (all) | Select Vulkan devices by index |

---

> 🍒 **Production is a practice, not a destination.** Start with the basics (firewall, systemd, health check), add HTTPS, then scale as your workload grows. Questions? Open a [GitHub Discussion](https://github.com/MD-Mushfiqur123/lychee/discussions).
