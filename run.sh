#!/bin/bash
source .env
export DATABASE_URL
go run ./cmd/server