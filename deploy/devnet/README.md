# Public Devnet Deployment

This directory provides a minimal Docker Compose deployment for running a single public AnchorChain devnet node on a Linux VM.

This is for development and testing only. `anchorchaind devnet` is not production infrastructure.

## What this setup does

- builds `anchorchaind` from the local repository
- runs `anchorchaind devnet`
- publishes the AnchorChain HTTP API on `8081`
- keeps factomd RPC on `8088` bound to `127.0.0.1` by default
- keeps the control panel on `8090` bound to `127.0.0.1` by default
- persists `.factom` state in a Docker volume
- restarts automatically with `restart: unless-stopped`

## Why `8088` is not public by default

For a public devnet node, exposing `8081` is the useful default.

`8088` is the legacy factomd RPC interface. It is safer to keep it host-local unless you have a very specific need to expose it and you understand the implications. The included compose file therefore binds:

- `8081` to all interfaces
- `8088` to `127.0.0.1`
- `8090` to `127.0.0.1`

If you intentionally want public `8088`, change:

```yaml
- "127.0.0.1:8088:8088"
```

to:

```yaml
- "8088:8088"
```

Do the same for `8090` only if you really want the control panel reachable off-host.

## Linux VM guidance

Recommended baseline:

- Ubuntu or Debian VM
- Docker Engine with the Compose plugin installed
- a persistent disk attached to the VM
- a firewall that allows only the ports you intend to expose

## Start the public devnet node

From the repository root:

```bash
docker compose -f deploy/devnet/docker-compose.yml up --build -d
```

Or from this directory:

```bash
docker compose up --build -d
```

## Stop the node

From the repository root:

```bash
docker compose -f deploy/devnet/docker-compose.yml down
```

## View logs

```bash
docker compose -f deploy/devnet/docker-compose.yml logs -f
```

## Persistence

Chain data is stored in the named Docker volume mounted at `/root/.factom`.

To inspect volumes:

```bash
docker volume ls
```

To fully reset devnet state:

```bash
docker compose -f deploy/devnet/docker-compose.yml down -v
```

## Firewall / exposure guidance

For a public devnet node, allow inbound:

- `8081/tcp` if you want the public HTTP API reachable

Keep these private unless you intentionally need them:

- `8088/tcp` factomd RPC
- `8090/tcp` control panel

Example UFW commands on Ubuntu:

```bash
sudo ufw allow OpenSSH
sudo ufw allow 8081/tcp
sudo ufw deny 8088/tcp
sudo ufw deny 8090/tcp
sudo ufw enable
```

If your cloud provider also has a security group or network ACL, match the same policy there.

## Optional API token

If you want to require a token for the public HTTP API, uncomment `ANCHORCHAIN_API_TOKEN` in `docker-compose.yml` and set a real value.

Then clients can call:

```bash
curl -H "X-Anchorchain-Api-Token: <TOKEN>" http://YOUR_HOST:8081/health
```

## Health checks

From the VM itself:

```bash
curl http://127.0.0.1:8081/health
```

From another machine on the internet:

```bash
curl http://YOUR_HOST_OR_IP:8081/health
```

If token-protected:

```bash
curl -H "X-Anchorchain-Api-Token: <TOKEN>" http://YOUR_HOST_OR_IP:8081/health
```

You can also inspect the running container:

```bash
docker compose -f deploy/devnet/docker-compose.yml ps
```

## Operator notes

- `restart: unless-stopped` handles basic reboot/restart behavior
- devnet remains single-node and non-production
- use `docker compose logs -f` to watch startup and write/read behavior
