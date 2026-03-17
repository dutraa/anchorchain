# AnchorChain

![License](https://img.shields.io/badge/license-MIT-blue)
![Status](https://img.shields.io/badge/status-hackathon%20ready-green)
![Built With](https://img.shields.io/badge/built%20with-Factom-orange)

---

## 🚀 AnchorChain

> **Turn any data into permanent, verifiable truth — in seconds.**

AnchorChain is a developer-first platform that makes it effortless to anchor data to blockchain using the Factom protocol.

With a single API call, you can:

- 🔒 Make data tamper-proof
- ⛓️ Anchor it to blockchain
- 🧾 Generate cryptographic proof
- 🔍 Verify it anytime, anywhere

---

## ⚡ The Problem

Today’s systems rely on **mutable databases**.

That means:

- Logs can be altered  
- Records can be deleted  
- Audits require trust  
- Compliance is expensive  

Even blockchain solutions are often:

- Too complex  
- Too slow  
- Too expensive for simple proofs  

---

## 💡 The Solution

AnchorChain provides a **simple, fast, and scalable way to create immutable records**.

Instead of building blockchain infrastructure yourself, you just:

```bash
POST /chains/{id}/entries
```

And your data is:

- hashed
- anchored to Factom
- secured via Bitcoin
- verifiable forever

---

## 🧠 Why Factom?

Factom is uniquely suited for data anchoring:

- ⚡ High throughput
- 💰 Low cost per entry
- 🔗 Anchored into Bitcoin for security
- 📦 Built specifically for data integrity

AnchorChain brings this power to modern developers.

---

## 🏗️ Architecture

```
Your App
   │
   ▼
AnchorChain API
   │
   ▼
Factom Network
   │
   ▼
Bitcoin (Final Anchor)
```

---

## 🛠️ What You Can Build

AnchorChain unlocks entirely new categories of applications:

### 📜 Audit & Compliance
- Immutable logs
- Regulatory proof systems
- Financial reporting integrity

### 🚚 Supply Chain
- Product provenance
- Shipment verification
- Anti-counterfeiting

### ⚖️ Legal Tech
- Evidence timestamping
- Contract notarization
- Chain-of-custody tracking

### 🧠 AI + Data Integrity
- Training data verification
- Model output proofs
- Dataset lineage tracking

---

## ⚡ Quick Demo (60 Seconds)

### 1. Create a Chain

```bash
curl -X POST http://localhost:8080/chains
```

---

### 2. Write Data

```bash
curl -X POST http://localhost:8080/chains/<CHAIN_ID>/entries \
  -H "Content-Type: application/json" \
  -d '{"event":"contract_signed","user":"alice"}'
```

---

### 3. Verify It

```bash
curl http://localhost:8080/receipts/<ENTRY_HASH>
```

---

## 🔥 Key Features

- ⚡ Simple REST API
- 🧾 Cryptographic receipt verification
- 🧰 CLI for developers
- 🧪 Local devnet (no external dependencies)
- 🌐 Web explorer
- 📦 Structured JSON support
- 🔐 Bitcoin-level security (via Factom anchoring)

---

## 🚀 Getting Started

### Clone

```bash
git clone https://github.com/anchorchain/anchorchain.git
cd anchorchain
```

---

### Run Devnet

```bash
docker compose up
```

---

### Start API

```bash
go run cmd/server/main.go
```

---

### Build CLI

```bash
go build -o anchorchain cmd/cli/main.go
```

---

## 🧪 Example CLI Flow

```bash
# Create chain
./anchorchain chain create

# Write entry
./anchorchain entry write \
  --chain <CHAIN_ID> \
  --data '{"status":"verified"}'

# Fetch entries
./anchorchain chain entries <CHAIN_ID>
```

---

## 🌐 Explorer

Visualize your data:

```
http://localhost:3000
```

Features:

- Browse chains
- View entries
- Inspect hashes
- Verify receipts

---

## 🧩 Project Structure

```
anchorchain/
├── cmd/
│   ├── server/
│   └── cli/
├── internal/
│   ├── api/
│   ├── chains/
│   ├── receipts/
│   └── factom/
├── explorer/
├── devnet/
├── docker/
└── README.md
```

---

## 🧭 Roadmap

### Phase 1 (Now)
- API + CLI
- Local devnet
- Receipt verification

### Phase 2
- SDKs (JS / Python / Go)
- Batch anchoring
- Auth & API keys

### Phase 3
- Hosted AnchorChain nodes
- Enterprise integrations
- Cross-chain proofs

---

## 🏆 Why AnchorChain Wins

- 🧠 Uses **Factom (underutilized, powerful tech)**
- ⚡ Focuses on **real-world use cases**
- 🧑‍💻 Built for **developers first**
- 🔐 Provides **true cryptographic guarantees**
- 🚀 Ready for **immediate deployment**

---

## 🤝 Contributing

We welcome contributors.

Open issues, submit PRs, or build something on top of AnchorChain.

---

## 📜 License

MIT

---

## 🔥 AnchorChain

> **Data you can prove. Trust you don’t have to assume.**
