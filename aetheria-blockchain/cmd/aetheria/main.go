package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aetheria/blockchain/pkg/api"
	"github.com/aetheria/blockchain/pkg/blockchain"
	"github.com/aetheria/blockchain/pkg/consensus"
	"github.com/aetheria/blockchain/pkg/crypto"
	"github.com/aetheria/blockchain/pkg/network"
	"github.com/aetheria/blockchain/pkg/wallet"
)

const (
	// Initial supply of Aetheria tokens
	InitialSupply = 1000000
	// Minimum stake to become a validator
	MinStake = 1000
	// Block time (time between blocks)
	BlockTime = 5 * time.Second
)

func main() {
	// Command line flags
	var (
		port        = flag.Int("port", 8080, "API server port")
		nodeID      = flag.String("node-id", "node1", "Node ID")
		isValidator = flag.Bool("validator", false, "Run as validator")
		walletFile  = flag.String("wallet", "", "Wallet file path")
		newWallet   = flag.Bool("new-wallet", false, "Create new wallet")
		genesisAddr = flag.String("genesis", "", "Genesis address (for first node)")
	)
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting Aetheria Blockchain Node")

	// Handle new wallet creation
	if *newWallet {
		createNewWallet()
		return
	}

	// Determine genesis address
	var genesisAddress string
	if *genesisAddr != "" {
		genesisAddress = *genesisAddr
	} else {
		// Generate a default genesis address
		keyPair, err := crypto.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Failed to generate genesis key pair: %v", err)
		}
		genesisAddress = keyPair.Address()
		log.Printf("Generated genesis address: %s", genesisAddress)
	}

	// Create blockchain
	bc := blockchain.NewBlockchain(genesisAddress, InitialSupply)
	log.Printf("Blockchain initialized with genesis address: %s", genesisAddress)
	log.Printf("Initial supply: %d Aetheria tokens", InitialSupply)

	// Create consensus engine
	pos := consensus.NewPoS(MinStake, BlockTime)
	log.Printf("PoS consensus initialized (MinStake: %d, BlockTime: %v)", MinStake, BlockTime)

	// Create node
	nodeAddress := fmt.Sprintf("localhost:%d", *port)
	node := network.NewNode(*nodeID, nodeAddress, bc, pos)

	// Setup validator if requested
	if *isValidator {
		if *walletFile == "" {
			log.Fatal("Validator mode requires --wallet flag")
		}

		w, err := wallet.LoadFromFile(*walletFile)
		if err != nil {
			log.Fatalf("Failed to load wallet: %v", err)
		}

		keyPair, err := w.GetKeyPair()
		if err != nil {
			log.Fatalf("Failed to get key pair: %v", err)
		}

		// Add initial stake for validator
		if err := bc.State.AddStake(w.Address, MinStake); err != nil {
			// If stake fails, give the validator some initial balance
			bc.State.SetBalance(w.Address, MinStake*2)
			if err := bc.State.AddStake(w.Address, MinStake); err != nil {
				log.Fatalf("Failed to add stake: %v", err)
			}
		}

		validator := consensus.ValidatorFromKeyPair(keyPair, MinStake)
		if err := node.SetValidator(validator); err != nil {
			log.Fatalf("Failed to set validator: %v", err)
		}

		log.Printf("Node running as validator: %s", w.Address)
		log.Printf("Validator stake: %d Aetheria", MinStake)
	}

	// Start node
	if err := node.Start(); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	// Create and start API server
	apiServer := api.NewServer(*port, node, bc, pos)
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Fatalf("Failed to start API server: %v", err)
		}
	}()

	log.Printf("Node %s started successfully", *nodeID)
	log.Printf("API server listening on http://localhost:%d", *port)
	log.Printf("Blockchain height: %d", bc.Height())

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	node.Stop()
}

// createNewWallet creates a new wallet and saves it to a file
func createNewWallet() {
	w, err := wallet.NewWallet()
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	filename := fmt.Sprintf("wallet_%s.json", w.Address[:8])
	if err := w.SaveToFile(filename); err != nil {
		log.Fatalf("Failed to save wallet: %v", err)
	}

	log.Printf("New wallet created successfully!")
	log.Printf("Address: %s", w.Address)
	log.Printf("Public Key: %s", w.PublicKey)
	log.Printf("Wallet saved to: %s", filename)
	log.Printf("\nIMPORTANT: Keep your wallet file safe! It contains your private key.")
}
