package network

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aetheria/blockchain/pkg/blockchain"
	"github.com/aetheria/blockchain/pkg/consensus"
)

// MessageType represents the type of network message
type MessageType string

const (
	MsgTypeBlock       MessageType = "block"
	MsgTypeTransaction MessageType = "transaction"
	MsgTypePing        MessageType = "ping"
	MsgTypePong        MessageType = "pong"
	MsgTypeGetBlocks   MessageType = "get_blocks"
	MsgTypeBlocks      MessageType = "blocks"
)

// Message represents a network message
type Message struct {
	Type      MessageType     `json:"type"`
	Data      json.RawMessage `json:"data"`
	From      string          `json:"from"`
	Timestamp int64           `json:"timestamp"`
}

// Node represents a blockchain node
type Node struct {
	ID           string
	Address      string
	Blockchain   *blockchain.Blockchain
	Consensus    *consensus.PoS
	Peers        map[string]*Peer
	IsValidator  bool
	Validator    *consensus.Validator
	mu           sync.RWMutex
	stopChan     chan struct{}
	messageChan  chan *Message
}

// NewNode creates a new node
func NewNode(id, address string, bc *blockchain.Blockchain, pos *consensus.PoS) *Node {
	return &Node{
		ID:          id,
		Address:     address,
		Blockchain:  bc,
		Consensus:   pos,
		Peers:       make(map[string]*Peer),
		IsValidator: false,
		stopChan:    make(chan struct{}),
		messageChan: make(chan *Message, 100),
	}
}

// SetValidator sets this node as a validator
func (n *Node) SetValidator(validator *consensus.Validator) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if err := n.Consensus.RegisterValidator(validator); err != nil {
		return fmt.Errorf("failed to register validator: %w", err)
	}

	n.IsValidator = true
	n.Validator = validator
	return nil
}

// Start starts the node
func (n *Node) Start() error {
	log.Printf("Starting node %s at %s", n.ID, n.Address)

	// Start message processing
	go n.processMessages()

	// Start block production if validator
	if n.IsValidator {
		go n.produceBlocks()
	}

	return nil
}

// Stop stops the node
func (n *Node) Stop() {
	close(n.stopChan)
	log.Printf("Node %s stopped", n.ID)
}

// processMessages processes incoming messages
func (n *Node) processMessages() {
	for {
		select {
		case <-n.stopChan:
			return
		case msg := <-n.messageChan:
			n.handleMessage(msg)
		}
	}
}

// handleMessage handles a network message
func (n *Node) handleMessage(msg *Message) {
	switch msg.Type {
	case MsgTypeBlock:
		var block blockchain.Block
		if err := json.Unmarshal(msg.Data, &block); err != nil {
			log.Printf("Failed to unmarshal block: %v", err)
			return
		}
		n.handleBlock(&block)

	case MsgTypeTransaction:
		var tx blockchain.Transaction
		if err := json.Unmarshal(msg.Data, &tx); err != nil {
			log.Printf("Failed to unmarshal transaction: %v", err)
			return
		}
		n.handleTransaction(&tx)

	case MsgTypePing:
		n.handlePing(msg.From)

	case MsgTypeGetBlocks:
		n.handleGetBlocks(msg.From)
	}
}

// handleBlock handles a received block
func (n *Node) handleBlock(block *blockchain.Block) {
	log.Printf("Node %s received block %d from validator %s", n.ID, block.Index, block.Validator)

	// Validate block
	prevBlock := n.Blockchain.GetLatestBlock()
	if err := n.Consensus.ValidateBlock(block, prevBlock); err != nil {
		log.Printf("Invalid block: %v", err)
		return
	}

	// Add block to blockchain
	if err := n.Blockchain.AddBlock(block); err != nil {
		log.Printf("Failed to add block: %v", err)
		return
	}

	log.Printf("Block %d added to chain", block.Index)

	// Broadcast to peers
	n.BroadcastBlock(block)
}

// handleTransaction handles a received transaction
func (n *Node) handleTransaction(tx *blockchain.Transaction) {
	log.Printf("Node %s received transaction %s", n.ID, tx.ID)

	// Add to blockchain
	if err := n.Blockchain.AddTransaction(tx); err != nil {
		log.Printf("Failed to add transaction: %v", err)
		return
	}

	// Broadcast to peers
	n.BroadcastTransaction(tx)
}

// handlePing handles a ping message
func (n *Node) handlePing(from string) {
	// Send pong response
	msg := &Message{
		Type:      MsgTypePong,
		From:      n.ID,
		Timestamp: time.Now().Unix(),
	}
	n.sendMessage(from, msg)
}

// handleGetBlocks handles a request for blocks
func (n *Node) handleGetBlocks(from string) {
	// Send all blocks
	blocks := n.Blockchain.Blocks
	data, _ := json.Marshal(blocks)
	msg := &Message{
		Type:      MsgTypeBlocks,
		Data:      data,
		From:      n.ID,
		Timestamp: time.Now().Unix(),
	}
	n.sendMessage(from, msg)
}

// produceBlocks produces blocks if this node is a validator
func (n *Node) produceBlocks() {
	ticker := time.NewTicker(n.Consensus.BlockTime)
	defer ticker.Stop()

	for {
		select {
		case <-n.stopChan:
			return
		case <-ticker.C:
			n.tryProduceBlock()
		}
	}
}

// tryProduceBlock attempts to produce a new block
func (n *Node) tryProduceBlock() {
	if !n.IsValidator {
		return
	}

	latestBlock := n.Blockchain.GetLatestBlock()
	
	// Check if it's time to create a block
	if !n.Consensus.ShouldCreateBlock(latestBlock.Timestamp) {
		return
	}

	// Select validator for this slot
	selectedValidator, err := n.Consensus.SelectValidator(latestBlock.Hash, time.Now().Unix())
	if err != nil {
		log.Printf("Failed to select validator: %v", err)
		return
	}

	// Check if this node is the selected validator
	if selectedValidator.Address != n.Validator.Address {
		return
	}

	log.Printf("Node %s selected to produce block", n.ID)

	// Create block
	block := n.Blockchain.CreateBlock(n.Validator.Address)

	// Sign block
	if err := block.Sign(n.Validator.PrivateKey); err != nil {
		log.Printf("Failed to sign block: %v", err)
		return
	}

	// Add block to blockchain
	if err := n.Blockchain.AddBlock(block); err != nil {
		log.Printf("Failed to add block: %v", err)
		return
	}

	log.Printf("Block %d produced by validator %s", block.Index, n.Validator.Address)

	// Broadcast block
	n.BroadcastBlock(block)
}

// BroadcastBlock broadcasts a block to all peers
func (n *Node) BroadcastBlock(block *blockchain.Block) {
	data, _ := json.Marshal(block)
	msg := &Message{
		Type:      MsgTypeBlock,
		Data:      data,
		From:      n.ID,
		Timestamp: time.Now().Unix(),
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, peer := range n.Peers {
		n.sendMessage(peer.ID, msg)
	}
}

// BroadcastTransaction broadcasts a transaction to all peers
func (n *Node) BroadcastTransaction(tx *blockchain.Transaction) {
	data, _ := json.Marshal(tx)
	msg := &Message{
		Type:      MsgTypeTransaction,
		Data:      data,
		From:      n.ID,
		Timestamp: time.Now().Unix(),
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, peer := range n.Peers {
		n.sendMessage(peer.ID, msg)
	}
}

// sendMessage sends a message to a peer
func (n *Node) sendMessage(peerID string, msg *Message) {
	n.mu.RLock()
	peer, exists := n.Peers[peerID]
	n.mu.RUnlock()

	if !exists {
		return
	}

	peer.SendMessage(msg)
}

// AddPeer adds a peer to the node
func (n *Node) AddPeer(peer *Peer) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Peers[peer.ID] = peer
	log.Printf("Node %s added peer %s", n.ID, peer.ID)
}

// RemovePeer removes a peer from the node
func (n *Node) RemovePeer(peerID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.Peers, peerID)
	log.Printf("Node %s removed peer %s", n.ID, peerID)
}

// ReceiveMessage receives a message from the network
func (n *Node) ReceiveMessage(msg *Message) {
	select {
	case n.messageChan <- msg:
	default:
		log.Printf("Message channel full, dropping message")
	}
}
