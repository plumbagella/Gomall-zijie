#!/bin/bash

# Kill go run processes
pkill -f "go run"

# Kill go build processes
pkill -f "go-build"

echo "All go-related processes stopped."