.PHONY: build

build:
	mkdir -p bin
	go build -o bin/anchorchaind ./cmd/anchorchaind
	go build -o bin/anchor-cli ./cmd/anchor-cli
