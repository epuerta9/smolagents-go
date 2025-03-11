#!/bin/bash

set -e

# Get the directory of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$DIR"

echo "Running tests for smolagents-go..."

# Run go tests for each package
echo "Testing tools package..."
go test -v ./pkg/tools

echo "Testing models package..."
go test -v ./pkg/models

echo "Testing memory package..."
go test -v ./pkg/memory 

echo "Testing agents package..."
go test -v ./pkg/agents

echo "All tests completed!" 