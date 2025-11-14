package network

import (
	"log"
	"sync"
)

// Peer represents a network peer
type Peer struct {
	ID          string
	Address     string
	Connected   bool
	messageChan chan *Message
	mu          sync.RWMutex
}

// NewPeer creates a new peer
func NewPeer(id, address string) *Peer {
	return &Peer{
		ID:          id,
		Address:     address,
		Connected:   false,
		messageChan: make(chan *Message, 100),
	}
}

// Connect connects to the peer
func (p *Peer) Connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Connected = true
	log.Printf("Connected to peer %s at %s", p.ID, p.Address)
	return nil
}

// Disconnect disconnects from the peer
func (p *Peer) Disconnect() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Connected = false
	log.Printf("Disconnected from peer %s", p.ID)
}

// SendMessage sends a message to the peer
func (p *Peer) SendMessage(msg *Message) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.Connected {
		return
	}

	select {
	case p.messageChan <- msg:
	default:
		log.Printf("Peer %s message channel full", p.ID)
	}
}

// IsConnected checks if peer is connected
func (p *Peer) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Connected
}
