#!/usr/bin/env bash
set -euo pipefail

if ! command -v buf &> /dev/null; then
  echo "buf is required. Install: https://buf.build/docs/installation"
  exit 1
fi

cd proto
buf generate
echo "proto generation complete"
