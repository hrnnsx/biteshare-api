# Simple Makefile for a Go project

# Build the application
all: build test

build:
	@echo "Building Ecommerce API..."
	@go build -o bin/ecommerce.exe cmd/ecommerce/main.go
	@echo "Building Donation API..."
	@go build -o bin/donation.exe cmd/donation/main.go

run-ecommerce:
	@go run cmd/ecommerce/main.go

run-donation:
	@go run cmd/donation/main.go

seed-admin:
	@go run cmd/seed-admin/main.go

test:
	@echo "Testing..."
	@go test ./... -v

clean:
	@echo "Cleaning..."
	@rm -rf bin/

.PHONY: all build run-ecommerce run-donation seed-admin test clean
