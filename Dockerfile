# syntax=docker/dockerfile:1

# Build a local Go toolchain image
FROM golang:1.24 AS go
USER root
# Intentionally empty: this stage serves as a runnable Go toolchain container


