# BangunGrid Enterprise Makefile
SHELL := /bin/bash

.PHONY: help docker-up docker-down docker-logs docker-clean tools lint sec vuln test-race verify-all

help:
	@echo "Perintah Manajemen BangunGrid:"
	@echo "  make docker-up       - Menyalakan seluruh kontainer (backing services + auth-service)"
	@echo "  make docker-infra    - Menyalakan hanya backing services (Postgres, Redis, NATS, MinIO)"
	@echo "  make docker-build    - Build seluruh Docker image microservice"
	@echo "  make docker-down     - Mematikan seluruh kontainer"
	@echo "  make docker-logs     - Melihat log kontainer"
	@echo "  make docker-clean    - Reset data kontainer dan volume"
	@echo "  make tools           - Instal perkakas linting & security"
	@echo "  make lint            - Jalankan golangci-lint"
	@echo "  make sec             - Jalankan gosec security scanner"
	@echo "  make vuln            - Jalankan govulncheck dependency scanner"
	@echo "  make test-race       - Jalankan unit tests dengan Race Detector"
	@echo "  make verify-all      - Jalankan seluruh gate pemeriksaan kualitas & keamanan"

docker-up:
	@echo "==> Menyalakan seluruh ekosistem Docker (Backing Services + Microservices)..."
	docker compose up -d

docker-infra:
	@echo "==> Menyalakan hanya Backing Services (Postgres, Redis, NATS, MinIO)..."
	docker compose up -d postgres redis nats minio

docker-build:
	@echo "==> Membangun seluruh Docker image microservices..."
	docker compose build

docker-down:
	@echo "==> Mematikan infrastruktur Docker..."
	docker compose down

docker-logs:
	@echo "==> Menampilkan log kontainer..."
	docker compose logs -f

docker-clean:
	@echo "==> Membersihkan seluruh kontainer dan volume..."
	docker compose down -v

tools:
	@echo "==> Menginstal perkakas linting & security..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest

lint:
	@echo "==> [1/4] Menjalankan golangci-lint..."
	@golangci-lint run ./...

sec:
	@echo "==> [2/4] Menjalankan gosec security scanner..."
	@gosec -quiet ./...

vuln:
	@echo "==> [3/4] Menjalankan govulncheck..."
	@govulncheck ./...

test-race:
	@echo "==> [4/4] Menjalankan Unit Tests dengan Data Race Detector..."
	@go test -race -v -coverprofile=coverage.out ./...

verify-all: lint sec vuln test-race
	@echo ""
	@echo "=========================================================="
	@echo "  SELURUH PENGECEKAN KUALITAS & KEAMANAN BERHASIL (LULUS)!"
	@echo "=========================================================="
