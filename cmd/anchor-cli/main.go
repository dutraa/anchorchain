package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := parseRootFlags(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	client, err := newAPIClient(cfg.baseURL, cfg.token)
	if err != nil {
		return err
	}
	app := &cli{client: client, jsonOutput: cfg.json}
	return app.dispatch(cfg.args)
}

type rootConfig struct {
	baseURL string
	token   string
	json    bool
	args    []string
}

func parseRootFlags(args []string) (rootConfig, error) {
	root := flag.NewFlagSet("anchor-cli", flag.ContinueOnError)
	root.Usage = printMainUsage

	apiDefault := defaultAPIBase()
	apiAddr := root.String("api", apiDefault, "AnchorChain HTTP API base URL (env ANCHORCHAIN_API)")
	token := root.String("token", strings.TrimSpace(os.Getenv("ANCHORCHAIN_API_TOKEN")), "AnchorChain API token (env ANCHORCHAIN_API_TOKEN)")
	jsonOut := root.Bool("json", false, "Print raw JSON responses")

	if err := root.Parse(args); err != nil {
		return rootConfig{}, err
	}

	return rootConfig{
		baseURL: strings.TrimSpace(*apiAddr),
		token:   strings.TrimSpace(*token),
		json:    *jsonOut,
		args:    root.Args(),
	}, nil
}

type cli struct {
	client     *apiClient
	jsonOutput bool
}

func (c *cli) dispatch(args []string) error {
	if len(args) == 0 {
		printMainUsage()
		return errors.New("missing command")
	}

	switch args[0] {
	case "chain":
		return c.chain(args[1:])
	case "entry":
		return c.entry(args[1:])
	case "receipt":
		return c.receipt(args[1:])
	case "node":
		return c.node(args[1:])
	case "help", "--help", "-h":
		printMainUsage()
		return nil
	default:
		printMainUsage()
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func (c *cli) chain(args []string) error {
	if len(args) == 0 {
		printChainUsage()
		return errors.New("missing chain subcommand")
	}

	switch args[0] {
	case "create":
		return c.chainCreate(args[1:])
	case "inspect":
		return c.chainInspect(args[1:])
	case "tail":
		return c.chainTail(args[1:])
	default:
		printChainUsage()
		return fmt.Errorf("unknown chain subcommand: %s", args[0])
	}
}

func (c *cli) entry(args []string) error {
	if len(args) == 0 {
		printEntryUsage()
		return errors.New("missing entry subcommand")
	}

	switch args[0] {
	case "write":
		return c.entryWrite(args[1:])
	case "show":
		return c.entryShow(args[1:])
	default:
		printEntryUsage()
		return fmt.Errorf("unknown entry subcommand: %s", args[0])
	}
}

func (c *cli) receipt(args []string) error {
	if len(args) == 0 {
		printReceiptUsage()
		return errors.New("missing receipt subcommand")
	}
	if args[0] != "verify" {
		printReceiptUsage()
		return fmt.Errorf("unknown receipt subcommand: %s", args[0])
	}
	return c.receiptVerify(args[1:])
}

func (c *cli) node(args []string) error {
	if len(args) == 0 {
		printNodeUsage()
		return errors.New("missing node subcommand")
	}
	if args[0] != "health" {
		printNodeUsage()
		return fmt.Errorf("unknown node subcommand: %s", args[0])
	}
	return c.nodeHealth(args[1:])
}

func (c *cli) chainCreate(args []string) error {
	fs := flag.NewFlagSet("anchor-cli chain create", flag.ContinueOnError)
	setUsage(fs, "Usage: anchor-cli chain create [options]")
	var extIDs multiFlag
	fs.Var(&extIDs, "extid", "External ID (repeatable)")
	schema := fs.String("schema", "", "Schema identifier for structured payloads")
	payload := fs.String("payload", "", "Structured JSON payload")
	content := fs.String("content", "", "Raw entry content")
	encoding := fs.String("encoding", "utf-8", "Content encoding for --content (utf-8 or base64)")
	ecKey := fs.String("ec-key", strings.TrimSpace(os.Getenv("ANCHORCHAIN_EC_PRIVATE")), "Entry credit private key (optional; env ANCHORCHAIN_EC_PRIVATE)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if *payload == "" && *content == "" {
		return errors.New("either --payload or --content is required")
	}
	if *payload != "" && *content != "" {
		return errors.New("use --payload or --content, not both")
	}

	body := map[string]interface{}{}
	if len(extIDs) > 0 {
		body["extIds"] = []string(extIDs)
		body["extIdsEncoding"] = "utf-8"
	}
	if strings.TrimSpace(*schema) != "" {
		body["schema"] = strings.TrimSpace(*schema)
	}
	if strings.TrimSpace(*ecKey) != "" {
		body["ecPrivateKey"] = strings.TrimSpace(*ecKey)
	}
	if *payload != "" {
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(*payload), &raw); err != nil {
			return fmt.Errorf("invalid JSON payload: %w", err)
		}
		body["payload"] = raw
	} else {
		body["content"] = *content
		body["contentEncoding"] = strings.TrimSpace(*encoding)
	}

	var resp writeResponse
	if err := c.client.postJSON("/chains", body, &resp); err != nil {
		return fmt.Errorf("failed to create chain: %w", err)
	}

	if c.jsonOutput {
		return printJSON(resp)
	}
	printWriteResponse("Chain", &resp)
	return nil
}

func (c *cli) entryWrite(args []string) error {
	fs := flag.NewFlagSet("anchor-cli entry write", flag.ContinueOnError)
	setUsage(fs, "Usage: anchor-cli entry write --chain <chainId> [options]")
	chainID := fs.String("chain", "", "Chain ID (required)")
	var extIDs multiFlag
	fs.Var(&extIDs, "extid", "External ID (repeatable)")
	schema := fs.String("schema", "", "Schema identifier for structured payloads")
	payload := fs.String("payload", "", "Structured JSON payload")
	content := fs.String("content", "", "Raw entry content")
	encoding := fs.String("encoding", "utf-8", "Content encoding for --content (utf-8 or base64)")
	ecKey := fs.String("ec-key", strings.TrimSpace(os.Getenv("ANCHORCHAIN_EC_PRIVATE")), "Entry credit private key (optional; env ANCHORCHAIN_EC_PRIVATE)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if strings.TrimSpace(*chainID) == "" {
		return errors.New("--chain is required")
	}
	if *payload == "" && *content == "" {
		return errors.New("either --payload or --content is required")
	}
	if *payload != "" && *content != "" {
		return errors.New("use --payload or --content, not both")
	}

	body := map[string]interface{}{}
	if len(extIDs) > 0 {
		body["extIds"] = []string(extIDs)
		body["extIdsEncoding"] = "utf-8"
	}
	if strings.TrimSpace(*schema) != "" {
		body["schema"] = strings.TrimSpace(*schema)
	}
	if strings.TrimSpace(*ecKey) != "" {
		body["ecPrivateKey"] = strings.TrimSpace(*ecKey)
	}
	if *payload != "" {
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(*payload), &raw); err != nil {
			return fmt.Errorf("invalid JSON payload: %w", err)
		}
		body["payload"] = raw
	} else {
		body["content"] = *content
		body["contentEncoding"] = strings.TrimSpace(*encoding)
	}

	path := fmt.Sprintf("/chains/%s/entries", url.PathEscape(strings.TrimSpace(*chainID)))
	var resp writeResponse
	if err := c.client.postJSON(path, body, &resp); err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}

	if c.jsonOutput {
		return printJSON(resp)
	}
	printWriteResponse("Entry", &resp)
	return nil
}

func (c *cli) chainInspect(args []string) error {
	fs := flag.NewFlagSet("anchor-cli chain inspect", flag.ContinueOnError)
	setUsage(fs, "Usage: anchor-cli chain inspect --chain <chainId>")
	chainID := fs.String("chain", "", "Chain ID")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(*chainID) == "" {
		return errors.New("--chain is required")
	}

	path := fmt.Sprintf("/chains/%s", url.PathEscape(strings.TrimSpace(*chainID)))
	var resp chainInspectResponse
	if err := c.client.getJSON(path, &resp); err != nil {
		return fmt.Errorf("failed to inspect chain: %w", err)
	}
	if strings.TrimSpace(resp.Error) != "" {
		return fmt.Errorf("failed to inspect chain: %s", resp.Error)
	}

	if c.jsonOutput {
		return printJSON(resp)
	}
	c.printChainInspectTable(&resp)
	return nil
}

func (c *cli) chainTail(args []string) error {
	fs := flag.NewFlagSet("anchor-cli chain tail", flag.ContinueOnError)
	setUsage(fs, "Usage: anchor-cli chain tail --chain <chainId> [--limit N] [--offset N]")
	chainID := fs.String("chain", "", "Chain ID")
	limit := fs.Int("limit", 10, "Number of entries to return")
	offset := fs.Int("offset", 0, "Starting offset")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(*chainID) == "" {
		return errors.New("--chain is required")
	}

	path := fmt.Sprintf("/chains/%s/entries?limit=%d&offset=%d", url.PathEscape(strings.TrimSpace(*chainID)), *limit, *offset)
	var resp interface{}
	if err := c.client.getJSON(path, &resp); err != nil {
		return fmt.Errorf("failed to list entries: %w", err)
	}
	return printJSON(resp)
}

func (c *cli) entryShow(args []string) error {
	fs := flag.NewFlagSet("anchor-cli entry show", flag.ContinueOnError)
	setUsage(fs, "Usage: anchor-cli entry show --entry <entryHash>")
	entryHash := fs.String("entry", "", "Entry hash")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(*entryHash) == "" {
		return errors.New("--entry is required")
	}

	path := fmt.Sprintf("/entries/%s", url.PathEscape(strings.TrimSpace(*entryHash)))
	var resp entryDetail
	if err := c.client.getJSON(path, &resp); err != nil {
		return fmt.Errorf("failed to load entry: %w", err)
	}
	if strings.TrimSpace(resp.Error) != "" {
		return fmt.Errorf("failed to load entry: %s", resp.Error)
	}

	if c.jsonOutput {
		return printJSON(resp)
	}
	return c.printEntryDetail(&resp)
}

func (c *cli) receiptVerify(args []string) error {
	fs := flag.NewFlagSet("anchor-cli receipt verify", flag.ContinueOnError)
	setUsage(fs, "Usage: anchor-cli receipt verify --entry <entryHash> [--include-raw]")
	entryHash := fs.String("entry", "", "Entry hash to verify")
	includeRaw := fs.Bool("include-raw", false, "Include raw entry bytes in the response")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(*entryHash) == "" {
		return errors.New("--entry is required")
	}

	body := map[string]interface{}{
		"entryHash": strings.TrimSpace(*entryHash),
	}
	if *includeRaw {
		body["includeRawEntry"] = true
	}

	var resp interface{}
	if err := c.client.postJSON("/receipts/verify", body, &resp); err != nil {
		return fmt.Errorf("failed to verify receipt: %w", err)
	}
	return printJSON(resp)
}

func (c *cli) nodeHealth(args []string) error {
	if len(args) != 0 {
		return errors.New("node health does not accept additional arguments")
	}
	var resp interface{}
	if err := c.client.getJSON("/health", &resp); err != nil {
		return fmt.Errorf("failed to fetch health: %w", err)
	}
	return printJSON(resp)
}

func (c *cli) printChainInspectTable(resp *chainInspectResponse) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "Chain ID:\t%s\n", resp.ChainID)
	if resp.EntryCount != nil {
		fmt.Fprintf(tw, "Entries:\t%d\n", *resp.EntryCount)
	}
	if resp.LatestEntryHash != "" {
		fmt.Fprintf(tw, "Latest Entry Hash:\t%s\n", resp.LatestEntryHash)
	}
	if resp.LatestEntryTimestamp != nil {
		ts := time.Unix(*resp.LatestEntryTimestamp, 0).UTC().Format(time.RFC3339)
		fmt.Fprintf(tw, "Latest Entry Timestamp:\t%s\n", ts)
	}
	tw.Flush()
}

func (c *cli) printEntryDetail(detail *entryDetail) error {
	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "Entry Hash:\t%s\n", detail.EntryHash)
	fmt.Fprintf(tw, "Chain ID:\t%s\n", detail.ChainID)
	fmt.Fprintf(tw, "Schema:\t%s\n", detail.Schema)
	fmt.Fprintf(tw, "Structured:\t%t\n", detail.Structured)
	fmt.Fprintf(tw, "Content Encoding:\t%s\n", detail.ContentEncoding)
	tw.Flush()

	if len(detail.ExtIDs) > 0 {
		fmt.Println("ExtIDs:")
		for idx, id := range detail.ExtIDs {
			fmt.Printf("  [%d] %s\n", idx, id)
		}
	}

	if detail.DecodedPayload != nil {
		payload, err := json.MarshalIndent(detail.DecodedPayload, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println("Decoded Payload:")
		fmt.Println(string(payload))
	} else if detail.Content != "" {
		fmt.Println("Content:")
		fmt.Println(detail.Content)
	}

	return nil
}

func printWriteResponse(label string, resp *writeResponse) {
	fmt.Printf("%s submitted (status=%s)\n", label, resp.Status)
	if resp.ChainID != "" {
		fmt.Printf("  Chain ID   : %s\n", resp.ChainID)
	}
	if resp.EntryHash != "" {
		fmt.Printf("  Entry Hash : %s\n", resp.EntryHash)
	}
	if resp.TxID != "" {
		fmt.Printf("  TxID       : %s\n", resp.TxID)
	}
	if strings.TrimSpace(resp.Message) != "" {
		fmt.Printf("  Message    : %s\n", resp.Message)
	}
}

func setUsage(fs *flag.FlagSet, usage string) {
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), usage)
		fs.PrintDefaults()
	}
}

func printMainUsage() {
	fmt.Fprintln(os.Stderr, "AnchorChain CLI")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  anchor-cli [--api URL] [--token TOKEN] [--json] <command> [<args>]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	tw := tabwriter.NewWriter(os.Stderr, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "  chain create\tCreate a new chain")
	fmt.Fprintln(tw, "  chain inspect\tInspect chain metadata")
	fmt.Fprintln(tw, "  chain tail\tList recent entries on a chain")
	fmt.Fprintln(tw, "  entry write\tAppend an entry to a chain")
	fmt.Fprintln(tw, "  entry show\tDisplay a specific entry")
	fmt.Fprintln(tw, "  receipt verify\tFetch and verify an entry receipt")
	fmt.Fprintln(tw, "  node health\tCheck node health/heights")
	fmt.Fprintln(tw, "  help\tShow this message")
	tw.Flush()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Global Flags:")
	fmt.Fprintln(os.Stderr, "  --api     AnchorChain HTTP API base URL (env ANCHORCHAIN_API)")
	fmt.Fprintln(os.Stderr, "  --token   AnchorChain API token (env ANCHORCHAIN_API_TOKEN)")
	fmt.Fprintln(os.Stderr, "  --json    Print raw JSON responses")
}

func printChainUsage() {
	fmt.Fprintln(os.Stderr, "Chain commands: create, inspect, tail")
}

func printEntryUsage() {
	fmt.Fprintln(os.Stderr, "Entry commands: write, show")
}

func printReceiptUsage() {
	fmt.Fprintln(os.Stderr, "Receipt commands: verify")
}

func printNodeUsage() {
	fmt.Fprintln(os.Stderr, "Node commands: health")
}

type multiFlag []string

func (m *multiFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

type apiClient struct {
	base  *url.URL
	token string
	http  *http.Client
}

func newAPIClient(raw, token string) (*apiClient, error) {
	addr := strings.TrimSpace(raw)
	if addr == "" {
		addr = defaultAPIBase()
	}
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	parsed, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid api address: %w", err)
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	return &apiClient{
		base:  parsed,
		token: token,
		http:  &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (c *apiClient) getJSON(path string, out interface{}) error {
	return c.doJSON(http.MethodGet, path, nil, out)
}

func (c *apiClient) postJSON(path string, body interface{}, out interface{}) error {
	return c.doJSON(http.MethodPost, path, body, out)
}

func (c *apiClient) doJSON(method, path string, body interface{}, out interface{}) error {
	rel, err := url.Parse(path)
	if err != nil {
		return err
	}
	u := c.base.ResolveReference(rel)

	var reader io.Reader
	if body != nil {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return err
		}
		reader = buf
	}

	req, err := http.NewRequest(method, u.String(), reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("X-Anchorchain-Api-Token", c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		msg := strings.TrimSpace(string(payload))
		var apiErr struct {
			Error   string `json:"error"`
			Message string `json:"message"`
			Status  string `json:"status"`
		}
		if err := json.Unmarshal(payload, &apiErr); err == nil {
			if apiErr.Error != "" {
				msg = apiErr.Error
			} else if apiErr.Message != "" {
				msg = apiErr.Message
			} else if apiErr.Status != "" {
				msg = apiErr.Status
			}
		}
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("API %s %s returned %d: %s", method, rel.String(), resp.StatusCode, msg)
	}

	if out == nil || len(payload) == 0 {
		return nil
	}
	return json.Unmarshal(payload, out)
}

func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func defaultAPIBase() string {
	if v := strings.TrimSpace(os.Getenv("ANCHORCHAIN_API")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("ANCHORCHAIN_API_ADDR")); v != "" {
		if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
			return v
		}
		return "http://" + v
	}
	return "http://127.0.0.1:8081"
}

type writeResponse struct {
	Success   bool   `json:"success"`
	Status    string `json:"status"`
	ChainID   string `json:"chainId"`
	EntryHash string `json:"entryHash"`
	TxID      string `json:"txId"`
	Message   string `json:"message"`
}

type chainInspectResponse struct {
	ChainID              string `json:"chainId"`
	EntryCount           *int   `json:"entryCount"`
	LatestEntryHash      string `json:"latestEntryHash"`
	LatestEntryTimestamp *int64 `json:"latestEntryTimestamp"`
	Error                string `json:"error"`
}

type entryDetail struct {
	EntryHash       string      `json:"entryHash"`
	ChainID         string      `json:"chainId"`
	ExtIDs          []string    `json:"extIds"`
	Schema          string      `json:"schema"`
	Structured      bool        `json:"structured"`
	Content         string      `json:"content"`
	ContentEncoding string      `json:"contentEncoding"`
	DecodedPayload  interface{} `json:"decodedPayload"`
	Error           string      `json:"error"`
}
