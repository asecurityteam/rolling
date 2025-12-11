# syntax=docker/dockerfile:1

# Build a local golangci-lint image
FROM golangci/golangci-lint:v2.6 AS lint
USER root
# Intentionally empty: this stage serves as a runnable golangci-lint container


