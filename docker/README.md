# Docker Devnet

This directory provides a minimal Docker-based development environment for running a local AnchorChain devnet with one command.

## What it does

- builds `anchorchaind` from the local repository
- runs `anchorchaind devnet` as the container entrypoint
- exposes:
  - `8081` for the AnchorChain HTTP API
  - `8088` for Factom RPC
  - `8090` for the optional control panel
- persists `.factom` state in a named Docker volume

## Files

- `Dockerfile` builds the Go daemon and packages a small runtime image
- `docker-compose.yml` builds and runs the devnet container

## Start devnet

From the repository root:

```bash
docker compose -f docker/docker-compose.yml up --build
```

Or from this directory:

```bash
docker compose up --build
```

## Stop devnet

```bash
docker compose -f docker/docker-compose.yml down
```

If you started from this directory, `docker compose down` also works.

## Reset persisted devnet state

```bash
docker compose -f docker/docker-compose.yml down -v
```

## Endpoints

- AnchorChain API: [http://127.0.0.1:8081](http://127.0.0.1:8081)
- Factom RPC: `127.0.0.1:8088`
- Control panel: [http://127.0.0.1:8090](http://127.0.0.1:8090)

## Notes

- The container sets `ANCHORCHAIN_API_ADDR=0.0.0.0:8081` and `ANCHORCHAIN_API_ALLOW_REMOTE=1` so the HTTP API is reachable from the host.
- Factom RPC stays on `127.0.0.1:8088` inside the container because the API and daemon run in the same process.
