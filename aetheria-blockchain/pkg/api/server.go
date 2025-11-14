package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aetheria/blockchain/pkg/blockchain"
	"github.com/aetheria/blockchain/pkg/consensus"
	"github.com/aetheria/blockchain/pkg/crypto"
	"github.com/aetheria/blockchain/pkg/network"
)

// Server represents the API server
type Server struct {
	Port       int
	Node       *network.Node
	Blockchain *blockchain.Blockchain
	Consensus  *consensus.PoS
}

// NewServer creates a new API server
func NewServer(port int, node *network.Node, bc *blockchain.Blockchain, pos *consensus.PoS) *Server {
	return &Server{
		Port:       port,
		Node:       node,
		Blockchain: bc,
		Consensus:  pos,
	}
}

// Start starts the API server
func (s *Server) Start() error {
	http.HandleFunc("/", s.handleRoot)
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/blocks", s.handleBlocks)
	http.HandleFunc("/block/", s.handleBlock)
	http.HandleFunc("/transactions", s.handleTransactions)
	http.HandleFunc("/transaction/", s.handleTransaction)
	http.HandleFunc("/balance/", s.handleBalance)
	http.HandleFunc("/stake", s.handleStake)
	http.HandleFunc("/validators", s.handleValidators)
	http.HandleFunc("/wallet/new", s.handleNewWallet)

	addr := fmt.Sprintf(":%d", s.Port)
	log.Printf("API server starting on %s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleRoot handles root endpoint
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"name":    "Aetheria Blockchain",
		"version": "1.0.0",
		"height":  s.Blockchain.Height(),
	}
	s.jsonResponse(w, response)
}

// handleHealth handles health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "healthy",
	}
	s.jsonResponse(w, response)
}

// handleBlocks handles blocks endpoint
func (s *Server) handleBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getBlocks(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getBlocks returns all blocks
func (s *Server) getBlocks(w http.ResponseWriter, r *http.Request) {
	blocks := s.Blockchain.Blocks
	s.jsonResponse(w, blocks)
}

// handleBlock handles single block endpoint
func (s *Server) handleBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract block index from URL
	indexStr := r.URL.Path[len("/block/"):]
	index, err := strconv.ParseUint(indexStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid block index", http.StatusBadRequest)
		return
	}

	block := s.Blockchain.GetBlock(index)
	if block == nil {
		http.Error(w, "Block not found", http.StatusNotFound)
		return
	}

	s.jsonResponse(w, block)
}

// handleTransactions handles transactions endpoint
func (s *Server) handleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.createTransaction(w, r)
	} else if r.Method == http.MethodGet {
		s.getPendingTransactions(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// TransactionRequest represents a transaction creation request
type TransactionRequest struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Amount     uint64 `json:"amount"`
	Fee        uint64 `json:"fee"`
	PrivateKey string `json:"private_key"`
}

// createTransaction creates a new transaction
func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create transaction
	tx := blockchain.NewTransaction(req.From, req.To, req.Amount, req.Fee)

	// Sign transaction
	privateKey, err := crypto.PrivateKeyFromHex(req.PrivateKey)
	if err != nil {
		http.Error(w, "Invalid private key", http.StatusBadRequest)
		return
	}

	if err := tx.Sign(privateKey); err != nil {
		http.Error(w, "Failed to sign transaction", http.StatusInternalServerError)
		return
	}

	// Add to blockchain
	if err := s.Blockchain.AddTransaction(tx); err != nil {
		http.Error(w, fmt.Sprintf("Failed to add transaction: %v", err), http.StatusBadRequest)
		return
	}

	// Broadcast to network
	s.Node.BroadcastTransaction(tx)

	s.jsonResponse(w, tx)
}

// getPendingTransactions returns pending transactions
func (s *Server) getPendingTransactions(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, s.Blockchain.PendingTxs)
}

// handleTransaction handles single transaction endpoint
func (s *Server) handleTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	txID := r.URL.Path[len("/transaction/"):]
	tx := s.Blockchain.GetTransaction(txID)
	if tx == nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	s.jsonResponse(w, tx)
}

// handleBalance handles balance endpoint
func (s *Server) handleBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	address := r.URL.Path[len("/balance/"):]
	balance := s.Blockchain.State.GetBalance(address)
	stake := s.Blockchain.State.GetStake(address)

	response := map[string]interface{}{
		"address": address,
		"balance": balance,
		"stake":   stake,
	}
	s.jsonResponse(w, response)
}

// StakeRequest represents a stake request
type StakeRequest struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
}

// handleStake handles staking endpoint
func (s *Server) handleStake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Add stake
	if err := s.Blockchain.State.AddStake(req.Address, req.Amount); err != nil {
		http.Error(w, fmt.Sprintf("Failed to add stake: %v", err), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"address": req.Address,
		"stake":   s.Blockchain.State.GetStake(req.Address),
	}
	s.jsonResponse(w, response)
}

// handleValidators handles validators endpoint
func (s *Server) handleValidators(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	validators := s.Consensus.ValidatorSet.GetValidatorInfos()
	s.jsonResponse(w, validators)
}

// handleNewWallet handles wallet creation endpoint
func (s *Server) handleNewWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		http.Error(w, "Failed to generate wallet", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"address":     keyPair.Address(),
		"public_key":  crypto.PublicKeyToHex(keyPair.PublicKey),
		"private_key": crypto.PrivateKeyToHex(keyPair.PrivateKey),
	}
	s.jsonResponse(w, response)
}

// jsonResponse sends a JSON response
func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
