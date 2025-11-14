#!/bin/bash

# Start Aetheria Node Script

set -e

# Default values
PORT=8080
NODE_ID="node1"
VALIDATOR=false
WALLET=""

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --port)
      PORT="$2"
      shift 2
      ;;
    --node-id)
      NODE_ID="$2"
      shift 2
      ;;
    --validator)
      VALIDATOR=true
      shift
      ;;
    --wallet)
      WALLET="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Build the application
echo "Building Aetheria blockchain..."
go build -o aetheria ./cmd/aetheria

# Start the node
echo "Starting node $NODE_ID on port $PORT..."

if [ "$VALIDATOR" = true ]; then
  if [ -z "$WALLET" ]; then
    echo "Error: Validator mode requires --wallet flag"
    exit 1
  fi
  ./aetheria --port=$PORT --node-id=$NODE_ID --validator --wallet=$WALLET
else
  ./aetheria --port=$PORT --node-id=$NODE_ID
fi
