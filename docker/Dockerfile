FROM golang:1.22-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
COPY node/factomd ./node/factomd
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/anchorchaind ./cmd/anchorchaind

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/anchorchaind /usr/local/bin/anchorchaind

ENV ANCHORCHAIN_API_ADDR=0.0.0.0:8081
ENV ANCHORCHAIN_API_ALLOW_REMOTE=1
ENV FACTOMD_RPC_ADDR=127.0.0.1:8088

VOLUME ["/root/.factom"]

EXPOSE 8081 8088 8090

ENTRYPOINT ["anchorchaind", "devnet"]
