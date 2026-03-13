package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/engine"
	factomstate "github.com/FactomProject/factomd/state"
	"github.com/anchorchain/anchorchain/api"
	walletcli "github.com/anchorchain/anchorchain/wallet/cli"
)

const (
	devnetFCTAddress = "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q"
	devnetFCTSecret  = "Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK"
	devnetECSecret   = "Es3LB2YW9bpdWmMnNQYb31kyPzqnecsNqmg5W4K7FKp4UP6omRTa"
	devnetInitialECCredits = 1000
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "wallet":
			if err := walletcli.Run(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "wallet command error: %v\n", err)
				os.Exit(1)
			}
			return
		case "devnet":
			runDevnet()
			return
		}
	}

	logger := log.New(os.Stdout, "[anchorchaind] ", log.LstdFlags)
	apiAddr, factomdRPC, apiOpts := prepareAPI(logger, false)
	api.Start(apiAddr, factomdRPC, logger, apiOpts)

	fmt.Println("Command Line Arguments:")

	for _, v := range os.Args[1:] {
		fmt.Printf("\t%s\n", v)
	}

	params := engine.ParseCmdLine(os.Args[1:])
	params.PrettyPrint()

	state := engine.Factomd(params)
	for state.GetRunState() != runstate.Stopped {
		time.Sleep(time.Second)
	}
	fmt.Println("Waiting to Shut Down")
	time.Sleep(time.Second * 5)
}

func runDevnet() {
	logger := log.New(os.Stdout, "[anchorchaind] ", log.LstdFlags)
	apiAddr, factomdRPC, apiOpts := prepareAPI(logger, true)
	api.Start(apiAddr, factomdRPC, logger, apiOpts)
	devnetECAddress := mustDevnetECAddress()

	apiURL := fmt.Sprintf("http://%s", apiAddr)
	sampleBody := `{"extIds":["devnet","chain"],"content":"{\"demo\":true}","contentEncoding":"utf-8"}`
	sampleCurl := fmt.Sprintf("curl -s -X POST %s/chains -H 'Content-Type: application/json' -d '%s'", apiURL, sampleBody)

	fmt.Println("!!! DEVNET MODE - NON-PRODUCTION ONLY !!!")
	fmt.Println("──────── AnchorChain Devnet ────────")
	fmt.Printf("HTTP API: %s\n", apiURL)
	fmt.Printf("Legacy RPC: %s\n", factomdRPC)
	fmt.Println("Block time: 60s")
	fmt.Printf("Devnet EC Address: %s\n", devnetECAddress)
	fmt.Printf("Devnet EC Secret: %s\n", devnetECSecret)
	fmt.Printf("Devnet FCT Address: %s\n", devnetFCTAddress)
	fmt.Printf("Devnet FCT Secret: %s\n", devnetFCTSecret)
	fmt.Printf("Startup EC Funding: %d credits to the devnet EC address above\n", devnetInitialECCredits)
	fmt.Println("Sample write:")
	fmt.Printf("  %s\n", sampleCurl)

	args := []string{"-network=CUSTOM", "-customnet=devnet-local", "-exclusive=true", "-blktime=60", "-prefix=devnet"}
	params := engine.ParseCmdLine(args)
	params.PrettyPrint()

	state := engine.Factomd(params)
	if st, ok := state.(*factomstate.State); ok {
		fundDevnetEC(logger, st, devnetECAddress)
	}
	for state.GetRunState() != runstate.Stopped {
		time.Sleep(time.Second)
	}
	fmt.Println("Devnet shutting down")
	time.Sleep(2 * time.Second)
}

func envOrDefault(fallback string, keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return fallback
}

func prepareAPI(logger *log.Logger, devMode bool) (string, string, api.Options) {
	apiAddr := envOrDefault("127.0.0.1:8081", "ANCHORCHAIN_API_ADDR", "FACTOM_NEXT_API_ADDR")
	factomdRPC := envOrDefault("localhost:8088", "FACTOMD_RPC_ADDR")
	allowRemote := boolEnv("ANCHORCHAIN_API_ALLOW_REMOTE", "FACTOM_NEXT_API_ALLOW_REMOTE")

	enforceBindSafety(logger, apiAddr, allowRemote)

	token := envOrDefault("", "ANCHORCHAIN_API_TOKEN", "FACTOM_NEXT_API_TOKEN")
	opts := api.Options{
		DevnetMode: devMode,
		AuthToken:  token,
	}
	if devMode {
		opts.DevnetECPrivateKey = devnetECSecret
	}

	return apiAddr, factomdRPC, opts
}

func mustDevnetECAddress() string {
	ecAddr, err := factom.GetECAddress(devnetECSecret)
	if err != nil {
		panic(fmt.Sprintf("invalid devnet EC secret: %v", err))
	}
	return ecAddr.String()
}

func fundDevnetEC(logger *log.Logger, st *factomstate.State, ecAddress string) {
	priceDeadline := time.Now().Add(30 * time.Second)
	for st.GetFactoshisPerEC() == 0 && time.Now().Before(priceDeadline) {
		time.Sleep(500 * time.Millisecond)
	}
	if st.GetFactoshisPerEC() == 0 {
		logger.Printf("[devnet] WARNING: factomd EC price is still zero; skipping automatic EC funding for %s", ecAddress)
		return
	}

	logger.Printf("[devnet] Funding default EC address %s with %d entry credits", ecAddress, devnetInitialECCredits)
	err, txid := engine.FundECWalletFromBank(st, ecAddress, devnetInitialECCredits)
	if err != nil {
		logger.Printf("[devnet] Failed to submit EC funding transaction for %s: %v", ecAddress, err)
		return
	}
	logger.Printf("[devnet] Submitted EC funding transaction %s for %s", txid, ecAddress)

	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		balance := engine.GetBalanceEC(st, ecAddress)
		if balance > 0 {
			logger.Printf("[devnet] Devnet EC address %s now has %d entry credits", ecAddress, balance)
			return
		}
		time.Sleep(2 * time.Second)
	}

	logger.Printf("[devnet] WARNING: funding transaction for %s was submitted but EC balance is still zero after waiting", ecAddress)
}

func enforceBindSafety(logger *log.Logger, addr string, allowRemote bool) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		logger.Fatalf("invalid ANCHORCHAIN_API_ADDR (or FACTOM_NEXT_API_ADDR) %q: %v", addr, err)
	}
	if host == "" {
		host = "0.0.0.0"
	}
	if isLoopbackHost(host) {
		logger.Printf("[HTTP] Binding HTTP API to %s (loopback only)", addr)
		return
	}
	if allowRemote {
		logger.Printf("[SECURITY] Remote HTTP API binding enabled on %s (ANCHORCHAIN_API_ALLOW_REMOTE acknowledged).", addr)
		return
	}
	logger.Printf("[SECURITY] WARNING: HTTP API is binding to %s without ANCHORCHAIN_API_ALLOW_REMOTE set.", addr)
	logger.Printf("[SECURITY] Set ANCHORCHAIN_API_ALLOW_REMOTE=1 (or FACTOM_NEXT_API_ALLOW_REMOTE) and secure the host to proceed intentionally.")
}

func boolEnv(keys ...string) bool {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			switch strings.ToLower(v) {
			case "1", "true", "yes", "on":
				return true
			}
			return false
		}
	}
	return false
}

func isLoopbackHost(host string) bool {
	if host == "" {
		return false
	}
	switch strings.ToLower(host) {
	case "localhost":
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback()
}
