#!/bin/sh

# Exit immediately if a command exits with a non-zero status
set -e

echo "Running database migrations..."
./pollapp --migrate-up

echo "Starting application..."
exec ./pollapp --workers 