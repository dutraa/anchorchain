package api

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/FactomProject/factom"
)

type Options struct {
	DevnetECPrivateKey string
	DevnetMode         bool
	AuthToken          string
}

// Server exposes a modern HTTP facade for basic write operations.
type Server struct {
	factomdAddr        string
	logger             *log.Logger
	devnetECPrivateKey string
	devnetMode         bool
	authToken          string
}

const (
	defaultEntryLimit = 50
	maxEntryLimit     = 500
)

// Start launches the API server on addr and routes requests into the legacy factomd RPC surface.
func Start(addr, factomdAddr string, logger *log.Logger, opts Options) *http.Server {
	srv := &http.Server{
		Addr:    addr,
		Handler: newServer(factomdAddr, logger, opts),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Printf("api server error: %v", err)
		}
	}()

	modeLabel := "standard"
	if opts.DevnetMode {
		modeLabel = "devnet"
	}
	authLabel := "open"
	if strings.TrimSpace(opts.AuthToken) != "" {
		authLabel = "token-required"
	}

	logger.Printf("AnchorChain HTTP API listening on http://%s [mode=%s auth=%s] (proxying to legacy factomd RPC %s)", addr, modeLabel, authLabel, factomdAddr)
	if opts.DevnetMode {
		logger.Printf("[NOTICE] Devnet mode is for non-production testing only.")
	}
	if strings.TrimSpace(opts.AuthToken) != "" {
		logger.Printf("[SECURITY] HTTP API authentication required via X-Anchorchain-Api-Token (preferred) or Authorization header. Legacy X-Factom headers remain supported.")
	}

	return srv
}

func newServer(factomdAddr string, logger *log.Logger, opts Options) http.Handler {
	return &Server{
		factomdAddr:        factomdAddr,
		logger:             logger,
		devnetECPrivateKey: opts.DevnetECPrivateKey,
		devnetMode:         opts.DevnetMode,
		authToken:          strings.TrimSpace(opts.AuthToken),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth(w, r) {
		return
	}

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/health":
		s.handleHealth(w, r)
		return
	case r.Method == http.MethodPost && r.URL.Path == "/receipts/verify":
		s.handleReceiptVerify(w, r)
		return
	case r.Method == http.MethodPost && r.URL.Path == "/chains":
		s.handleCreateChain(w, r)
		return
	case strings.HasPrefix(r.URL.Path, "/chains/"):
		s.switchChainRoutes(w, r)
		return
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/entries/"):
		s.handleEntryLookup(w, r)
		return
	default:
		http.NotFound(w, r)
		return
	}
}

func (s *Server) switchChainRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.handleChainEntry(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/chains/")
	path = strings.Trim(path, "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	parts := strings.Split(path, "/")
	switch len(parts) {
	case 1:
		s.handleGetChain(w, r, parts[0])
	case 2:
		if parts[1] == "entries" {
			s.handleListChainEntries(w, r, parts[0])
			return
		}
		fallthrough
	default:
		http.NotFound(w, r)
	}
}

type createChainRequest struct {
	ExtIDs          []string        `json:"extIds"`
	ExtIDsEncoding  string          `json:"extIdsEncoding,omitempty"`
	Content         string          `json:"content"`
	ContentEncoding string          `json:"contentEncoding,omitempty"`
	Schema          string          `json:"schema,omitempty"`
	Payload         json.RawMessage `json:"payload"`
	ECPrivateKey    string          `json:"ecPrivateKey"`
}

type entryRequest struct {
	ExtIDs          []string        `json:"extIds"`
	ExtIDsEncoding  string          `json:"extIdsEncoding,omitempty"`
	Content         string          `json:"content"`
	ContentEncoding string          `json:"contentEncoding,omitempty"`
	Schema          string          `json:"schema,omitempty"`
	Payload         json.RawMessage `json:"payload"`
	ECPrivateKey    string          `json:"ecPrivateKey"`
}

type writeResponse struct {
	Success    bool   `json:"success"`
	Status     string `json:"status"`
	ChainID    string `json:"chainId,omitempty"`
	EntryHash  string `json:"entryHash,omitempty"`
	TxID       string `json:"txId,omitempty"`
	Message    string `json:"message,omitempty"`
	Schema     string `json:"schema,omitempty"`
	Structured bool   `json:"structured,omitempty"`
}

type receiptVerifyRequest struct {
	EntryHash       string `json:"entryHash"`
	IncludeRawEntry bool   `json:"includeRawEntry,omitempty"`
}

type receiptVerifyResponse struct {
	Success   bool            `json:"success"`
	EntryHash string          `json:"entryHash"`
	Receipt   *factom.Receipt `json:"receipt,omitempty"`
	Message   string          `json:"message,omitempty"`
}

type healthResponse struct {
	Status  string                  `json:"status"`
	Heights *factom.HeightsResponse `json:"heights"`
}

func (s *Server) handleCreateChain(w http.ResponseWriter, r *http.Request) {
	var req createChainRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: fmt.Sprintf("invalid JSON: %v", err)})
		return
	}

	ecKey := strings.TrimSpace(req.ECPrivateKey)
	if ecKey == "" && s.devnetMode && s.devnetECPrivateKey != "" {
		ecKey = s.devnetECPrivateKey
	}
	if ecKey == "" {
		respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: "ecPrivateKey is required"})
		return
	}

	structured := strings.TrimSpace(req.Schema) != ""
	var content []byte
	var err error
	if structured {
		if len(req.Payload) == 0 {
			respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: "payload is required when schema is set"})
			return
		}
		if !json.Valid(req.Payload) {
			respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: "payload must be valid JSON"})
			return
		}
		content, err = compactJSON(req.Payload)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: fmt.Sprintf("unable to normalize payload: %v", err)})
			return
		}
	} else {
		content, err = decodePayload(req.Content, req.ContentEncoding)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: err.Error()})
			return
		}
	}

	extIDs, err := decodeExtIDs(req.ExtIDs, req.ExtIDsEncoding)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: err.Error()})
		return
	}
	if structured {
		extIDs = prependSchemaExtID(extIDs, req.Schema)
	}

	entry := &factom.Entry{
		ExtIDs:  extIDs,
		Content: content,
	}

	chain := factom.NewChain(entry)

	ecAddr, err := factom.GetECAddress(ecKey)
	if err != nil {
		s.logger.Printf("invalid ecPrivateKey: %v", err)
		respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: "invalid ecPrivateKey supplied"})
		return
	}

	factom.SetFactomdServer(s.factomdAddr)
	txID, err := factom.CommitChain(chain, ecAddr)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, writeResponse{Success: false, Status: "error", Message: fmt.Sprintf("commit failed: %v", err)})
		return
	}

	entryHash, err := factom.RevealChain(chain)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, writeResponse{Success: false, Status: "error", Message: fmt.Sprintf("reveal failed: %v", err)})
		return
	}

	respondJSON(w, http.StatusAccepted, writeResponse{
		Success:    true,
		Status:     "pending",
		ChainID:    chain.ChainID,
		EntryHash:  entryHash,
		TxID:       txID,
		Message:    "chain submitted",
		Schema:     strings.TrimSpace(req.Schema),
		Structured: structured,
	})
}

func (s *Server) handleChainEntry(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/chains/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "entries" {
		http.NotFound(w, r)
		return
	}
	chainID := parts[0]
	if chainID == "" {
		respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: "chainId missing in path"})
		return
	}

	var req entryRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: fmt.Sprintf("invalid JSON: %v", err)})
		return
	}

	ecKey := strings.TrimSpace(req.ECPrivateKey)
	if ecKey == "" && s.devnetMode && s.devnetECPrivateKey != "" {
		ecKey = s.devnetECPrivateKey
	}
	if ecKey == "" {
		respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: "ecPrivateKey is required"})
		return
	}

	structured := strings.TrimSpace(req.Schema) != ""
	var content []byte
	var err error
	if structured {
		if len(req.Payload) == 0 {
			respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: "payload is required when schema is set"})
			return
		}
		if !json.Valid(req.Payload) {
			respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: "payload must be valid JSON"})
			return
		}
		content, err = compactJSON(req.Payload)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: fmt.Sprintf("unable to normalize payload: %v", err)})
			return
		}
	} else {
		content, err = decodePayload(req.Content, req.ContentEncoding)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: err.Error()})
			return
		}
	}

	extIDs, err := decodeExtIDs(req.ExtIDs, req.ExtIDsEncoding)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: err.Error()})
		return
	}
	if structured {
		extIDs = prependSchemaExtID(extIDs, req.Schema)
	}

	entry := &factom.Entry{
		ChainID: chainID,
		ExtIDs:  extIDs,
		Content: content,
	}

	ecAddr, err := factom.GetECAddress(ecKey)
	if err != nil {
		s.logger.Printf("invalid ecPrivateKey: %v", err)
		respondJSON(w, http.StatusBadRequest, writeResponse{Success: false, Status: "error", Message: "invalid ecPrivateKey supplied"})
		return
	}

	factom.SetFactomdServer(s.factomdAddr)
	txID, err := factom.CommitEntry(entry, ecAddr)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, writeResponse{Success: false, Status: "error", Message: fmt.Sprintf("commit failed: %v", err)})
		return
	}

	entryHash, err := factom.RevealEntry(entry)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, writeResponse{Success: false, Status: "error", Message: fmt.Sprintf("reveal failed: %v", err)})
		return
	}

	respondJSON(w, http.StatusAccepted, writeResponse{
		Success:    true,
		Status:     "pending",
		ChainID:    chainID,
		EntryHash:  entryHash,
		TxID:       txID,
		Message:    "entry submitted",
		Schema:     strings.TrimSpace(req.Schema),
		Structured: structured,
	})
}

func (s *Server) handleGetChain(w http.ResponseWriter, r *http.Request, chainID string) {
	factom.SetFactomdServer(s.factomdAddr)

	blocks, err := s.loadChainBlocks(chainID)
	if err != nil {
		s.respondFactomError(w, err)
		return
	}

	summary := summarizeChain(chainID, blocks)
	respondJSONAny(w, http.StatusOK, summary)
}

func (s *Server) handleListChainEntries(w http.ResponseWriter, r *http.Request, chainID string) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	factom.SetFactomdServer(s.factomdAddr)

	blocks, err := s.loadChainBlocks(chainID)
	if err != nil {
		s.respondFactomError(w, err)
		return
	}

	refs := entriesFromBlocks(blocks)
	total := len(refs)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	entries := make([]entrySummary, 0, end-offset)
	for _, meta := range refs[offset:end] {
		entry, err := factom.GetEntry(meta.Hash)
		if err != nil {
			s.respondFactomError(w, err)
			return
		}
		entries = append(entries, buildEntrySummary(meta, entry))
	}

	respondJSONAny(w, http.StatusOK, entryListResponse{
		ChainID: chainID,
		Entries: entries,
		Limit:   limit,
		Offset:  offset,
		Total:   total,
	})
}

func (s *Server) handleEntryLookup(w http.ResponseWriter, r *http.Request) {
	entryHash := strings.TrimPrefix(r.URL.Path, "/entries/")
	entryHash = strings.Trim(entryHash, "/")
	if entryHash == "" {
		http.NotFound(w, r)
		return
	}

	factom.SetFactomdServer(s.factomdAddr)

	entry, err := factom.GetEntry(entryHash)
	if err != nil {
		s.respondFactomError(w, err)
		return
	}

	detail, err := buildEntryDetail(entryHash, entry)
	if err != nil {
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}

	respondJSONAny(w, http.StatusOK, detail)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	factom.SetFactomdServer(s.factomdAddr)
	res, err := factom.GetHeights()
	if err != nil {
		s.respondFactomError(w, err)
		return
	}
	respondJSONAny(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Heights: res,
	})
}

func (s *Server) handleReceiptVerify(w http.ResponseWriter, r *http.Request) {
	var req receiptVerifyRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}
	entryHash := strings.TrimSpace(req.EntryHash)
	if entryHash == "" {
		respondError(w, http.StatusBadRequest, "entryHash is required")
		return
	}
	factom.SetFactomdServer(s.factomdAddr)
	receipt, err := factom.GetReceipt(entryHash)
	if err != nil {
		s.respondFactomError(w, err)
		return
	}
	if receipt == nil {
		respondError(w, http.StatusNotFound, "receipt not available")
		return
	}
	if !req.IncludeRawEntry {
		receipt.Entry.Raw = ""
	}
	respondJSONAny(w, http.StatusOK, receiptVerifyResponse{
		Success:   true,
		EntryHash: entryHash,
		Receipt:   receipt,
		Message:   "receipt retrieved",
	})
}

func prependSchemaExtID(extIDs [][]byte, schema string) [][]byte {
	s := fmt.Sprintf("schema:%s", strings.TrimSpace(schema))
	out := make([][]byte, 0, len(extIDs)+1)
	out = append(out, []byte(s))
	out = append(out, extIDs...)
	return out
}

func compactJSON(raw []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodePayload(content, enc string) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(enc)) {
	case "", "utf-8", "utf8", "plain", "text":
		return []byte(content), nil
	case "base64", "b64":
		data, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, fmt.Errorf("unable to decode base64 content: %w", err)
		}
		return data, nil
	default:
		return nil, fmt.Errorf("unsupported contentEncoding %q", enc)
	}
}

func decodeExtIDs(ids []string, enc string) ([][]byte, error) {
	result := make([][]byte, 0, len(ids))
	for _, id := range ids {
		data, err := decodePayload(id, enc)
		if err != nil {
			return nil, fmt.Errorf("invalid extId: %w", err)
		}
		result = append(result, data)
	}
	return result, nil
}

func (s *Server) requireAuth(w http.ResponseWriter, r *http.Request) bool {
	if s.authToken == "" {
		return true
	}

	token := extractAuthToken(r)
	if token == "" {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}

	if secureCompare(token, s.authToken) {
		return true
	}

	respondError(w, http.StatusUnauthorized, "unauthorized")
	return false
}

func extractAuthToken(r *http.Request) string {
	checks := []string{
		"X-Anchorchain-Api-Token",
		"X-Anchorchain-Api-Key",
		"X-Factom-Api-Token",
		"X-Factom-Api-Key",
	}
	for _, header := range checks {
		token := strings.TrimSpace(r.Header.Get(header))
		if token != "" {
			return token
		}
	}
	authz := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		return strings.TrimSpace(authz[7:])
	}
	return ""
}

func secureCompare(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func (s *Server) loadChainBlocks(chainID string) ([]*factom.EBlock, error) {
	head, inProcess, err := factom.GetChainHead(chainID)
	if err != nil {
		return nil, err
	}
	if head == "" {
		if inProcess {
			return nil, factom.ErrChainPending
		}
		return nil, fmt.Errorf("chain %s not found", chainID)
	}

	blocks := make([]*factom.EBlock, 0)
	for cursor := head; cursor != factom.ZeroHash; {
		eb, err := factom.GetEBlock(cursor)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, eb)
		cursor = eb.Header.PrevKeyMR
	}
	return blocks, nil
}

func entriesFromBlocks(blocks []*factom.EBlock) []entryMetadata {
	entries := make([]entryMetadata, 0)
	for i := len(blocks) - 1; i >= 0; i-- {
		eb := blocks[i]
		for _, e := range eb.EntryList {
			meta := entryMetadata{Hash: e.EntryHash}
			if e.Timestamp != 0 {
				meta.Timestamp = e.Timestamp
				meta.HasTimestamp = true
			}
			entries = append(entries, meta)
		}
	}
	return entries
}

func summarizeChain(chainID string, blocks []*factom.EBlock) chainResponse {
	summary := chainResponse{ChainID: chainID}
	total := 0
	var latestHash string
	var latestTS int64
	hasLatest := false

	for _, eb := range blocks {
		total += len(eb.EntryList)
		for _, entry := range eb.EntryList {
			ts := entry.Timestamp
			if !hasLatest || ts > latestTS {
				hasLatest = true
				latestTS = ts
				latestHash = entry.EntryHash
			}
		}
	}

	summary.EntryCount = intPtr(total)
	if hasLatest {
		summary.LatestEntryHash = latestHash
		summary.LatestEntryTimestamp = int64Ptr(latestTS)
	}

	return summary
}

func buildEntrySummary(meta entryMetadata, entry *factom.Entry) entrySummary {
	extIDs, schema, structured := formatExtIDs(entry.ExtIDs)
	summary := entrySummary{
		EntryHash:  meta.Hash,
		ExtIDs:     extIDs,
		Structured: structured,
	}
	if schema != "" {
		summary.Schema = schema
	}
	if meta.HasTimestamp {
		summary.Timestamp = int64Ptr(meta.Timestamp)
	}
	return summary
}

func buildEntryDetail(entryHash string, entry *factom.Entry) (*entryDetailResponse, error) {
	extIDs, schema, structured := formatExtIDs(entry.ExtIDs)
	content, encoding, decoded, err := formatEntryContent(entry, structured)
	if err != nil {
		return nil, err
	}

	detail := &entryDetailResponse{
		EntryHash:       entryHash,
		ChainID:         entry.ChainID,
		ExtIDs:          extIDs,
		Schema:          schema,
		Structured:      structured,
		Content:         content,
		ContentEncoding: encoding,
	}
	if decoded != nil {
		detail.DecodedPayload = decoded
	}

	return detail, nil
}

func formatExtIDs(ids [][]byte) ([]string, string, bool) {
	encoded := make([]string, len(ids))
	for i, id := range ids {
		encoded[i] = base64.StdEncoding.EncodeToString(id)
	}
	schema, structured := extractSchema(ids)
	return encoded, schema, structured
}

func extractSchema(extIDs [][]byte) (string, bool) {
	if len(extIDs) == 0 {
		return "", false
	}
	raw := string(extIDs[0])
	const prefix = "schema:"
	if len(raw) >= len(prefix) && strings.EqualFold(raw[:len(prefix)], prefix) {
		return strings.TrimSpace(raw[len(prefix):]), true
	}
	return "", false
}

func formatEntryContent(entry *factom.Entry, structured bool) (string, string, interface{}, error) {
	if structured {
		content := string(entry.Content)
		if len(entry.Content) == 0 {
			return content, "utf-8", nil, nil
		}
		var decoded interface{}
		if err := json.Unmarshal(entry.Content, &decoded); err != nil {
			return "", "", nil, fmt.Errorf("unable to decode structured entry payload: %w", err)
		}
		return content, "utf-8", decoded, nil
	}

	return base64.StdEncoding.EncodeToString(entry.Content), "base64", nil, nil
}

func parsePagination(r *http.Request) (int, int, error) {
	limit := defaultEntryLimit
	if v := r.URL.Query().Get("limit"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed <= 0 {
			return 0, 0, fmt.Errorf("limit must be a positive integer")
		}
		limit = parsed
	}
	if limit > maxEntryLimit {
		limit = maxEntryLimit
	}

	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 0 {
			return 0, 0, fmt.Errorf("offset must be zero or a positive integer")
		}
		offset = parsed
	}

	return limit, offset, nil
}

func (s *Server) respondFactomError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	if errors.Is(err, factom.ErrChainPending) {
		respondError(w, http.StatusAccepted, "chain is pending confirmation")
		return
	}

	var je *factom.JSONError
	if errors.As(err, &je) {
		switch je.Code {
		case -32008, -32009:
			respondError(w, http.StatusNotFound, je.Message)
			return
		case -32602:
			respondError(w, http.StatusBadRequest, je.Message)
			return
		default:
			respondError(w, http.StatusBadGateway, je.Error())
			return
		}
	}

	respondError(w, http.StatusBadGateway, err.Error())
}

func respondJSONAny(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSONAny(w, status, apiErrorResponse{Error: message})
}

func intPtr(v int) *int {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

type chainResponse struct {
	ChainID              string `json:"chainId"`
	EntryCount           *int   `json:"entryCount,omitempty"`
	LatestEntryHash      string `json:"latestEntryHash,omitempty"`
	LatestEntryTimestamp *int64 `json:"latestEntryTimestamp,omitempty"`
}

type entryListResponse struct {
	ChainID string         `json:"chainId"`
	Entries []entrySummary `json:"entries"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
	Total   int            `json:"total"`
}

type entrySummary struct {
	EntryHash  string   `json:"entryHash"`
	Timestamp  *int64   `json:"timestamp,omitempty"`
	ExtIDs     []string `json:"extIds"`
	Schema     string   `json:"schema,omitempty"`
	Structured bool     `json:"structured"`
}

type entryDetailResponse struct {
	EntryHash       string      `json:"entryHash"`
	ChainID         string      `json:"chainId"`
	ExtIDs          []string    `json:"extIds"`
	Schema          string      `json:"schema,omitempty"`
	Structured      bool        `json:"structured"`
	Content         string      `json:"content"`
	ContentEncoding string      `json:"contentEncoding"`
	DecodedPayload  interface{} `json:"decodedPayload,omitempty"`
}

type entryMetadata struct {
	Hash         string
	Timestamp    int64
	HasTimestamp bool
}

type apiErrorResponse struct {
	Error string `json:"error"`
}

func respondJSON(w http.ResponseWriter, status int, payload writeResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}
