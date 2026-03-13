# Deploying a Public Devnet Node

This guide shows a conservative way to run a single public AnchorChain devnet node on an Ubuntu or Debian-like Linux VM.

This is for development and testing only. `anchorchaind devnet` is a non-production mode, and devnet state may be reset. Do not treat it as durable infrastructure.

## 1. Install prerequisites

Install the basic build and runtime tools:

```bash
sudo apt update
sudo apt install -y git make tmux curl
```

Install Go 1.22 or newer. The repo currently declares `go 1.22` in `go.mod`.

If your distro package is too old, install Go from an official tarball or package source appropriate for your environment before continuing.

## 2. Clone the repo

```bash
git clone <YOUR-ANCHORCHAIN-REPO-URL>
cd anchorchain
```

If you are working from a local copy instead of a remote clone, place that copy on the VM and `cd` into the project root.

## 3. Build the binaries

```bash
make build
```

This should create:

- `./bin/anchorchaind`
- `./bin/anchor-cli`

## 4. Run devnet

For a local-only bind on the VM:

```bash
./bin/anchorchaind devnet
```

By default the devnet API binds to `127.0.0.1:8081`, which is the safest starting point.

If you intentionally need the API to bind on a non-loopback address, the code clearly requires:

- `ANCHORCHAIN_API_ADDR=<host:port>`
- `ANCHORCHAIN_API_ALLOW_REMOTE=1`

Optional token auth is also clearly implemented through `ANCHORCHAIN_API_TOKEN`.

Example:

```bash
export ANCHORCHAIN_API_ADDR=0.0.0.0:8081
export ANCHORCHAIN_API_ALLOW_REMOTE=1
export ANCHORCHAIN_API_TOKEN=<TOKEN>
./bin/anchorchaind devnet
```

The daemon also prints the devnet EC/FCT addresses and a sample write command on startup.

## 5. Keep it running

### Option A: `tmux`

Start a session:

```bash
tmux new -s anchorchain-devnet
```

Run the daemon inside the session:

```bash
./bin/anchorchaind devnet
```

Detach with `Ctrl+B`, then `D`. Reattach later with:

```bash
tmux attach -t anchorchain-devnet
```

### Option B: `systemd`

If you want a simple service, create `/etc/systemd/system/anchorchain-devnet.service`:

```ini
[Unit]
Description=AnchorChain Devnet
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/anchorchain
Environment=ANCHORCHAIN_API_ADDR=127.0.0.1:8081
ExecStart=/opt/anchorchain/bin/anchorchaind devnet
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Then enable and start it:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now anchorchain-devnet
sudo systemctl status anchorchain-devnet
```

If you want public API exposure, change `ANCHORCHAIN_API_ADDR` and add `ANCHORCHAIN_API_ALLOW_REMOTE=1`. If you want token auth, also add `Environment=ANCHORCHAIN_API_TOKEN=<TOKEN>`.

## 6. Firewall and reverse proxy notes

The safest setup is:

- keep `anchorchaind` bound to `127.0.0.1:8081`
- expose it through a reverse proxy such as Nginx
- require TLS and, if needed, an API token

If you expose `0.0.0.0:8081` directly, restrict access with your VM firewall or cloud security group. On Ubuntu with UFW, for example:

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

I am not documenting a full reverse-proxy configuration here because this repo does not include one, and it is better to keep that choice environment-specific.

## 7. Health checks

From the VM:

```bash
curl -s http://127.0.0.1:8081/health
```

If the API is token-protected:

```bash
curl -s http://127.0.0.1:8081/health \
  -H "X-Anchorchain-Api-Token: <TOKEN>"
```

Using the CLI locally on the VM:

```bash
./bin/anchor-cli --api http://127.0.0.1:8081 node health
```

Against a remote API:

```bash
./bin/anchor-cli --api http://api.example.com:8081 --token <TOKEN> node health
```

## 8. Reset expectations

Devnet should be treated as disposable:

- chain and entry data may be lost or reset
- addresses printed on startup are for testing only
- operational changes should assume rebuilds and restarts are normal

If you need stable, long-lived behavior, you should validate that separately rather than assuming `devnet` provides it.
